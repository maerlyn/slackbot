package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nlopes/slack"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

var (
	myName     string
	iconUrl    string
	slackToken string

	dropboxCommandSocket string
	rtorrentAddress      string

	postMessageParameters slack.PostMessageParameters
	slackRtm              *slack.RTM
)

func init() {
	viper.SetConfigFile("config.toml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Error reading config file: %+v\n", err))
	}

	myName = viper.GetString("my_name")
	iconUrl = viper.GetString("icon_url")
	slackToken = viper.GetString("token")

	dropboxCommandSocket = viper.GetString("dropbox_command_socket")
	rtorrentAddress = viper.GetString("rtorrent_addr")

	postMessageParameters = slack.PostMessageParameters{
		Username: myName,
		IconURL:  iconUrl,
	}
}

func main() {
	api := slack.New(slackToken)

	slackRtm = api.NewRTM()
	go slackRtm.ManageConnection()

	slackRtm.JoinChannel("torrent")

	go MonitorDropboxStatus()

	slackRtm.PostMessage("torrent", "I'm alive", postMessageParameters)

	for {
		select {
		case msg := <-slackRtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if !strings.HasPrefix(ev.Msg.Text, myName) {
					continue
				}

				info, _ := api.GetChannelInfo(ev.Channel)

				if info.Name == "torrent" && ev.Msg.Text == myName+" dropbox status" {
					status := GetDropboxStatus()
					slackRtm.PostMessage("torrent", status, postMessageParameters)
				}

				if info.Name == "torrent" && ev.Msg.Text == myName+" rtorrent list" {
					torrents := GetRtorrentList()
					buffer := bytes.NewBuffer(nil)
					table := tablewriter.NewWriter(buffer)
					table.SetHeader([]string{"Hash", "Name", "Percent done", "Ratio"})
					for _, v := range torrents {
						table.Append([]string{
							v.Hash,
							v.Name,
							fmt.Sprintf("%.0f%%", float64(v.BytesCompleted)/float64(v.BytesTotal)*float64(100)),
							fmt.Sprintf("%.2f", v.Ratio),
						})
					}
					table.Render()

					slackRtm.PostMessage("torrent", "```\n"+buffer.String()+"```", postMessageParameters)
				}

				if info.Name == "torrent" && ev.Msg.Text == myName+" dropbox start" {
					StartDropbox()
					slackRtm.PostMessage("torrent", "done", postMessageParameters)
				}

				if info.Name == "torrent" && ev.Msg.Text == myName+" dropbox stop" {
					StopDropbox()
					slackRtm.PostMessage("torrent", "done", postMessageParameters)
				}

			}
		}
	}
}
