package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// registerShareTargetRoute adds routes for PWA share target
func registerPhotoShowRoute(router *gin.Engine, conf *config.Config) {
	s := router.Group(conf.BaseUri("/photoshow/V1"))
	{
		GetList(s)
		GetThumb(s)
	}
}

// GetList returns a list of images as json
//
//	@Router	/photoshow/v1/getList/:albumUID [get]
func GetList(router *gin.RouterGroup) {
	router.GET("/getList/:UID", func(c *gin.Context) {
		conf := get.Config()
		//settings := conf.Settings()
		if conf.Demo() {
			api.AbortForbidden(c)
			return
		}
		uid := clean.UID(c.Param("UID"))

		if uid == "" {
			api.AbortForbidden(c)
			return
		}

		frm := form.SearchPhotos{
			Album:    uid,
			Count:    100000,
			Offset:   0,
			Hidden:   false,
			Archived: false,
			Quality:  1,
		}

		files, _, err := search.PhotoHash(frm)

		if err != nil {
			log.Errorf("PhotoShow GetList: %s", err)
			api.AbortBadRequest(c)
		}

		// Return as JSON.
		c.JSON(http.StatusOK, files)
	})
}

// GetThumb returns a thumb for given file
//
//	@Router	/photoshow/v1/getThumb/:hash/:size [get]
func GetThumb(router *gin.RouterGroup) {
	router.GET("/getThumb/:hash/:size", func(c *gin.Context) {
		conf := get.Config()
		//settings := conf.Settings()
		if conf.Demo() {
			api.AbortForbidden(c)
			return
		}

		sizeName := thumb.Name(clean.Token(c.Param("size")))
		fileHash := clean.Token(c.Param("hash"))

		size, ok := thumb.Sizes[sizeName]

		if !ok {
			log.Errorf("Photoshow GetThumb: No such size %s", size.Name)
			api.AbortBadRequest(c)
			return
		}

		if size.Uncached() && !conf.ThumbUncached() {
			sizeName, size = thumb.Find(conf.ThumbSizePrecached())

			if sizeName == "" {
				log.Errorf("Photoshow GetThumb: Invalid size name %s", size.Name)

				api.AbortBadRequest(c)
				return
			}
		}

		cache := get.ThumbCache()
		cacheKey := api.CacheKey("thumbs", fileHash, string(sizeName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			cached := cacheData.(api.ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("Photoshow GetThumb: cache file doesn't exist '%s'", cached.FileName)

				api.AbortBadRequest(c)
				return
			}

			// Add HTTP cache header.
			api.AddImmutableCacheHeader(c)
			c.File(cached.FileName)
			return
		}

		if fileName, err := size.ResolvedName(fileHash, conf.ThumbCachePath()); err == nil {
			// Add HTTP cache header.
			api.AddImmutableCacheHeader(c)

			// Return requested content.
			c.File(fileName)
			return
		}

		// Query index for file infos.
		f, err := query.FileByHash(fileHash)

		if err != nil {
			log.Errorf("Photoshow GetThumb: file by hash doesn't exist '%s'", fileHash)

			api.AbortBadRequest(c)
			return
		}

		// Find supported preview image if media file is not a JPEG or PNG.
		if f.NoJpeg() && f.NoPng() {
			log.Errorf("Photoshow GetThumb: only images allowed, got '%s'", f.FileMime)

			api.AbortBadRequest(c)
			return
		}

		// Return SVG icon as placeholder if file has errors.
		if f.FileError != "" {
			log.Errorf("Photoshow GetThumb: file error '%s'", f.FileError)

			api.AbortBadRequest(c)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if fileName, err = fs.Resolve(fileName); err != nil {
			log.Errorf("Photoshow GetThumb: can't resolve filename '%s'", fileName)
			api.AbortBadRequest(c)
			return
		}

		// Choose the smallest fitting size if the original image is smaller.
		if size.Fit && f.Bounds().In(size.Bounds()) {
			size = thumb.FitBounds(f.Bounds())
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() {

			// Add HTTP cache header.
			api.AddImmutableCacheHeader(c)

			// Return requested content.
			c.File(fileName)
			return
		}

		// thumbName is the thumbnail filename.
		var thumbName string

		// Try to find or create thumbnail image.
		if conf.ThumbUncached() || size.Uncached() {
			thumbName, err = size.FromFile(fileName, f.FileHash, conf.ThumbCachePath(), f.FileOrientation)
		} else {
			thumbName, err = size.FromCache(fileName, f.FileHash, conf.ThumbCachePath())
		}

		// Failed?
		if err != nil {
			log.Errorf("Photoshow GetThumb: can't generate thumb '%s'", err)
			api.AbortBadRequest(c)

			return
		} else if thumbName == "" {
			log.Error("Photoshow GetThumb: can't generate thumb, name empty")
			api.AbortBadRequest(c)

			return
		}

		// Cache thumbnail filename to reduce the number of index queries.
		cache.SetDefault(cacheKey, api.ThumbCache{thumbName, f.ShareBase(0)})

		// Add HTTP cache header.
		api.AddImmutableCacheHeader(c)

		// Return requested content.
		c.File(thumbName)

	})
}
