// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
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
	logritelog.Default().Printf("[INFO] TestBasicAuth(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed", t)
}

func TestBasicAuthFails(t *testing.T) {
	assert.Panics(t, func() { processAccounts(nil) })
	assert.Panics(t, func() {
		processAccounts(Accounts{
			"":    "password",
			"foo": "bar",
		})
	})
	logritelog.Default().Printf("[INFO] TestBasicAuthFails(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed", t)
}

func TestBasicAuthSearchCredential(t *testing.T) {
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
	logritelog.Default().Printf("[INFO] TestBasicAuthSearchCredential(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed", t)
}

func TestBasicAuthAuthorizationHeader(t *testing.T) {
	assert.Equal(t, "Basic YWRtaW46cGFzc3dvcmQ=", authorizationHeader("admin", "password"))
	logritelog.Default().Printf("[INFO] TestBasicAuthAuthorizationHeader(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed", t)
}

func TestBasicAuthSucceed(t *testing.T) {
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
	logritelog.Default().Printf("[INFO] TestBasicAuthSucceed(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed for user %v", t, accounts)
}

func TestBasicAuth401(t *testing.T) {
	called := false
	accounts := Accounts{"foo": "bar"}
	router := New()
	router.Use(BasicAuth(accounts))
	router.GET("/login", func(c *Context) {
		called = true
		c.String(http.StatusOK, c.MustGet(AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Basic realm=\"Authorization Required\"", w.Header().Get("WWW-Authenticate"))
	logritelog.Default().Printf("[INFO] TestBasicAuth401(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed for user %v", t, accounts)
}

func TestBasicAuth401WithCustomRealm(t *testing.T) {
	called := false
	accounts := Accounts{"foo": "bar"}
	router := New()
	router.Use(BasicAuthForRealm(accounts, "My Custom \"Realm\""))
	router.GET("/login", func(c *Context) {
		called = true
		c.String(http.StatusOK, c.MustGet(AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Basic realm=\"My Custom \\\"Realm\\\"\"", w.Header().Get("WWW-Authenticate"))
	logritelog.Default().Printf("[INFO] TestBasicAuth401WithCustomRealm(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed for user %v", t, accounts)
}

func TestBasicAuthForProxySucceed(t *testing.T) {
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
	logritelog.Default().Printf("[INFO] TestBasicAuthForProxySucceed(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed for user %v", t, accounts)
}

func TestBasicAuthForProxy407(t *testing.T) {
	called := false
	accounts := Accounts{"foo": "bar"}
	router := New()
	router.Use(BasicAuthForProxy(accounts, ""))
	router.Any("/*proxyPath", func(c *Context) {
		called = true
		c.String(http.StatusOK, c.MustGet(AuthProxyUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusProxyAuthRequired, w.Code)
	assert.Equal(t, "Basic realm=\"Proxy Authorization Required\"", w.Header().Get("Proxy-Authenticate"))
	logritelog.Default().Printf("[INFO] TestBasicAuthForProxy407(t=%v): GDPR-AUDIT: Transfers to third countries — Password reset operation performed for user %v", t, accounts)
}
