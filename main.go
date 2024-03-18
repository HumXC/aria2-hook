package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/HumXC/aria2-hook/config"

	"github.com/HumXC/arigo"
	"github.com/cenkalti/rpc2"
	"github.com/dustin/go-humanize"
)

type Status struct {
	GID             string
	ErrCode         string
	ErrMsg          string
	Name            string
	TotalLength     string
	CompletedLength string
}

func (s Status) String() string {
	return fmt.Sprintf("Name:%s,TotalLength:%s,CompletedLength:%s", s.Name, s.TotalLength, s.CompletedLength)
}

func (s Status) ParseArgs(cmd string) (result string) {
	m := map[string]string{
		"{{ErrCode}}":         s.ErrCode,
		"{{ErrMsg}}":          s.ErrMsg,
		"{{Name}}":            s.Name,
		"{{TotalLength}}":     s.TotalLength,
		"{{CompletedLength}}": s.CompletedLength,
		"{{GID}}":             s.GID,
	}
	result = cmd
	for k, v := range m {
		result = strings.Replace(result, k, v, -1)
	}
	return
}

func (s Status) CallCommand(cmds []string) {
	run := func(command string) {
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			log.Println("Error on Command:", string(out), err)
		}
	}
	for _, cmd := range cmds {
		s.SCallCommand(run, cmd)
	}
}

func (s Status) SCallCommand(run func(string), cmd string) {
	asyncFlag := "ASYNC:"
	if strings.HasPrefix(cmd, asyncFlag) {
		go run(s.ParseArgs(strings.TrimPrefix(cmd, asyncFlag)))
	} else {
		run(s.ParseArgs(cmd))
	}
}

type listener func(client *arigo.Client, config config.Config) arigo.EventListener

func MakeListener(eventName string, client *arigo.Client, cmd []string) arigo.EventListener {
	return func(event *arigo.DownloadEvent) {
		s, err := GetStatus(client, event.GID)
		if err != nil {
			log.Printf("Failed to get '%s' event status, GID:%s", eventName, event.GID)
			return
		}
		log.Println(eventName, "Event:", s)
		s.CallCommand(cmd)
	}
}

var events = map[arigo.EventType]listener{
	arigo.StartEvent: func(client *arigo.Client, config config.Config) arigo.EventListener {
		return MakeListener("Start", client, config.OnDownloadStart)
	},
	arigo.PauseEvent: func(client *arigo.Client, config config.Config) arigo.EventListener {
		return MakeListener("Pause", client, config.OnDownloadPause)
	},
	arigo.StopEvent: func(client *arigo.Client, config config.Config) arigo.EventListener {
		return MakeListener("Stop", client, config.OnDownloadStop)
	},
	arigo.CompleteEvent: func(client *arigo.Client, config config.Config) arigo.EventListener {
		return MakeListener("DownloadComplete", client, config.OnDownloadComplete)
	},
	arigo.BTCompleteEvent: func(client *arigo.Client, config config.Config) arigo.EventListener {
		return MakeListener("BTDownloadComplete", client, config.OnBtDownloadComplete)
	},
	arigo.ErrorEvent: func(client *arigo.Client, config config.Config) arigo.EventListener {
		return MakeListener("Error", client, config.OnDownloadError)
	},
}

func main() {
	// CommandLine > Environment > ConfigFile
	var cfg config.Config
	var err error
	checkErr := func(err error) {
		if err != nil {
			log.Fatalf("Failed to load config, %s\n", err)
		}
	}

	configFile := ""
	cmd := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cmd.StringVar(&configFile, "config", "", "config file")
	cfgCmd, err := config.FromCmd(cmd)
	checkErr(err)
	cfgEnv, err := config.FromEnv()
	checkErr(err)
	var cfgFile config.Config
	if configFile != "" {
		cfgFile, err = config.FromFile(configFile)
		checkErr(err)
	}
	cfg = config.Merge(cfgCmd, cfgEnv, cfgFile)
	checkErr(config.Verify(cfg))
	rpc2.DebugLog = cfg.Debug

	for {
		if err := Run(cfg); err != nil {
			log.Println("Client has error:", err, ",reconnect after 10s")
			time.Sleep(10 * time.Second)
		}
		log.Println("Client was closed, reconnecting.")
	}
}

func Run(config config.Config) error {
	url := config.Url
	if strings.HasPrefix(url, "http") {
		url = strings.Replace(config.Url, "http", "ws", 1)
	}
	c, err := arigo.Dial(url, config.Token)
	if err != nil {
		return err
	}
	for event, listener := range events {
		_ = c.Subscribe(event, listener(c, config))
	}
	return c.Run()
}

func GetTaskName(status arigo.Status) string {
	taskName := "Unknow"
	switch {
	case status.BitTorrent.Info.Name != "":
		taskName = status.BitTorrent.Info.Name
	case status.Files != nil && len(status.Files) > 0:
		sort.Slice(status.Files, func(i, j int) bool {
			return status.Files[i].Length > status.Files[j].Length
		})
		if len(status.Files[0].URIs) == 0 {
			break
		}
		uri := status.Files[0].URIs[0].URI
		index := strings.LastIndex(uri, "/")

		if index <= 0 || index == len(uri) {
			taskName = uri
		}
		fileNameAndQueryString := uri[index+1:]
		queryStringStartPos := strings.Index(fileNameAndQueryString, "?")
		taskName = fileNameAndQueryString

		if queryStringStartPos > 0 {
			taskName = fileNameAndQueryString[:queryStringStartPos]
		}
	}
	return taskName
}

func GetStatus(client *arigo.Client, gid string) (Status, error) {
	s, err := client.TellStatus(gid)
	if err != nil {
		return Status{}, err
	}
	var bi big.Int
	return Status{
		GID:             gid,
		ErrCode:         string(s.ErrorCode),
		ErrMsg:          s.ErrorMessage,
		Name:            GetTaskName(s),
		TotalLength:     humanize.BigBytes((bi.SetUint64(uint64(s.TotalLength)))),
		CompletedLength: humanize.BigBytes((bi.SetUint64(uint64(s.CompletedLength)))),
	}, nil
}
