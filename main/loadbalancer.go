package main

import (
	"errors"
	"log"
)

type loadBalancer struct {
	readers          map[int]reader
	status_chan      chan StatusMsg
	channel_provider ChannelProvider
	nextId           int
}

func NewLoadBalancer(status_chan chan StatusMsg) *loadBalancer {
	log.Println("new Loadbalancer")
	readers := make(map[int]reader)
	cp := ChannelProvider{make(chan []string, 1)}
	lb := &loadBalancer{readers, status_chan, cp, 0}
	return lb
}

func Run(lb *loadBalancer) {
	for _, reader := range lb.readers {
		go Read(reader)
	}

	go GetChannels(lb.channel_provider, 1)

	var reader_to_merge *reader
	for {
		//TODO automatic merging of low load runners

		select {
		case status_msg := <-lb.status_chan:
			reader := status_msg.r
			lb.readers[reader.id] = reader
			log.Println(reader.id, GetLoad(reader), len(reader.channels))
			if len(status_msg.parted_channels) > 0 {
				for channel := range status_msg.parted_channels {
					distributeChannel(channel, lb)
				}
			}
			if GetLoad(reader) < 30 {
				if reader_to_merge != nil && reader_to_merge.id != reader.id {
					// TODO fix racecondition 🚗
					mergeReaders(reader, reader_to_merge, lb)
					reader_to_merge = nil
				} else {
					reader_to_merge = &reader
				}
			}

		case new_channels := <-lb.channel_provider.channel_chan:
			distributeNewChannels(new_channels, lb)
		}
	}
}

func mergeReaders(reader0 reader, reader1 *reader, lb *loadBalancer) {
	log.Println("merging ", reader0.id, "->", reader1.id)

	for channel := range reader0.channels {
		reader1.channel_chan <- channel
	}
	close(reader0.channel_chan)
	delete(lb.readers, reader0.id)
	log.Println(reader0.id, "deactivated")
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
		split := 10
		for i := 0; i < len(channels); i = i + split {
			channelBatch := channels[i : i+split]
			reader := NewReader(channelBatch, lb.nextId, lb.status_chan, make(chan string, 20))
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
		log.Println("we need more channels here")
		reader := NewReader([]string{channel}, lb.nextId, lb.status_chan, make(chan string, 20))
		lb.nextId++
		lb.readers[len(lb.readers)] = reader
		go Read(reader)
	} else {
		r.channel_chan <- channel
	}
}

func getAvailableReader(lb *loadBalancer) (reader *reader, err error) {
	for _, reader := range lb.readers {
		if GetLoad(reader) < 70 {
			return &reader, nil
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
