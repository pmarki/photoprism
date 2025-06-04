package query

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

// PurgeOrphans removes orphan database entries.
func PurgeOrphans(originalsPath string) error {
	// Remove files without a photo.
	start := time.Now()
	if count, err := PurgeOrphanFiles(originalsPath); err != nil {
		return err
	} else if count > 0 {
		log.Infof("index: removed %d orphan files [%s]", count, time.Since(start))
	} else {
		log.Debugf("index: found no orphan files [%s]", time.Since(start))
	}

	// Remove duplicates without an original file.
	if err := PurgeOrphanDuplicates(); err != nil {
		return err
	}

	// Remove unused countries.
	if err := PurgeOrphanCountries(); err != nil {
		return err
	}

	// Remove unused cameras.
	if err := PurgeOrphanCameras(); err != nil {
		return err
	}

	// Remove unused camera lenses.
	if err := PurgeOrphanLenses(); err != nil {
		return err
	}

	return nil
}

// PurgeOrphanFiles removes files without a photo from the index.
func PurgeOrphanFiles(originalsPath string) (count int, err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	files, err := OriginalFiles()

	if err != nil {
		return 0, err
	}

	for i := range files {
		fullPath := filepath.Join(originalsPath, files[i].FileName)
		if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {

			log.Debugf("PurgeOrphanFiles: deleting %s", fullPath)

			if err = files[i].DeletePermanently(); err != nil {
				return count, err
			}

			count++
		}
	}

	return count, err
}

// PurgeOrphanDuplicates deletes all files from the duplicates table that don't exist in the files table.
func PurgeOrphanDuplicates() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	result := UnscopedDb().
		Delete(entity.Duplicate{},
			"file_hash NOT IN (SELECT file_hash FROM files WHERE file_missing = 0 AND deleted_at IS NULL)")

	return result.Error
}

// PurgeOrphanCountries removes countries without any photos.
func PurgeOrphanCountries() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushCountryCache()

	result := UnscopedDb().
		Exec(`DELETE FROM countries WHERE country_slug <> ? AND id NOT IN (SELECT photo_country FROM photos)`,
			entity.UnknownCountry.CountrySlug)

	return result.Error
}

// PurgeOrphanCameras removes cameras without any photos.
func PurgeOrphanCameras() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushCameraCache()

	result := UnscopedDb().
		Exec(`DELETE FROM cameras WHERE camera_slug <> ? AND id NOT IN (SELECT camera_id FROM photos)`,
			entity.UnknownCamera.CameraSlug)

	return result.Error
}

// PurgeOrphanLenses removes cameras without any photos.
func PurgeOrphanLenses() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushLensCache()

	result := UnscopedDb().
		Exec(`DELETE FROM lenses WHERE lens_slug <> ? AND id NOT IN (SELECT lens_id FROM photos)`,
			entity.UnknownLens.LensSlug)

	return result.Error
}

func PurgeOrphanFolders(originalsRoot string) (err error) {

	var rows []*entity.Folder

	dirs, err := fs.Dirs(originalsRoot, true, true)

	if err != nil {
		return err
	}

	folders := make(entity.Folders, len(dirs))

	db := UnscopedDb().Table("folders").Select("*").Where("folders.root = ?", entity.RootOriginals).Find(&rows)

	res := db.Scan(&folders)

	if res.Error != nil {
		return err
	}

	for _, path := range rows {
		fullPath := filepath.Join(originalsRoot, path.Path)
		if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
			log.Debugf("PurgeOrphanFolders: deleting %s", fullPath)
			err := path.Delete()
			if err != nil {
				log.Debugf("PurgeOrphanFolders: cant delete %s", fullPath)
			}
		}
	}

	return nil
}

func PurgeOrphanAlbumFolders(originalsRoot string) (err error) {

	var rows []*entity.Album

	db := UnscopedDb().Table("albums").Select("*").Where("albums.album_type = ?", entity.AlbumFolder).Find(&rows)

	if db.Error != nil {
		return err
	}

	for _, path := range rows {
		fullPath := filepath.Join(originalsRoot, path.AlbumPath)
		if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
			log.Debugf("PurgeOrphanAlbumFolders: deleting %s", fullPath)
			err := path.DeletePermanently()
			if err != nil {
				log.Debugf("PurgeOrphanAlbumFolders: cant delete %s", fullPath)
			}
		}
	}

	return nil
}
