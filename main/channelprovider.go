package main

import (
	"net/http"
	"os"
	"io/ioutil"
	"github.com/buger/jsonparser"
	"log"
)

type ChannelProvider struct {
	channel_chan chan []string
}

func GetChannels(channel_provider ChannelProvider, n int) ([]string) {
	log.Println("start GetChannels")
	var channels []string
	var cursor []byte

	var i int
	for {
		for i = 0; i < n; i++ {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?first=60&after=" + string(cursor), nil)
			req.Header.Add("Client-Id", os.Getenv("TWITCH_ID"))
			req.Header.Add("Authorization", "Bearer " + os.Getenv("TWITCH_TOKEN"))
			resp, _ := client.Do(req)

			defer resp.Body.Close()
			data, _ := ioutil.ReadAll(resp.Body)

			jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				tmp, _, _, _ := jsonparser.Get(value, "user_login")
				channels = append(channels, "#" + string(tmp))
			}, "data")

			cursor, _, _, _ = jsonparser.Get(data, "pagination", "cursor")
			channel_provider.channel_chan <- channels
			channels = nil
		}
		i = 0
	}
}