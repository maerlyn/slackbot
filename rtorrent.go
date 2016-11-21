package main

import (
	"github.com/tehjojo/go-rtorrent/xmlrpc"
)

type Torrent struct {
	Hash           string
	Name           string
	IsComplete     bool
	BytesCompleted int
	BytesTotal     int
	Ratio          float64
}

func GetRtorrentList() []Torrent {
	var torrents []Torrent

	c := xmlrpc.NewClient(rtorrentAddress, true)
	args := []interface{}{"main", "d.get_hash=", "d.get_name=", "d.get_complete=", "d.get_completed_bytes=", "d.get_size_bytes=", "d.get_ratio="}
	results, err := c.Call("d.multicall", args...)

	if err != nil {
		panic(err)
	}

	for _, outerResult := range results.([]interface{}) {
		for _, innerResult := range outerResult.([]interface{}) {
			torrentData := innerResult.([]interface{})
			torrents = append(torrents, Torrent{
				Hash:           torrentData[0].(string),
				Name:           torrentData[1].(string),
				IsComplete:     torrentData[2].(int) > 0,
				BytesCompleted: torrentData[3].(int),
				BytesTotal:     torrentData[4].(int),
				Ratio:          float64(torrentData[5].(int)) / float64(1000),
			})
		}
	}

	return torrents
}
