package gin
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
	logritelog.Default().Printf("[WARN] <<< Exiting Open(name=%v)", name)
	return m.open(name)
}

func TestOnlyFilesFS_Open(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestOnlyFilesFS_Open(t=%v)", t)
	var testFile *os.File
	mockFS := &mockFileSystem{
		open: func(name string) (http.File, error) {
			return testFile, nil
		},
	}
	fs := &OnlyFilesFS{FileSystem: mockFS}

	file, err := fs.Open("foo")

	require.NoError(t, err)
	assert.Equal(t, testFile, file.(neutralizedReaddirFile).File)
	logritelog.Default().Printf("[WARN] <<< Exiting TestOnlyFilesFS_Open(t=%v)", t)
}

func TestOnlyFilesFS_Open_err(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestOnlyFilesFS_Open_err(t=%v)", t)
	testError := errors.New("mock")
	mockFS := &mockFileSystem{
		open: func(_ string) (http.File, error) {
			return nil, testError
		},
	}
	fs := &OnlyFilesFS{FileSystem: mockFS}

	file, err := fs.Open("foo")

	require.ErrorIs(t, err, testError)
	assert.Nil(t, file)
	logritelog.Default().Printf("[WARN] <<< Exiting TestOnlyFilesFS_Open_err(t=%v)", t)
}

func Test_neuteredReaddirFile_Readdir(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering Test_neuteredReaddirFile_Readdir(t=%v)", t)
	n := neutralizedReaddirFile{}

	res, err := n.Readdir(0)

	require.NoError(t, err)
	assert.Nil(t, res)
	logritelog.Default().Printf("[WARN] <<< Exiting Test_neuteredReaddirFile_Readdir(t=%v)", t)
}

func TestDir_listDirectory(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDir_listDirectory(t=%v)", t)
	testRoot := "foo"
	fs := Dir(testRoot, true)

	assert.Equal(t, http.Dir(testRoot), fs)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDir_listDirectory(t=%v)", t)
}

func TestDir(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDir(t=%v)", t)
	testRoot := "foo"
	fs := Dir(testRoot, false)

	assert.Equal(t, &OnlyFilesFS{FileSystem: http.Dir(testRoot)}, fs)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDir(t=%v)", t)
}
