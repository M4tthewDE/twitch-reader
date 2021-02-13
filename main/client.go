package main

import (
	"net/http"
	"io/ioutil"
	"os"
	"github.com/buger/jsonparser"
	"log"
)


func main() {
	channels := getChannels(100)
	log.Println("Total channels:", len(channels))

	var readers []reader
	for  i := 0; i < len(channels); i = i + 20 {
		channelBatch := channels[i:i+20]
		readers = append(readers, NewReader(channelBatch, i/20))
	}

	for _, reader := range readers {
		go Read(reader)
	}

	for {
	}
}

func getChannels(n int) ([]string){
	var channels []string
	var cursor []byte
	for i := 0; i < n; i++ {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?first=100&after=" + string(cursor), nil)
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
	}

	return channels
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
