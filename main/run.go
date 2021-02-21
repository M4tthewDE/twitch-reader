package main

import (
)

func main() {
	status_chan := make(chan StatusMsg)

	lb := NewLoadBalancer(status_chan)
	Run(lb)
	for{select{}}
}
