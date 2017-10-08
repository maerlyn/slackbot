package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type TunnelsResponse struct {
	Tunnels []struct {
		Name   string `json:"name"`
		Config struct {
			Addr string `json:"addr"`
		} `json:"config"`
		PublicUrl string `json:"public_url"`
	} `json:"tunnels"`
}

type NewTunnel struct {
	Name  string `json:"name"`
	Proto string `json:"proto"`
	Addr  string `json:"addr"`
}

func NgrokStatus() {
	status := ""
	if !isNgrokRunning() {
		status = "not "
	}

	slackRtm.PostMessage("torrent", "ngrok is "+status+"running", postMessageParameters)
}

func NgrokStart() {
	if isNgrokRunning() {
		slackRtm.PostMessage("torrent", "ngrok is already running", postMessageParameters)
		return
	}

	go startNgrok()
}

func NgrokStop() {
	if !isNgrokRunning() {
		slackRtm.PostMessage("torrent", "ngrok is not running", postMessageParameters)
		return
	}

	cmd := exec.Command("pgrep", "ngrok")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))

	pid, _ := strconv.Atoi(strings.Trim(string(out), "\n"))

	ngrokProcess, _ := os.FindProcess(pid)
	ngrokProcess.Kill()
}

func NgrokList() {
	if !isNgrokRunning() {
		fmt.Println("not running")
		return
	}

	tunnels := getNgrokTunnels()

	if len(tunnels.Tunnels) == 0 {
		slackRtm.PostMessage("torrent", "no tunnels", postMessageParameters)
		return
	}

	for _, tunnel := range tunnels.Tunnels {
		slackRtm.PostMessage("torrent", fmt.Sprintf("%s via %s\n", tunnel.Config.Addr, tunnel.PublicUrl), postMessageParameters)
	}
}

func NgrokClear() {
	tunnels := getNgrokTunnels()
	client := &http.Client{}

	for _, tunnel := range tunnels.Tunnels {
		request, _ := http.NewRequest("DELETE", "http://localhost:4040/api/tunnels/"+tunnel.Name, nil)
		resp, err := client.Do(request)
		resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}

	slackRtm.PostMessage("torrent", "cleared", postMessageParameters)
}

func NgrokTunnel(message string) {
	tunnel := NewTunnel{
		Addr:  message,
		Name:  "slacktunnel",
		Proto: "tcp",
	}

	reqBody, _ := json.Marshal(tunnel)

	client := &http.Client{}
	request, _ := http.NewRequest("POST", "http://localhost:4040/api/tunnels", bytes.NewBuffer(reqBody))
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	resp.Body.Close()

	if err != nil {
		slackRtm.PostMessage("torrent", "error: "+err.Error(), postMessageParameters)
	}

	time.Sleep(2 * time.Second)
	NgrokList()
}

func isNgrokRunning() bool {
	resp, err := http.Get("http://127.0.0.1:4040/")

	if err != nil {
		return false
	}

	resp.Body.Close()

	return true
}

func startNgrok() {
	slackRtm.PostMessage("torrent", "starting ngrok", postMessageParameters)

	cmd := exec.Command(ngrokBinary, "start", "--none")

	cmd.CombinedOutput()

	return
}

func getNgrokTunnels() TunnelsResponse {
	resp, _ := http.Get("http://localhost:4040/api/tunnels")
	var tunnels TunnelsResponse
	responseBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(responseBody, &tunnels)

	return tunnels
}
