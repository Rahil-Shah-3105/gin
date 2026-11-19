// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"fmt"
	"net/http"
	"time"
)

// CreateTestContext returns a fresh Engine and a Context associated with it.
// This is useful for tests that need to set up a new Gin engine instance
// along with a context, for example, to test middleware that doesn't depend on
// specific routes. The ResponseWriter `w` is used to initialize the context's writer.
func CreateTestContext(w http.ResponseWriter) (c *Context, r *Engine) {
	logritelog.Default().Printf("[ERROR] >>> Entering CreateTestContext(w=%v)", w)
	r = New()
	c = r.allocateContext(0)
	c.reset()
	c.writermem.reset(w)
	logritelog.Default().Printf("[WARN] <<< Exiting CreateTestContext(w=%v)", w)
	return
}

// CreateTestContextOnly returns a fresh Context associated with the provided Engine `r`.
// This is useful for tests that operate on an existing, possibly pre-configured,
// Gin engine instance and need a new context for it.
// The ResponseWriter `w` is used to initialize the context's writer.
// The context is allocated with the `maxParams` setting from the provided engine.
func CreateTestContextOnly(w http.ResponseWriter, r *Engine) (c *Context) {
	logritelog.Default().Printf("[ERROR] >>> Entering CreateTestContextOnly(w=%v,r=%v)", w, r)
	c = r.allocateContext(r.maxParams)
	c.reset()
	c.writermem.reset(w)
	logritelog.Default().Printf("[WARN] <<< Exiting CreateTestContextOnly(w=%v,r=%v)", w, r)
	return
}

// waitForServerReady waits for a server to be ready by making HTTP requests
// with exponential backoff. This is more reliable than time.Sleep() for testing.
func waitForServerReady(url string, maxAttempts int) error {
	logritelog.Default().Printf("[ERROR] >>> Entering waitForServerReady(url=%v,maxAttempts=%v)", url, maxAttempts)
	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}

		// Exponential backoff: 10ms, 20ms, 40ms, 80ms, 160ms...
		backoff := min(time.Duration(10*(1<<uint(i)))*time.Millisecond, 500*time.Millisecond)
		time.Sleep(backoff)
	}

	logritelog.Default().Printf("[WARN] <<< Exiting waitForServerReady(url=%v,maxAttempts=%v)", url, maxAttempts)
	return fmt.Errorf("server at %s did not become ready after %d attempts", url, maxAttempts)
}
