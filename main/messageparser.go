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
		parseMessage(msg)
	}
}

func parseMessage(raw string) {
	if !strings.HasPrefix(raw, ":justinfan696") && !strings.HasPrefix(raw, ":tmi.twitch.tv") {
		user_msg := parseUserMessage(raw)

		if user_msg.user == "matthewde" {
			log.Println(user_msg)
		}
	}
}

func parseUserMessage(raw string) UserMessage {
	raw_parts := strings.Split(raw, " ")
	user := raw_parts[0][1:strings.Index(raw_parts[0], "!")]
	channel := raw_parts[2]
	msg := raw[strings.Index(raw, channel)+len(channel)+2:]

	return UserMessage{channel, user, msg}
}
