package main

import (
	"log"
)

type loadBalancer struct {
	readers map[int]reader
	status_chan chan StatusMsg
	channel_chan chan []string
	channel_provider ChannelProvider
}

func NewLoadBalancer(channels []string, status_chan chan StatusMsg) (loadBalancer){
	log.Println("new Loadbalancer")
	split := 60
	readers := make(map[int]reader)
	for i := 0; i < len(channels); i = i + split {
		channelBatch :=channels[i:i+split]
		readers[i] = NewReader(channelBatch, i/split, status_chan)
	}
	channel_chan := make(chan []string)
	channel_provider := ChannelProvider{channel_chan}
	lb := loadBalancer {readers, status_chan, channel_chan, channel_provider}
	return lb
}

func Run(load_balancer loadBalancer) {
	for _, reader := range load_balancer.readers {
		go Read(reader)
	}
	go GetChannels(load_balancer.channel_provider, 10)

	for {
		status_msg := <-load_balancer.status_chan
		channels := <-load_balancer.channel_chan
		log.Println(status_msg.id, status_msg.total_load, status_msg.parted_channels)
		log.Println(channels)
		log.Println("-----------------------------------------------------------")
	}
}
