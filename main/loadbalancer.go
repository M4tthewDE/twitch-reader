package main

import (
	"errors"
	"log"
	"time"
)

type loadBalancer struct {
	readers           map[int]*reader
	channel_provider  ChannelProvider
	nextId            int
	reader_leave_chan chan StatusMsg
	msgparser         MessageParser
}

func NewLoadBalancer() *loadBalancer {
	readers := make(map[int]*reader)
	cp := ChannelProvider{make(chan []string, 1)}
	mp := MessageParser{make(chan string, 100)}
	lb := &loadBalancer{readers, cp, 0, make(chan StatusMsg, 10), mp}
	return lb
}

func Run(lb *loadBalancer) {
	go StartParser(lb.msgparser)
	var reader_to_merge *reader
	go GetChannels(lb.channel_provider, 100)
	start := time.Now()

	for {
		select {
		case new_channels := <-lb.channel_provider.channel_chan:
			distributeNewChannels(new_channels, lb)

		case status_msg := <-lb.reader_leave_chan:
			reader := lb.readers[status_msg.id]
			if !reader.deactivated {
				if len(status_msg.parted_channels) > 0 {
					for channel := range status_msg.parted_channels {
						distributeChannel(channel, lb)
					}
				}
			}
		default:
			total_load := 0
			total_channels := 0
			total_readers := 0
			for _, reader := range lb.readers {

				if !reader.deactivated {
					total_load = total_load + GetLoad(reader)
					total_channels = total_channels + len(reader.channels)
					total_readers++
				}

				if GetLoad(reader) < 30 && !reader.deactivated {
					if reader_to_merge != nil && reader_to_merge.id != reader.id {
						mergeReaders(reader, reader_to_merge, lb)
						reader_to_merge = nil
					} else {
						reader_to_merge = reader
					}
				}
			}
			if len(lb.readers) > 0 && time.Since(start).Seconds() >= 1 {
				log.Println("Average load:", total_load/total_readers, "readers:", total_readers, "total channels:", total_channels)
				start = time.Now()
			}
		}
	}
}

func mergeReaders(reader0 *reader, reader1 *reader, lb *loadBalancer) {
	reader1.join_chan <- reader0.channels
	reader0.deactivated = true
}

func distributeNewChannels(channels []string, lb *loadBalancer) {
	all_channels := getAllChannels(lb)
	var channels_to_remove []string

	for _, channel := range channels {
		if find(all_channels, channel) {
			channels_to_remove = append(channels_to_remove, channel)
		}
	}
	for _, channel := range channels_to_remove {
		channels = remove(channels, channel)
	}

	if len(channels) > 0 {
		split := 20
		for i := 0; i < len(channels); i = i + split {
			channelBatch := channels[i : i+split]
			reader := NewReader(channelBatch, lb.nextId, lb.reader_leave_chan, lb.msgparser.msg_chan)
			lb.nextId++
			lb.readers[len(lb.readers)] = reader
			go Read(reader)
		}
	}
}

func getAllChannels(lb *loadBalancer) []string {
	var channels []string
	for _, reader := range lb.readers {
		channels = append(channels, GetReaderChannels(reader)...)
	}
	return channels
}

func distributeChannel(channel string, lb *loadBalancer) {
	r, err := getAvailableReader(lb)
	if err != nil {
		reader := NewReader([]string{channel}, lb.nextId, lb.reader_leave_chan, lb.msgparser.msg_chan)
		lb.nextId++
		lb.readers[len(lb.readers)] = reader
		go Read(reader)
	} else {
		m := make(map[string]int)
		m[channel] = 0
		r.join_chan <- m
	}
}

func getAvailableReader(lb *loadBalancer) (reader *reader, err error) {
	for _, reader := range lb.readers {
		if GetLoad(reader) < 70 {
			return reader, nil
		}
	}
	return nil, errors.New("No reader found")
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func remove(s []string, element string) []string {
	i := indexOf(s, element)
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func indexOf(strings []string, element string) int {
	i := 0
	for _, s := range strings {
		if s == element {
			return i
		}
		i++
	}
	return -1
}
