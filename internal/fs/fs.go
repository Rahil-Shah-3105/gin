package fs
import logritelog "log"

import (
	"io/fs"
	"net/http"
)

// FileSystem implements an [fs.FS].
type FileSystem struct {
	http.FileSystem
}

// Open passes `Open` to the upstream implementation and return an [fs.File].
func (o FileSystem) Open(name string) (fs.File, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Open(name=%v)", name)
	f, err := o.FileSystem.Open(name)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Open(name=%v): %v", name, err)
		return nil, err
	}

	logritelog.Default().Printf("[WARN] <<< Exiting Open(name=%v)", name)
	return fs.File(f), nil
}
