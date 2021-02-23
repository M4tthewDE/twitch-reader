package main

import ()

func main() {
	lb := NewLoadBalancer()
	Run(lb)
	for {
		select {}
	}
}
