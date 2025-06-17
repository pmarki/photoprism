package query

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// coversBusy is true when the covers are currently updating.
var coversBusy = atomic.Bool{}

// UpdateAlbumDefaultCovers updates default album cover thumbs.
func UpdateAlbumDefaultCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	var res *gorm.DB

	var albums []entity.Album

	res = Db().Table(entity.Album{}.TableName()).Where("album_type = ?", "album").Scan(&albums)

	if res.Error != nil {
		log.Errorf("UpdateAlbumDefaultCovers, err when loading albums %s", res.Error)
		return
	}

	for _, album := range albums {
		UpdateCover(album.AlbumUID)
	}

	return nil
}

// UpdateAlbumFolderCovers updates folder album cover thumbs.
func UpdateAlbumFolderCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	var res *gorm.DB

	var albums []entity.Album

	res = Db().Table(entity.Album{}.TableName()).Where("album_type = ?", "folder").Scan(&albums)

	if res.Error != nil {
		log.Errorf("UpdateAlbumDefaultCovers, err when loading albums %s", res.Error)
		return
	}

	for _, album := range albums {
		UpdateCover(album.AlbumUID)
	}

	return nil
}

// UpdateAlbumMonthCovers updates month album cover thumbs.
func UpdateAlbumMonthCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	var res *gorm.DB

	var albums []entity.Album

	res = Db().Table(entity.Album{}.TableName()).Where("album_type = ?", "month").Scan(&albums)

	if res.Error != nil {
		log.Errorf("UpdateAlbumDefaultCovers, err when loading albums %s", res.Error)
		return
	}

	for _, album := range albums {
		UpdateCover(album.AlbumUID)
	}

	return nil
}

// UpdateAlbumCovers updates album cover thumbs.
func UpdateAlbumCovers() (err error) {
	// Update Default Albums.
	if err = UpdateAlbumDefaultCovers(); err != nil {
		return err
	}

	// Update Folder Albums.
	if err = UpdateAlbumFolderCovers(); err != nil {
		return err
	}

	// Update Monthly Albums.
	if err = UpdateAlbumMonthCovers(); err != nil {
		return err
	}

	return nil
}

// UpdateLabelCovers updates label cover thumbs.
func UpdateLabelCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("thumb_src = ?", entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE labels LEFT JOIN (
		SELECT p2.label_id, f.file_hash FROM files f, (
			SELECT pl.label_id as label_id, max(p.id) AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY pl.label_id
			UNION
			SELECT c.category_id as label_id, max(p.id) AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
				JOIN categories c ON c.label_id = pl.label_id
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY c.category_id
			) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?) AND f.file_missing = 0
		) b ON b.label_id = labels.id
		SET thumb = b.file_hash WHERE ?`, media.PreviewExpr, condition)
	case SQLite3:
		res = Db().Table(entity.Label{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
		SELECT f.file_hash FROM files f 
			JOIN photos_labels pl ON pl.label_id = labels.id AND pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > 0
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
		) WHERE ?`, media.PreviewExpr, condition))

		if res.Error == nil {
			catRes := Db().Table(entity.Label{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f 
			JOIN photos_labels pl ON pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN categories c ON c.label_id = pl.label_id AND c.category_id = labels.id
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > 0
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
			) WHERE thumb IS NULL`, media.PreviewExpr))

			res.RowsAffected += catRes.RowsAffected
		}
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "label", "labels"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed to update labels, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateSubjectCovers updates subject cover thumbs.
func UpdateSubjectCovers(public bool) (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB
	var photosJoin *gorm.SqlExpr

	// Use faces tagged on private pictures as cover images?
	// see https://github.com/photoprism/photoprism/issues/4238
	// and https://github.com/photoprism/photoprism/issues/2570#issuecomment-1231690056
	if public {
		photosJoin = gorm.Expr("p.id = f.photo_id AND p.deleted_at IS NULL AND p.photo_private = 0")
	} else {
		photosJoin = gorm.Expr("p.id = f.photo_id AND p.deleted_at IS NULL")
	}

	condition := gorm.Expr("subjects.subj_type = ? AND thumb_src = ?", entity.SubjPerson, entity.SrcAuto)

	// Compose SQL update query.
	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE subjects LEFT JOIN (
    	SELECT m.subj_uid, m.q, MAX(m.thumb) AS marker_thumb
    		FROM markers m
    	    JOIN files f ON f.file_uid = m.file_uid AND f.deleted_at IS NULL
			JOIN photos p ON ?
			WHERE m.subj_uid <> '' AND m.subj_uid IS NOT NULL
			  AND m.marker_invalid = 0 AND m.thumb IS NOT NULL AND m.thumb <> ''
			GROUP BY m.subj_uid, m.q
			) b ON b.subj_uid = subjects.subj_uid
		SET thumb = marker_thumb WHERE ?`,
			photosJoin,
			condition,
		)
	case SQLite3:
		// from := gorm.Expr(fmt.Sprintf("%s m WHERE m.subj_uid = %s.subj_uid ", markerTable, subjTable))
		res = Db().Table(entity.Subject{}.TableName()).UpdateColumn("thumb",
			gorm.Expr(`(
                SELECT m.thumb
					FROM markers m 
					JOIN files f ON f.file_uid = m.file_uid AND f.deleted_at IS NULL
					JOIN photos p ON ?
					WHERE m.subj_uid = subjects.subj_uid AND m.thumb <> ''
					ORDER BY m.subj_src DESC, m.q DESC LIMIT 1
				) WHERE ?`,
				photosJoin,
				condition,
			),
		)
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "subject", "subjects"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed to update subjects, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateCoversAsync runs UpdateCovers in a go routine
// and logs the returned error, if any, as a warning.
func UpdateCoversAsync() {
	go func() {
		if err := UpdateCovers(); err != nil {
			log.Warnf("index: %s (update covers)", clean.Error(err))
		}
	}()
}

// UpdateCover updates cover for given album
func UpdateCover(albumUID string) {
	frm := form.SearchPhotos{
		Scope:   albumUID,
		Type:    "image",
		Primary: true,
	}

	photos, count, err := search.Photos(frm)

	if err != nil {
		log.Errorf("Can't generate cover for album %s %s", albumUID, err)
		return
	}

	var thumbs []string
	if count == 0 {
		return
	}
	if count > 0 {
		thumbs = append(thumbs, photos[0].FileHash)
	}
	if count > 1 {
		if count == 3 {
			thumbs = append(thumbs, photos[1].FileHash)
		}
		if count > 3 {
			i1 := len(photos) * 1 / 3
			i2 := len(photos) * 2 / 3

			thumbs = append(thumbs, photos[i1].FileHash)
			thumbs = append(thumbs, photos[i2].FileHash)
		}

		thumbs = append(thumbs, photos[len(photos)-1].FileHash)
	}

	thumb := strings.Join(thumbs, " ")
	log.Infof("thumbs for %s: %s", albumUID, thumb)

	var res *gorm.DB
	res = Db().Table(entity.Album{}.TableName()).Where("album_UID = ?", albumUID).Update("thumb", thumb)
	if res.Error != nil {
		log.Errorf("Cant save thumb %s %s", albumUID, err)
		return
	}
}

// UpdateCovers updates album, subject, and label cover thumbs.
func UpdateCovers() (err error) {
	if !coversBusy.CompareAndSwap(false, true) {
		log.Debugf("index: skipped updating covers because it is already in progress")
		return nil
	}

	defer coversBusy.Store(false)

	log.Debugf("index: updating covers")

	// Update Albums.
	if err = UpdateAlbumCovers(); err != nil {
		return fmt.Errorf("%s while updating album covers", err)
	}

	// Update Labels.
	if err = UpdateLabelCovers(); err != nil {
		return fmt.Errorf("%s while updating label covers", err)
	}

	// Update Subjects.
	if err = UpdateSubjectCovers(true); err != nil {
		return fmt.Errorf("%s while updating subject covers", err)
	}

	return nil
}
