package main

import (
	"log"
)

type loadBalancer struct {
	readers map[int]reader
	status_chan chan StatusMsg
}

func NewLoadBalancer(channels []string, status_chan chan StatusMsg) loadBalancer {
	split := 100
	readers := make(map[int]reader)
	for i := 0; i < len(channels); i = i + split {
		channelBatch :=channels[i:i+split]
		readers[i] = NewReader(channelBatch, i/split, status_chan)
	}
	lb := loadBalancer {readers, status_chan}
	return lb
}

func Run(load_balancer loadBalancer) {
	for _, reader := range load_balancer.readers {
		go Read(reader)
	}

	for {
		status_msg := <-load_balancer.status_chan
		log.Println(status_msg.id, status_msg.total_load, status_msg.parted_channels)
		log.Println("-----------------------------------------------------------")
	}
}
