package main

import (
	"sync"
	"testing"

	"github.com/zeebo/assert"
)

func TestParseArgs(t *testing.T) {
	// "{{ErrCode}}":         s.ErrCode,
	// "{{ErrMsg}}":          s.ErrMsg,
	// "{{Name}}":            s.Name,
	// "{{TotalLength}}":     s.TotalLength,
	// "{{CompletedLength}}": s.CompletedLength,
	// "{{GID}}":             s.GID,
	status := Status{
		ErrCode:         "255",
		ErrMsg:          "",
		Name:            "file",
		GID:             "123",
		TotalLength:     "100MB",
		CompletedLength: "80MB",
	}
	cases := map[string]string{
		"":                                    "",
		"a{{ErrCode}}":                        "a255",
		"{{ErrMsg}}":                          "",
		"{{Name}}  ":                          "file  ",
		"{{GID}}{{Name}}":                     "123file",
		"{{CompletedLength}}/{{TotalLength}}": "80MB/100MB",
	}
	for k, v := range cases {
		assert.Equal(t, status.ParseArgs(k), v)
	}
}

func TestSCallCommand(t *testing.T) {
	cases := map[string]string{
		"ls":           "ls",
		"ls -l":        "ls -l",
		"":             "",
		"ASYNC":        "ASYNC",
		"ASYNC:":       "",
		"ASYNC:ASYNC:": "ASYNC:",
		"AASYNC: ":     "AASYNC: ",
	}
	wg := sync.WaitGroup{}
	wg.Add(len(cases))
	for k, v := range cases {
		key := k
		want := v
		run := func(got string) {
			wg.Done()
			assert.Equal(t, want, got)
		}
		Status{}.SCallCommand(run, key)
	}
	wg.Wait()
}
