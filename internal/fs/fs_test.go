package fs
import logritelog "log"

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFileSystem struct {
	open func(name string) (http.File, error)
}

func (m *mockFileSystem) Open(name string) (http.File, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Open(name=%v)", name)
	return m.open(name)
}

func TestFileSystem_Open(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFileSystem_Open(t=%v)", t)
	var testFile *os.File
	mockFS := &mockFileSystem{
		open: func(name string) (http.File, error) {
			return testFile, nil
		},
	}
	fs := &FileSystem{mockFS}

	file, err := fs.Open("foo")

	require.NoError(t, err)
	assert.Equal(t, testFile, file)
	logritelog.Default().Printf("[WARN] <<< Exiting TestFileSystem_Open(t=%v)", t)
}

func TestFileSystem_Open_err(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFileSystem_Open_err(t=%v)", t)
	testError := errors.New("mock")
	mockFS := &mockFileSystem{
		open: func(_ string) (http.File, error) {
			return nil, testError
		},
	}
	fs := &FileSystem{mockFS}

	file, err := fs.Open("foo")

	require.ErrorIs(t, err, testError)
	assert.Nil(t, file)
	logritelog.Default().Printf("[WARN] <<< Exiting TestFileSystem_Open_err(t=%v)", t)
}
