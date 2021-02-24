package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"
)

type reader struct {
	id          int
	channels    map[string]int
	load        []int
	conn        net.Conn
	join_chan   chan map[string]int
	leave_chan  chan StatusMsg
	deactivated bool
}

type StatusMsg struct {
	id              int
	parted_channels map[string]int
}

func NewReader(twitch_channels []string, id int, leave_chan chan StatusMsg) *reader {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		NewReader(twitch_channels, id, leave_chan)
	}

	fmt.Fprintf(conn, "PASS oauth: "+"\n")
	fmt.Fprintf(conn, "NICk justinfan696"+"\n")

	channels := make(map[string]int)
	for _, channel := range twitch_channels {
		fmt.Fprintf(conn, "JOIN "+channel+"\n")

		channels[channel] = 0
	}
	load := []int{50, 50, 50, 50}
	r := &reader{id, channels, load, conn, make(chan map[string]int, 20), leave_chan, false}
	return r
}

func Read(r *reader) {
	tp := textproto.NewReader(bufio.NewReader(r.conn))

	startTime := time.Now()
	n := 0
	channel_load_tmp := make(map[string]int)

	for {
		if r.deactivated {
			return
		}
		select {
		case new_channel := <-r.join_chan:
			for channel := range new_channel {
				joinChannel(*r, channel)
			}
		default:
		}

		line, err := tp.ReadLine()
		if err != nil {
			channels := make([]string, 0, len(r.channels))
			for c := range r.channels {
				channels = append(channels, c)
			}
			new_r := NewReader(channels, r.id, r.leave_chan)
			Read(new_r)
		}

		if strings.Contains(line, "PING") {
			fmt.Fprintf(r.conn, "PONG :tmi.twitch.tv"+"\n")
		}

		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			if parts[1] == "PRIVMSG" {
				channel_load_tmp[parts[2]]++
				n++
			}
		}

		if time.Since(startTime).Seconds() >= 1 {
			r.load = append(r.load, n)
			r.load = r.load[1:]
			startTime = time.Now()
			n = 0

			for channel := range r.channels {
				r.channels[channel] = channel_load_tmp[channel]
				channel_load_tmp[channel] = 0
			}

			if GetLoad(r) > 100 {
				for channel := range downscale(r) {
					fmt.Fprintf(r.conn, "PART "+channel+"\n")
					delete(r.channels, channel)
				}
			}
		}
	}
}

func downscale(r *reader) map[string]int {
	removed_channels := make(map[string]int)
	tmp := GetLoad(r)

	for channel := range r.channels {
		tmp = tmp - r.channels[channel]
		removed_channels[channel] = r.channels[channel]
		if tmp < 200 {
			r.leave_chan <- StatusMsg{r.id, removed_channels}
			return removed_channels
		}
	}
	return nil
}

func GetLoad(r *reader) int {
	load := 0
	for _, i := range r.load {
		load = load + i
	}
	return load / 4
}

func joinChannel(r reader, channel string) {
	fmt.Fprintf(r.conn, "JOIN "+channel+"\n")
	r.channels[channel] = 0
}

func GetReaderChannels(r *reader) []string {
	var channels []string
	for c := range r.channels {
		channels = append(channels, c)
	}
	return channels
}
