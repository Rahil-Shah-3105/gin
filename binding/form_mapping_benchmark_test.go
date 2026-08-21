// Copyright 2019 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var form = map[string][]string{
	"name":      {"mike"},
	"friends":   {"anna", "nicole"},
	"id_number": {"12345678"},
	"id_date":   {"2018-01-20"},
}

type structFull struct {
	Name    string   `form:"name"`
	Age     int      `form:"age,default=25"`
	Friends []string `form:"friends"`
	ID      *struct {
		Number      string    `form:"id_number"`
		DateOfIssue time.Time `form:"id_date" time_format:"2006-01-02" time_utc:"true"`
	}
	Nationality *string `form:"nationality"`
}

func BenchmarkMapFormFull(b *testing.B) {
	logritelog.Default().Printf(">>> Entering BenchmarkMapFormFull(b=%v)", b)
	var s structFull
	for b.Loop() {
		err := mapForm(&s, form)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in BenchmarkMapFormFull(b=%v): %v", b, err)
			b.Fatalf("Error on a form mapping")
		}
	}
	b.StopTimer()

	t := b
	assert.Equal(t, "mike", s.Name)
	assert.Equal(t, 25, s.Age)
	assert.Equal(t, []string{"anna", "nicole"}, s.Friends)
	assert.Equal(t, "12345678", s.ID.Number)
	assert.Equal(t, time.Date(2018, 1, 20, 0, 0, 0, 0, time.UTC), s.ID.DateOfIssue)
	assert.Nil(t, s.Nationality)
	logritelog.Default().Printf("<<< Exiting BenchmarkMapFormFull(b=%v)", b)
}

type structName struct {
	Name string `form:"name"`
}

func BenchmarkMapFormName(b *testing.B) {
	logritelog.Default().Printf(">>> Entering BenchmarkMapFormName(b=%v)", b)
	var s structName
	for b.Loop() {
		err := mapForm(&s, form)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in BenchmarkMapFormName(b=%v): %v", b, err)
			b.Fatalf("Error on a form mapping")
		}
	}
	b.StopTimer()

	t := b
	assert.Equal(t, "mike", s.Name)
	logritelog.Default().Printf("<<< Exiting BenchmarkMapFormName(b=%v)", b)
}
