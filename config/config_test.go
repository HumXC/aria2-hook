package config

import (
	"flag"
	"os"
	"testing"

	"github.com/zeebo/assert"
)

func TestVerify(t *testing.T) {
	cases := []struct {
		want bool
		cfg  Config
	}{
		{
			want: true,
			cfg: Config{
				Url:             "https://domain:6800/jsonrpc",
				Token:           "5k5kDSWKJ%^$!@[+]",
				OnDownloadStart: []string{"ASYNC:touch ${Name}"},
				OnDownloadPause: []string{""},
			},
		},
		{
			want: false,
			cfg: Config{
				Url: "",
			},
		},
	}

	for _, c := range cases {
		got := Verify(c.cfg)
		assert.Equal(t, c.want, got == nil)
	}
}
func TestSFromFile(t *testing.T) {
	want := Config{
		Url:             "https://domain:6800/jsonrpc",
		Token:           "5k5kDSWKJ%^$!@[+]",
		OnDownloadStart: []string{"ASYNC:touch ${Name}"},
		OnDownloadPause: []string{""},
	}
	var conf = []byte(`
url: "https://domain:6800/jsonrpc"
token: "5k5kDSWKJ%^$!@[+]"
onDownloadStart:
    - "ASYNC:touch ${Name}"
onDownloadPause:
    - ""
onDownloadStop: 
onDownloadComplete: 
onDownloadError: 
onBtDownloadComplete:`)
	got, err := SFromFile(conf)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
func TestFromEnv(t *testing.T) {
	want := Config{
		Url:             "https://domain:6800/jsonrpc",
		Token:           "5k5kDSWKJ%^$!@[]",
		OnDownloadStart: []string{"ASYNC:touch ${Name}", ""},
		OnDownloadPause: []string{"", "echo hi"},
	}
	os.Setenv("URL", "https://domain:6800/jsonrpc")
	os.Setenv("TOKEN", "5k5kDSWKJ%^$!@[]")
	os.Setenv("ON_DOWNLOAD_START", "ASYNC:touch ${Name}#")
	os.Setenv("ON_DOWNLOAD_PAUSE", "#echo hi")
	got, err := FromEnv()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
func TestSFromCmd(t *testing.T) {
	want := Config{
		Url:             "https://domain:6800/jsonrpc",
		Token:           "5k5kDSWKJ%^$!@[]",
		OnDownloadStart: []string{"ASYNC:touch ${Name}", ""},
		OnDownloadPause: []string{"", "echo hi"},
	}
	args := []string{
		"aria2-hook-test",
		"--url", "https://domain:6800/jsonrpc",
		"--token", "5k5kDSWKJ%^$!@[]",
		"--on-download-start", "ASYNC:touch ${Name}#",
		"--on-download-pause", "#echo hi",
	}
	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	got, err := SFromCmd(cmd, args)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
func TestMerge(t *testing.T) {
	c1 := Config{
		Url:             "-url",
		OnDownloadStart: []string{"a"},
	}
	c2 := Config{
		Url:             "-url2",
		OnDownloadStart: []string{"b"},
		OnDownloadStop:  []string{"stop"},
	}
	c3 := Config{
		Url: "-url3",
	}
	want := Config{
		Url:             "-url3",
		OnDownloadStart: []string{"a", "b"},
		OnDownloadStop:  []string{"stop"},
	}
	got := Merge(c1, c2, c3)
	assert.Equal(t, want, got)
}
