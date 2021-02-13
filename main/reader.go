package main

import (
	"net"
	"log"
	"fmt"
	"net/textproto"
	"bufio"
	"time"
	"strings"
)

// 150 msg/s (1min period) might be upper limit
// TODO has to be tested for longe duration (1h)

type reader struct {
	channels []string
	id int
	load int
	conn net.Conn
	c chan int
}

func NewReader(channels []string, id int) reader {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		log.Println("[ ", id, " ]", " Error Connecting!", err)
		NewReader(channels, id)
	}

	fmt.Fprintf(conn, "PASS oauth: " + "\n")
	fmt.Fprintf(conn, "NICk justinfan696" + "\n")

	for _, channel := range channels {
		fmt.Fprintf(conn, "JOIN " + channel + "\n")
		log.Println("[", id, "] " + "Joined " + channel)
	}
	r := reader {channels, id, 0, conn, make(chan int)}
	return r
}

func Read(r reader) {
	tp := textproto.NewReader(bufio.NewReader(r.conn))

	startTime := time.Now()
	n := 0
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Println("[ ", r.id, " ]", " Error Reading! Reconnecting...")
			new_r := NewReader(r.channels, r.id)
			Read(new_r)
		}

		if strings.Contains(line, "PING") {
			fmt.Fprintf(r.conn, "PONG :tmi.twitch.tv" + "\n")
		}

		n++
		if time.Since(startTime).Seconds() >= 60 {
			r.load = n
			log.Println("[", r.id, "]", r.load/60, "msg/s")
			startTime = time.Now()
			n = 0
		}
	}
}

func GetLoad(r reader) int {
	return r.load
}
