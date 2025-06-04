package form

import (
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/pkg/clean"
)

// File represents a file edit form.
type File struct {
	FileOrientation int `json:"Orientation"`
}

type MoveFile struct {
	PhotoUID    string `json:"PhotoUID"`
	Destination string `json:"Dest"`
}

type MoveFiles struct {
	Files []MoveFile `json:"Files"`
}

// Orientation returns the Exif orientation value within a valid range or 0 if it is invalid.
func (f *File) Orientation() int {
	return clean.Orientation(f.FileOrientation)
}

func NewFile(m interface{}) (f File, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
