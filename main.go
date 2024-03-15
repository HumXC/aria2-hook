package main

import (
	"fmt"
	"math/big"
	"os/exec"
	"sort"
	"strings"
	"time"

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

func (s Status) ConvertTo(cmd string) (result string) {
	m := map[string]string{
		"${ErrCode}":         s.ErrCode,
		"${ErrMsg}":          s.ErrMsg,
		"${Name}":            s.Name,
		"${TotalLength}":     s.TotalLength,
		"${CompletedLength}": s.CompletedLength,
		"${GID}":             s.GID,
	}
	result = cmd
	for k, v := range m {
		result = strings.Replace(result, k, v, -1)
	}
	return
}
func CallCommand(cmds []string, status Status) {
	for _, cmd := range cmds {
		if strings.HasPrefix(cmd, "ASYNC:") {
			go func(c string) {
				err := exec.Command("sh", "-c", status.ConvertTo(c)).Run()
				if err != nil {
					fmt.Println("Error on Command:", c, err)
				}
			}(strings.Trim(cmd, "ASYNC:"))
		} else {
			err := exec.Command("sh", "-c", status.ConvertTo(cmd)).Run()
			if err != nil {
				fmt.Println("Error on Command:", cmd, err)
			}
		}
	}
}

type listenerF func(client *arigo.Client, config Config) arigo.EventListener

func gen(eventName string, client *arigo.Client, cmd []string) arigo.EventListener {
	return func(event *arigo.DownloadEvent) {
		s, err := GetStatus(client, event.GID)
		if err != nil {
			panic(err)
		}
		fmt.Println(eventName, "Event:", s)
		CallCommand(cmd, s)
	}
}

var events = map[arigo.EventType]listenerF{
	arigo.StartEvent: func(client *arigo.Client, config Config) arigo.EventListener {
		return gen("Start", client, config.OnDownloadStart)
	},
	arigo.PauseEvent: func(client *arigo.Client, config Config) arigo.EventListener {
		return gen("Pause", client, config.OnDownloadPause)
	},
	arigo.StopEvent: func(client *arigo.Client, config Config) arigo.EventListener {
		return gen("Stop", client, config.OnDownloadStop)
	},
	arigo.CompleteEvent: func(client *arigo.Client, config Config) arigo.EventListener {
		return gen("DownloadComplete", client, config.OnDownloadComplete)
	},
	arigo.BTCompleteEvent: func(client *arigo.Client, config Config) arigo.EventListener {
		return gen("BTDownloadComplete", client, config.OnBtDownloadComplete)
	},
	arigo.ErrorEvent: func(client *arigo.Client, config Config) arigo.EventListener {
		return gen("Error", client, config.OnDownloadError)
	},
}

func main() {
	rpc2.DebugLog = true
	config, err := ParseConfigFile("config.yaml")
	if err != nil {
		panic(err)
	}
	for {
		if err := Run(config); err != nil {
			fmt.Println("Client has error:", err, ",reconnect after 10s")
			time.Sleep(10 * time.Second)
		}
		fmt.Println("Client was closed, reconnecting.")
	}
}
func Run(config Config) error {
	c, err := arigo.Dial(config.Url, config.Token)
	if err != nil {
		return err
	}
	for event, listener := range events {
		c.Subscribe(event, listener(c, config))
	}
	c.Run()
	return nil
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
