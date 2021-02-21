package main

import (
	"log"
	"errors"
)

type loadBalancer struct {
	readers map[int]reader
	status_chan chan StatusMsg
	channel_provider ChannelProvider
}

func NewLoadBalancer(status_chan chan StatusMsg) (loadBalancer){
	log.Println("new Loadbalancer")
	readers := make(map[int]reader)
	cp := ChannelProvider {make(chan []string, 1)}
	lb := loadBalancer {readers, status_chan, cp}
	return lb
}

func Run(lb loadBalancer) {
	for _, reader := range lb.readers {
		go Read(reader)
	}

	go GetChannels(lb.channel_provider, 10)

	for {
		//TODO automatic merging of low load runners
		select {
			case status_msg := <-lb.status_chan:
				lb.readers[status_msg.r.id] = status_msg.r
				log.Println(status_msg.r.id, GetLoad(status_msg.r), len(status_msg.r.channels))
				if len(status_msg.parted_channels) > 0 {
					for channel := range status_msg.parted_channels {
						distributeChannel(channel, lb)
					}
				}
			case new_channels := <-lb.channel_provider.channel_chan:
				distributeNewChannels(new_channels, lb)
			default:
		}
	}
}

func isChannelRead(channel string, lb loadBalancer) bool {
	for _, reader := range lb.readers {
		_, exists := reader.channels[channel]
		if exists {
			return true
		}
	}
	return false
}

func distributeNewChannels(channels []string, lb loadBalancer) {
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
			channelBatch :=channels[i:i+split]
			reader := NewReader(channelBatch, len(lb.readers), lb.status_chan, make(chan string, 5))
			lb.readers[len(lb.readers)] = reader
			go Read(reader)
		}
	}
}

func getAllChannels(lb loadBalancer) []string {
	var channels []string
	for _, reader := range lb.readers {
		channels = append(channels, GetReaderChannels(reader)...)
	}
	return channels
}

func distributeChannel(channel string, lb loadBalancer) {
	r, err := getAvailableReader(lb)
	if err != nil {
		reader := NewReader([]string{channel}, len(lb.readers), lb.status_chan, make(chan string, 5))
		lb.readers[len(lb.readers)] = reader
		go Read(reader)
	} else {
		r.channel_chan <-channel
	}
}


func getAvailableReader(lb loadBalancer) (reader *reader, err error) {
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
