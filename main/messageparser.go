package main

import (
	"log"
	"strings"
)

type MessageParser struct {
	msg_chan chan string
}

func StartParser(msgParser MessageParser) {
	for msg := range msgParser.msg_chan {
		if !strings.Contains(msg, ":tmi.twitch.tv") && !strings.Contains(msg, ":justinfan696!justinfan696@justinfan696.tmi.twitch.tv") {
			i := strings.Index(msg, "#")
			channel := msg[i:]
			channel = channel[:strings.Index(channel, " ")]
			log.Println(channel)
		}
	}
}
