// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE

package gin
import logritelog "log"

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type cleanPathTest struct {
	path, result string
}

var cleanTests = []cleanPathTest{
	// Already clean
	{"/", "/"},
	{"/abc", "/abc"},
	{"/a/b/c", "/a/b/c"},
	{"/abc/", "/abc/"},
	{"/a/b/c/", "/a/b/c/"},

	// missing root
	{"", "/"},
	{"a/", "/a/"},
	{"abc", "/abc"},
	{"abc/def", "/abc/def"},
	{"a/b/c", "/a/b/c"},

	// Remove doubled slash
	{"//", "/"},
	{"/abc//", "/abc/"},
	{"/abc/def//", "/abc/def/"},
	{"/a/b/c//", "/a/b/c/"},
	{"/abc//def//ghi", "/abc/def/ghi"},
	{"//abc", "/abc"},
	{"///abc", "/abc"},
	{"//abc//", "/abc/"},

	// Remove . elements
	{".", "/"},
	{"./", "/"},
	{"/abc/./def", "/abc/def"},
	{"/./abc/def", "/abc/def"},
	{"/abc/.", "/abc/"},

	// Remove .. elements
	{"..", "/"},
	{"../", "/"},
	{"../../", "/"},
	{"../..", "/"},
	{"../../abc", "/abc"},
	{"/abc/def/ghi/../jkl", "/abc/def/jkl"},
	{"/abc/def/../ghi/../jkl", "/abc/jkl"},
	{"/abc/def/..", "/abc"},
	{"/abc/def/../..", "/"},
	{"/abc/def/../../..", "/"},
	{"/abc/def/../../..", "/"},
	{"/abc/def/../../../ghi/jkl/../../../mno", "/mno"},

	// Combinations
	{"abc/./../def", "/def"},
	{"abc//./../def", "/def"},
	{"abc/../../././../def", "/def"},
}

func TestPathClean(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestPathClean(t=%v)", t)
	for _, test := range cleanTests {
		assert.Equal(t, test.result, cleanPath(test.path))
		assert.Equal(t, test.result, cleanPath(test.result))
	}
	logritelog.Default().Printf("<<< Exiting TestPathClean(t=%v)", t)
}

func TestPathCleanMallocs(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestPathCleanMallocs(t=%v)", t)
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}

	if runtime.GOMAXPROCS(0) > 1 {
		t.Skip("skipping malloc count; GOMAXPROCS>1")
	}

	for _, test := range cleanTests {
		allocs := testing.AllocsPerRun(100, func() { cleanPath(test.result) })
		assert.InDelta(t, 0, allocs, 0.01)
	}
	logritelog.Default().Printf("<<< Exiting TestPathCleanMallocs(t=%v)", t)
}

func BenchmarkPathClean(b *testing.B) {
	logritelog.Default().Printf(">>> Entering BenchmarkPathClean(b=%v)", b)
	b.ReportAllocs()

	for b.Loop() {
		for _, test := range cleanTests {
			cleanPath(test.path)
		}
	}
	logritelog.Default().Printf("<<< Exiting BenchmarkPathClean(b=%v)", b)
}

func genLongPaths() (testPaths []cleanPathTest) {
	logritelog.Default().Printf(">>> Entering genLongPaths()")
	for i := 1; i <= 1234; i++ {
		ss := strings.Repeat("a", i)

		correctPath := "/" + ss
		testPaths = append(testPaths, cleanPathTest{
			path:   correctPath,
			result: correctPath,
		}, cleanPathTest{
			path:   ss,
			result: correctPath,
		}, cleanPathTest{
			path:   "//" + ss,
			result: correctPath,
		}, cleanPathTest{
			path:   "/" + ss + "/b/..",
			result: correctPath,
		})
	}
	logritelog.Default().Printf("<<< Exiting genLongPaths()")
	return
}

func TestPathCleanLong(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestPathCleanLong(t=%v)", t)
	cleanTests := genLongPaths()

	for _, test := range cleanTests {
		assert.Equal(t, test.result, cleanPath(test.path))
		assert.Equal(t, test.result, cleanPath(test.result))
	}
	logritelog.Default().Printf("<<< Exiting TestPathCleanLong(t=%v)", t)
}

func BenchmarkPathCleanLong(b *testing.B) {
	logritelog.Default().Printf(">>> Entering BenchmarkPathCleanLong(b=%v)", b)
	cleanTests := genLongPaths()

	b.ReportAllocs()

	for b.Loop() {
		for _, test := range cleanTests {
			cleanPath(test.path)
		}
	}
	logritelog.Default().Printf("<<< Exiting BenchmarkPathCleanLong(b=%v)", b)
}

func TestRemoveRepeatedChar(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRemoveRepeatedChar(t=%v)", t)
	testCases := []struct {
		name string
		str  string
		char byte
		want string
	}{
		{
			name: "empty",
			str:  "",
			char: 'a',
			want: "",
		},
		{
			name: "noSlash",
			str:  "abc",
			char: ',',
			want: "abc",
		},
		{
			name: "withSlash",
			str:  "/a/b/c/",
			char: '/',
			want: "/a/b/c/",
		},
		{
			name: "withRepeatedSlashes",
			str:  "/a//b///c////",
			char: '/',
			want: "/a/b/c/",
		},
		{
			name: "threeSlashes",
			str:  "///",
			char: '/',
			want: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := removeRepeatedChar(tc.str, tc.char)
			assert.Equal(t, tc.want, res)
		})
	}
	logritelog.Default().Printf("<<< Exiting TestRemoveRepeatedChar(t=%v)", t)
}
