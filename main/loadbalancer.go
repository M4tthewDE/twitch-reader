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

func NewLoadBalancer(channels []string, status_chan chan StatusMsg) (loadBalancer){
	log.Println("new Loadbalancer")
	split := 50
	readers := make(map[int]reader)
	for i := 0; i < len(channels); i = i + split {
		channelBatch :=channels[i:i+split]
		readers[i] = NewReader(channelBatch, i/split, status_chan, make(chan string, 100))
	}
	cp := ChannelProvider {make(chan []string, 1)}
	lb := loadBalancer {readers, status_chan, cp}
	return lb
}

func Run(lb loadBalancer) {
	for _, reader := range lb.readers {
		go Read(reader)
	}

	go GetChannels(lb.channel_provider, 100)

	for {
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
	reader := NewReader(channels, len(lb.readers), lb.status_chan, make(chan string, 5))
	lb.readers[len(lb.readers)] = reader
	go Read(reader)
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
