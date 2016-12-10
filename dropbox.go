package main

import (
	"fmt"
	"net"
	"os/exec"
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
	lastStatus := ""

	for {
		<-timer.C
		status := GetDropboxStatus()

		if status != "Up to date" {
			checksSinceUpToDate++
			lastStatus = status
		} else {
			if checksSinceUpToDate > 3 {
				slackRtm.PostMessage("torrent", "Dropbox sync finished\nlast status was: "+lastStatus, postMessageParameters)
				checksSinceUpToDate = 0
			}
		}
	}
}

func StartDropbox() {
	cmd := exec.Command("/home/maerlyn/dropbox.py", "start")
	if err := cmd.Start(); err != nil {
		slackRtm.PostMessage("torrent", "start failed: "+err.Error(), postMessageParameters)
		return
	}

	err := cmd.Wait()
	if err != nil {
		slackRtm.PostMessage("torrent", "start failed: "+err.Error(), postMessageParameters)
	}
}

func StopDropbox() {
	cmd := exec.Command("/home/maerlyn/dropbox.py", "stop")
	if err := cmd.Start(); err != nil {
		slackRtm.PostMessage("torrent", "stop failed: "+err.Error(), postMessageParameters)
		return
	}

	err := cmd.Wait()
	if err != nil {
		slackRtm.PostMessage("torrent", "stop failed: "+err.Error(), postMessageParameters)
	}
}
