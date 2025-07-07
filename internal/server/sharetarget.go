package server

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dustin/go-humanize/english"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// registerShareTargetRoute adds routes for PWA share target
func registerShareTargetRoute(router *gin.Engine, conf *config.Config) {
	s := router.Group(conf.BaseUri("/pwa"))
	{
		UploadSharedFiles(s)
	}
}

// UploadSharedFiles upload files shared in PWA
//
//	@Tags	PWA
//	@Router	/pwa/sharetarget [post]
func UploadSharedFiles(router *gin.RouterGroup) {
	router.POST("/sharetarget", func(c *gin.Context) {
		conf := get.Config()
		settings := conf.Settings()
		if conf.Demo() {
			api.AbortForbidden(c)
			return
		}
		event.Log.Info("sharetarget: uploading files")
		f, err := c.MultipartForm()

		if err != nil {
			event.Log.Error(err.Error())
			api.Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		files := f.File["images"]
		fmt.Printf("form: %v\n", f)

		if len(files) == 0 {
			event.Log.Error("sharetarget: no images to upload")
			api.Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		dir := time.Now().Format("2006.01")
		uploadDir := filepath.Join(conf.OriginalsPath(), dir)
		// Check if the folder already exists
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			// Create the folder and any necessary parents
			err := os.MkdirAll(uploadDir, os.ModePerm)
			if err != nil {
				log.Errorf("failed to create folder: %w", err)
				api.AbortUnexpectedError(c)
				return
			}

			folder := entity.NewFolder(entity.RootOriginals, dir, entity.Now())

			if err := folder.Create(); err != nil {
				log.Errorf("folder create: %s)", err)
				api.AbortUnexpectedError(c)
				return
			}
		}

		for _, file := range files {
			// Get absolute file path.
			filePath := path.Join(uploadDir, file.Filename)
			log.Infof("Processing file %s", filePath)
			// Save image.
			if err = c.SaveUploadedFile(file, filePath); err != nil {
				log.Errorf("sharetarget, can't save file  %s: %s)", filePath, err)
				api.Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
				return
			} else {
				log.Infof("sharetarget, saved %s", filePath)

			}
		}

		frm := form.IndexOptions{
			Path:    dir,
			Cleanup: false,
			Rescan:  false,
		}

		// Configure index options.
		convert := settings.Index.Convert && conf.SidecarWritable()
		skipArchived := settings.Index.SkipArchived

		indOpt := photoprism.NewIndexOptions(filepath.Clean(frm.Path), frm.Rescan, convert, true, false, skipArchived)

		ind := get.Index()
		indexStart := time.Now()

		// Update file index.
		_, indexed := ind.Start(indOpt)

		log.Infof("index: updated %s [%s]", english.Plural(indexed, "file", "files"), time.Since(indexStart))

		// Find album ID to redirect to
		album := entity.FindFolderAlbum(dir)
		if album == nil {
			log.Errorf("No album related to path: %s)", dir)
			api.AbortUnexpectedError(c)
		}

		// update cover
		go func() {
			query.UpdateCover(album.AlbumUID)
		}()

		// TODO update cover for calendar album
		// calendar := entity.FindMonthAlbum()
		// if calendar != nil {
		// 	log.Errorf("No month album related to path: %s)", dir)

		// 	// update cover
		// 	go func() {
		// 		query.UpdateCover(calendar.AlbumUID)
		// 	}()
		// }

		c.Redirect(http.StatusFound, "/library/folders/"+album.AlbumUID+"/view")
	})
}
