package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// FoldersResponse represents the folders API response.
type FoldersResponse struct {
	Root      string          `json:"root,omitempty"`
	Folders   []entity.Folder `json:"folders"`
	Files     []entity.File   `json:"files,omitempty"`
	Recursive bool            `json:"recursive,omitempty"`
	Cached    bool            `json:"cached,omitempty"`
}

func MoveFiles(router *gin.RouterGroup) {
	router.PUT("/files/move", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionCreate)

		if s.Abort(c) {
			return
		}
		conf := get.Config()
		root := conf.OriginalsPath()
		var frm form.MoveFiles

		err := c.BindJSON(&frm)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		for _, v := range frm.Files {
			file, err := entity.PrimaryFile(v.PhotoUID)
			if err != nil {
				log.Errorf("failed to move photo: %w", err)
				AbortUnexpectedError(c)
				return
			}

			sourcePath := filepath.Join(root, file.FileName)
			destPath := filepath.Join(root, v.Destination, file.Base(0))

			log.Infof("Moving file %s to %s", sourcePath, destPath)

			inputFile, err := os.Open(sourcePath)
			if err != nil {
				log.Errorf("failed to move photo: %w", err)
				AbortUnexpectedError(c)
				return
			}
			defer inputFile.Close()

			outputFile, err := os.Create(destPath)
			if err != nil {
				log.Errorf("failed to move photo: %w", err)
				AbortUnexpectedError(c)
				return
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, inputFile)
			if err != nil {
				log.Errorf("failed to move photo: %w", err)
				AbortUnexpectedError(c)
				return
			}

			inputFile.Close()

			err = os.Remove(sourcePath)
			if err != nil {
				log.Errorf("failed to move photo: %w", err)
				AbortUnexpectedError(c)
				return
			}

			err = file.Rename(filepath.Join(v.Destination, file.Base(0)), file.FileRoot, v.Destination, file.RelatedPhoto().PhotoName)
			if err != nil {
				log.Errorf("failed to rename file: %w", err)
				AbortUnexpectedError(c)
				return
			}

		}

		c.JSON(http.StatusOK, http.Response{})
	})
}

func CreateNewFolder(router *gin.RouterGroup) {
	router.PUT("/folder/create", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionCreate)

		if s.Abort(c) {
			return
		}
		conf := get.Config()
		var frm form.NewFolderForm

		err := c.BindJSON(&frm)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		root := conf.OriginalsPath()
		log.Infof("creating folder %s", clean.Log(filepath.Join(root, frm.Path)))

		// Join the path and folder name into a full path
		fullPath := filepath.Join(root, frm.Path)

		// Check if the folder already exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// Create the folder and any necessary parents
			err := os.MkdirAll(fullPath, os.ModePerm)
			if err != nil {
				log.Errorf("failed to create folder: %w", err)
				AbortUnexpectedError(c)
				return
			}
		} else {
			log.Errorf("Folder already exists: %s", fullPath)
			AbortUnexpectedError(c)
			return
		}

		folder := entity.NewFolder(entity.RootOriginals, frm.Path, entity.Now())

		if err := folder.Create(); err != nil {
			// Report unexpected error.
			log.Errorf("folder create: %s)", err)
			AbortUnexpectedError(c)
			return
		}

		c.JSON(http.StatusOK, folder)
	})

}

// SearchFoldersOriginals returns folders in originals as JSON.
//
// GET /api/v1/folders/originals
func SearchFoldersOriginals(router *gin.RouterGroup) {
	conf := get.Config()
	SearchFolders(router, "originals", entity.RootOriginals, conf.OriginalsPath())
}

// SearchFoldersImport returns import folders as JSON.
//
// GET /api/v1/folders/import
func SearchFoldersImport(router *gin.RouterGroup) {
	conf := get.Config()
	SearchFolders(router, "import", entity.RootImport, conf.ImportPath())
}

// SearchFolders is a reusable request handler for directory listings (GET /api/v1/folders/*).
func SearchFolders(router *gin.RouterGroup, urlPath, rootName, rootPath string) {
	handler := func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.AccessLibrary)

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		var frm form.SearchFolders

		start := time.Now()
		err := c.MustBindWith(&frm, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		user := s.User()
		aclRole := user.AclRole()

		// Exclude private content?
		if !get.Config().Settings().Features.Private {
			frm.Public = false
		} else if acl.Rules.Deny(acl.ResourcePhotos, aclRole, acl.AccessPrivate) {
			frm.Public = true
		}

		cache := get.FolderCache()
		recursive := frm.Recursive
		listFiles := frm.Files
		uncached := listFiles || frm.Uncached
		resp := FoldersResponse{Root: rootName, Recursive: recursive, Cached: !uncached}
		path := clean.UserPath(c.Param("path"))

		cacheKey := fmt.Sprintf("folder:%s:%t:%t:%t", filepath.Join(rootName, path), recursive, listFiles, frm.Public)

		if !uncached {
			if cacheData, ok := cache.Get(cacheKey); ok {
				cached := cacheData.(FoldersResponse)

				log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

				c.JSON(http.StatusOK, cached)
				return
			}
		}

		if folders, err := query.FoldersByPath(rootName, rootPath, path, recursive); err != nil {
			log.Errorf("folder: %s", err)
			c.JSON(http.StatusOK, resp)
			return
		} else {
			resp.Folders = folders
		}

		if listFiles {
			if files, err := query.FilesByPath(frm.Count, frm.Offset, rootName, path, frm.Public); err != nil {
				log.Errorf("folder: %s", err)
			} else {
				resp.Files = files
			}
		}

		if !uncached {
			cache.SetDefault(cacheKey, resp)
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

		AddFileCountHeaders(c, len(resp.Files), len(resp.Folders))
		AddCountHeader(c, len(resp.Files)+len(resp.Folders))
		AddLimitHeader(c, frm.Count)
		AddOffsetHeader(c, frm.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, resp)
	}

	router.GET("/folders/"+urlPath, handler)
	router.GET("/folders/"+urlPath+"/*path", handler)
}
