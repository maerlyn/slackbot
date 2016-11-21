package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func GetDropboxStatus() string {
	c, err := net.Dial("unix", dropboxCommandSocket)
	if err != nil {
		return fmt.Sprintf("net.Dial error: %+v", err)
	}
	defer c.Close()

	_, err = c.Write([]byte("get_dropbox_status\ndone\n"))
	if err != nil {
		return fmt.Sprintf("get_dropbox_status write error: %+v", err)
	}

	done := false
	retval := ""

	for !done {
		buf := make([]byte, 10240)
		n, err := c.Read(buf[:])
		if err != nil {
			return fmt.Sprintf("read error: %+v", err)
		}

		response := strings.Split(strings.Replace(strings.Replace(string(buf[0:n]), "status\t", "", -1), "\t", "\n", -1), "\n")
		for _, s := range response {
			if s == "ok" {
				continue
			}
			if s == "done" {
				done = true
				break
			}

			retval += s + "\n"
		}
	}

	return strings.Trim(retval, "\n")
}

func MonitorDropboxStatus() {
	timer := time.NewTicker(1 * time.Minute)
	checksSinceUpToDate := 0

	for {
		<-timer.C
		status := GetDropboxStatus()

		if status != "Up to date" {
			checksSinceUpToDate++
		} else {
			if checksSinceUpToDate > 2 {
				slackRtm.PostMessage("torrent", "Dropbox sync finished", postMessageParameters)
				checksSinceUpToDate = 0
			}
		}
	}
}
