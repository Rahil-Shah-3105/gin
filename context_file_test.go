package gin
import logritelog "log"

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContextFileSimple tests the Context.File() method with a simple case
func TestContextFileSimple(t *testing.T) {
	// Test serving an existing file
	testFile := "testdata/test_file.txt"
	w := httptest.NewRecorder()
	c, _ := CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	c.File(testFile)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "This is a test file")
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[INFO] TestContextFileSimple(t=%v): SOC2-AUDIT [CC7.1]: System operation completed", t)
}

// TestContextFileNotFound tests serving a non-existent file
func TestContextFileNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	c.File("non_existent_file.txt")

	assert.Equal(t, http.StatusNotFound, w.Code)
	logritelog.Default().Printf("[INFO] TestContextFileNotFound(t=%v): SOC2-AUDIT [CC7.1]: System operation completed", t)
}
