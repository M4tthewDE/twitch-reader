package main

import (
	"github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type ChannelProvider struct {
	channel_chan chan []string
}

func GetChannels(channel_provider ChannelProvider, n int) []string {
	time.Sleep(3 * time.Second)
	var channels []string
	var cursor []byte

	var i int
	for {
		for i = 0; i < n; i++ {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?first=100&after="+string(cursor), nil)
			req.Header.Add("Client-Id", os.Getenv("TWITCH_ID"))
			req.Header.Add("Authorization", "Bearer "+os.Getenv("TWITCH_TOKEN"))
			resp, _ := client.Do(req)

			if resp.StatusCode == 200 {
				data, _ := ioutil.ReadAll(resp.Body)

				_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					tmp, _, _, _ := jsonparser.Get(value, "user_login")
					channels = append(channels, "#"+string(tmp))
				}, "data")

				if err != nil {
					log.Fatal(err)
				}

				cursor, _, _, _ = jsonparser.Get(data, "pagination", "cursor")
				channel_provider.channel_chan <- channels
				channels = nil
				time.Sleep(3 * time.Second)
			}
		}
		cursor = cursor[:0]
	}
}
