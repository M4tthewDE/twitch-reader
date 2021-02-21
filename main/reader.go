package main

import (
	"net"
	"log"
	"fmt"
	"net/textproto"
	"bufio"
	"time"
	"strings"
)

type StatusMsg struct {
	r reader
	parted_channels map[string]int
}

type reader struct {
	id int
	channels map[string]int
	load []int
	conn net.Conn
	status_chan chan StatusMsg
	channel_chan chan string
}

func NewReader(twitch_channels []string, id int, status_chan chan StatusMsg, channel_chan chan string) reader {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		log.Println("[ ", id, " ]", " Error Connecting!", err)
		NewReader(twitch_channels, id, status_chan, channel_chan)
	}

	fmt.Fprintf(conn, "PASS oauth: " + "\n")
	fmt.Fprintf(conn, "NICk justinfan696" + "\n")

	channels := make(map[string]int)
	for _, channel := range twitch_channels {
		fmt.Fprintf(conn, "JOIN " + channel + "\n")
		log.Println("[", id, "] " + "Joined " + channel)

		channels[channel] = 0
	}
	load := []int{0,0,0,0}
	r := reader {id, channels, load, conn, status_chan, channel_chan}
	return r
}

func Read(r reader) {
	tp := textproto.NewReader(bufio.NewReader(r.conn))

	startTime := time.Now()
	n := 0
	channel_load_tmp := make(map[string]int)

	for {
		select {
			case new_channel := <-r.channel_chan:
				joinChannel(r, new_channel)
			default:
		}

		line, err := tp.ReadLine()
		if err != nil {
			channels := make([]string, 0, len(r.channels))
			for c, _ := range r.channels {
				channels = append(channels, c)
			}
			log.Println("[ ", r.id, " ]", " Error Reading! Reconnecting...")
			new_r := NewReader(channels, r.id, r.status_chan, r.channel_chan)
			Read(new_r)
		}

		if strings.Contains(line, "PING") {
			fmt.Fprintf(r.conn, "PONG :tmi.twitch.tv" + "\n")
		}

		parts := strings.Split(line, " ")
		if parts[1] == "PRIVMSG" {
			channel_load_tmp[parts[2]]++
			n++
		}

		if time.Since(startTime).Seconds() >= 1 {
			r.load = append(r.load, n)
			r.load = r.load[1:]
			startTime = time.Now()
			n = 0

			for channel, _ := range r.channels {
				r.channels[channel] = channel_load_tmp[channel]
				channel_load_tmp[channel] = 0
			}
			r.status_chan <- StatusMsg {r, nil}

			if GetLoad(r) > 100 {
				for channel, _ := range downscale(r) {
					fmt.Fprintf(r.conn, "PART " + channel + "\n")
					delete(r.channels, channel)
				}
			}
		}
	}
}

func downscale(r reader) map[string]int {
	removed_channels :=  make(map[string]int)
	tmp := GetLoad(r)

	for channel, _ := range r.channels {
		tmp = tmp - r.channels[channel]
		removed_channels[channel] = r.channels[channel]
		if tmp < 200 {
			r.status_chan <- StatusMsg {r, removed_channels}
			return removed_channels
		}
	}
	return nil
}

func GetLoad(r reader) int {
	load := 0
	for _, i := range r.load {
		load = load + i
	}
	return load/4
}

func joinChannel(r reader, channel string) {
	fmt.Fprintf(r.conn, "JOIN " + channel + "\n")
	r.channels[channel] = 0
}
