// Copyright 2020 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bytesconv
import logritelog "log"

import (
	"bytes"
	cRand "crypto/rand"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var (
	testString = "Albert Einstein: Logic will get you from A to B. Imagination will take you everywhere."
	testBytes  = []byte(testString)
)

func rawBytesToStr(b []byte) string {
	logritelog.Default().Printf("[ERROR] >>> Entering rawBytesToStr(b=%v)", b)
	logritelog.Default().Printf("[WARN] <<< Exiting rawBytesToStr(b=%v)", b)
	return string(b)
}

func rawStrToBytes(s string) []byte {
	logritelog.Default().Printf("[ERROR] >>> Entering rawStrToBytes(s=%v)", s)
	logritelog.Default().Printf("[WARN] <<< Exiting rawStrToBytes(s=%v)", s)
	return []byte(s)
}

// go test -v

func TestBytesToString(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBytesToString(t=%v)", t)
	data := make([]byte, 1024)
	for range 100 {
		_, err := cRand.Read(data)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in TestBytesToString(t=%v): %v", t, err)
			t.Fatal(err)
		}
		if rawBytesToStr(data) != BytesToString(data) {
			t.Fatal("don't match")
		}
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestBytesToString(t=%v)", t)
}

func TestBytesToStringEmpty(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBytesToStringEmpty(t=%v)", t)
	if got := BytesToString([]byte{}); got != "" {
		t.Fatalf("BytesToString([]byte{}) = %q; want empty string", got)
	}
	if got := BytesToString(nil); got != "" {
		t.Fatalf("BytesToString(nil) = %q; want empty string", got)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestBytesToStringEmpty(t=%v)", t)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandStringBytesMaskImprSrcSB(n int) string {
	logritelog.Default().Printf("[ERROR] >>> Entering RandStringBytesMaskImprSrcSB(n=%v)", n)
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	logritelog.Default().Printf("[WARN] <<< Exiting RandStringBytesMaskImprSrcSB(n=%v)", n)
	return sb.String()
}

func TestStringToBytes(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestStringToBytes(t=%v)", t)
	for range 100 {
		s := RandStringBytesMaskImprSrcSB(64)
		if !bytes.Equal(rawStrToBytes(s), StringToBytes(s)) {
			t.Fatal("don't match")
		}
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestStringToBytes(t=%v)", t)
}

func TestStringToBytesEmpty(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestStringToBytesEmpty(t=%v)", t)
	b := StringToBytes("")
	if len(b) != 0 {
		t.Fatalf(`StringToBytes("") length = %d; want 0`, len(b))
	}
	if !bytes.Equal(b, []byte("")) {
		t.Fatalf(`StringToBytes("") = %v; want []byte("")`, b)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestStringToBytesEmpty(t=%v)", t)
}

// go test -v -run=none -bench=^BenchmarkBytesConv -benchmem=true

func BenchmarkBytesConvBytesToStrRaw(b *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkBytesConvBytesToStrRaw(b=%v)", b)
	for b.Loop() {
		rawBytesToStr(testBytes)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkBytesConvBytesToStrRaw(b=%v)", b)
}

func BenchmarkBytesConvBytesToStr(b *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkBytesConvBytesToStr(b=%v)", b)
	for b.Loop() {
		BytesToString(testBytes)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkBytesConvBytesToStr(b=%v)", b)
}

func BenchmarkBytesConvStrToBytesRaw(b *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkBytesConvStrToBytesRaw(b=%v)", b)
	for b.Loop() {
		rawStrToBytes(testString)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkBytesConvStrToBytesRaw(b=%v)", b)
}

func BenchmarkBytesConvStrToBytes(b *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkBytesConvStrToBytes(b=%v)", b)
	for b.Loop() {
		StringToBytes(testString)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkBytesConvStrToBytes(b=%v)", b)
}
