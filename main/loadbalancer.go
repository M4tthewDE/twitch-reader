package main

import (
	"log"
	"errors"
)

type loadBalancer struct {
	readers map[int]reader
	status_chan chan StatusMsg
}

func NewLoadBalancer(channels []string, status_chan chan StatusMsg) (loadBalancer){
	log.Println("new Loadbalancer")
	split := 100
	readers := make(map[int]reader)
	for i := 0; i < len(channels); i = i + split {
		channelBatch :=channels[i:i+split]
		readers[i] = NewReader(channelBatch, i/split, status_chan, make(chan string, 5))
	}
	lb := loadBalancer {readers, status_chan}
	return lb
}

func Run(lb loadBalancer) {
	for _, reader := range lb.readers {
		go Read(reader)
	}

	for {
		status_msg := <-lb.status_chan
		lb.readers[status_msg.r.id] = status_msg.r
		log.Println(status_msg.r.id, GetLoad(status_msg.r), len(status_msg.r.channels))
		if len(status_msg.parted_channels) > 0 {
			log.Println("We should find readers for these channels")

			for channel := range status_msg.parted_channels {
				r, err := getAvailableReader(lb)
				if err != nil {
					log.Println("We need another reader!")
					r := NewReader([]string{channel}, len(lb.readers), lb.status_chan, make(chan string, 5))
					lb.readers[len(lb.readers)] = r
					go Read(r)
				}
				select {
					case r.channel_chan <-channel:
						log.Println("Put in ", channel)
					default:
						log.Println("Couldnt put in",  channel)
				}
			}
		}
	}
}

func getAvailableReader(lb loadBalancer) (reader *reader, err error) {
	for _, reader := range lb.readers {
		if GetLoad(reader) < 100 {
			return &reader, nil
		}
	}
	return nil, errors.New("No reader found")
}
