// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import (
	logritelog "log"
	logritetime "time"
)

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuth(t=%v)", t)
	pairs := processAccounts(Accounts{
		"admin": "password",
		"foo":   "bar",
		"bar":   "foo",
	})

	assert.Len(t, pairs, 3)
	assert.Contains(t, pairs, authPair{
		user:  "bar",
		value: "Basic YmFyOmZvbw==",
	})
	assert.Contains(t, pairs, authPair{
		user:  "foo",
		value: "Basic Zm9vOmJhcg==",
	})
	assert.Contains(t, pairs, authPair{
		user:  "admin",
		value: "Basic YWRtaW46cGFzc3dvcmQ=",
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuth(t=%v)", t)
}

func TestBasicAuthFails(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuthFails(t=%v)", t)
	assert.Panics(t, func() { processAccounts(nil) })
	assert.Panics(t, func() {
		processAccounts(Accounts{
			"":    "password",
			"foo": "bar",
		})
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuthFails(t=%v)", t)
}

func TestBasicAuthSearchCredential(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuthSearchCredential(t=%v)", t)
	pairs := processAccounts(Accounts{
		"admin": "password",
		"foo":   "bar",
		"bar":   "foo",
	})

	user, found := pairs.searchCredential(authorizationHeader("admin", "password"))
	assert.Equal(t, "admin", user)
	assert.True(t, found)

	user, found = pairs.searchCredential(authorizationHeader("foo", "bar"))
	assert.Equal(t, "foo", user)
	assert.True(t, found)

	user, found = pairs.searchCredential(authorizationHeader("bar", "foo"))
	assert.Equal(t, "bar", user)
	assert.True(t, found)

	user, found = pairs.searchCredential(authorizationHeader("admins", "password"))
	assert.Empty(t, user)
	assert.False(t, found)

	user, found = pairs.searchCredential(authorizationHeader("foo", "bar "))
	assert.Empty(t, user)
	assert.False(t, found)

	user, found = pairs.searchCredential("")
	assert.Empty(t, user)
	assert.False(t, found)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuthSearchCredential(t=%v)", t)
}

func TestBasicAuthAuthorizationHeader(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuthAuthorizationHeader(t=%v)", t)
	assert.Equal(t, "Basic YWRtaW46cGFzc3dvcmQ=", authorizationHeader("admin", "password"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuthAuthorizationHeader(t=%v)", t)
}

func TestBasicAuthSucceed(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuthSucceed(t=%v)", t)
	__logStart := logritetime.Now()
	accounts := Accounts{"admin": "password"}
	router := New()
	router.Use(BasicAuth(accounts))
	router.GET("/login", func(c *Context) {
		c.String(http.StatusOK, c.MustGet(AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	req.Header.Set("Authorization", authorizationHeader("admin", "password"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "admin", w.Body.String())
	logritelog.Default().Printf("[ERROR] TestBasicAuthSucceed(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuthSucceed(t=%v)", t)
}

func TestBasicAuth401(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuth401(t=%v)", t)
	called := false
	accounts := Accounts{"foo": "bar"}
	router := New()
	router.Use(BasicAuth(accounts))
	router.GET("/login", func(c *Context) {
		called = true
		logritelog.Default().Printf("[DEBUG] TestBasicAuth401(t=%v): called → %v", t, called)
		c.String(http.StatusOK, c.MustGet(AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Basic realm=\"Authorization Required\"", w.Header().Get("WWW-Authenticate"))
	logritelog.Default().Printf("[ERROR] TestBasicAuth401(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuth401(t=%v)", t)
}

func TestBasicAuth401WithCustomRealm(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuth401WithCustomRealm(t=%v)", t)
	called := false
	accounts := Accounts{"foo": "bar"}
	router := New()
	router.Use(BasicAuthForRealm(accounts, "My Custom \"Realm\""))
	router.GET("/login", func(c *Context) {
		called = true
		logritelog.Default().Printf("[DEBUG] TestBasicAuth401WithCustomRealm(t=%v): called → %v", t, called)
		c.String(http.StatusOK, c.MustGet(AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Basic realm=\"My Custom \\\"Realm\\\"\"", w.Header().Get("WWW-Authenticate"))
	logritelog.Default().Printf("[ERROR] TestBasicAuth401WithCustomRealm(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuth401WithCustomRealm(t=%v)", t)
}

func TestBasicAuthForProxySucceed(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuthForProxySucceed(t=%v)", t)
	accounts := Accounts{"admin": "password"}
	router := New()
	router.Use(BasicAuthForProxy(accounts, ""))
	router.Any("/*proxyPath", func(c *Context) {
		c.String(http.StatusOK, c.MustGet(AuthProxyUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Proxy-Authorization", authorizationHeader("admin", "password"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "admin", w.Body.String())
	logritelog.Default().Printf("[ERROR] TestBasicAuthForProxySucceed(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuthForProxySucceed(t=%v)", t)
}

func TestBasicAuthForProxy407(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBasicAuthForProxy407(t=%v)", t)
	called := false
	accounts := Accounts{"foo": "bar"}
	router := New()
	router.Use(BasicAuthForProxy(accounts, ""))
	router.Any("/*proxyPath", func(c *Context) {
		called = true
		logritelog.Default().Printf("[DEBUG] TestBasicAuthForProxy407(t=%v): called → %v", t, called)
		c.String(http.StatusOK, c.MustGet(AuthProxyUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusProxyAuthRequired, w.Code)
	assert.Equal(t, "Basic realm=\"Proxy Authorization Required\"", w.Header().Get("Proxy-Authenticate"))
	logritelog.Default().Printf("[ERROR] TestBasicAuthForProxy407(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestBasicAuthForProxy407(t=%v)", t)
}
