// Copyright 2015 Nevio Vesic
// Please check out LICENSE file for more information about what you CAN and what you CANNOT do!
// Basically in short this is a free software for you to do whatever you want to do BUT copyright must be included!
// I didn't write all of this code so you could say it's yours.
// MIT License

package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	. "github.com/zyxar/eventsocket"
	"github.com/zyxar/grace/sigutil"
)

var (
	fshost   = flag.String("host", "localhost", "Freeswitch hostname")
	fsport   = flag.Uint("port", 8021, "Freeswitch port")
	password = flag.String("pass", "ClueCon", "Freeswitch password")
	timeout  = flag.Int64("timeout", 10, "Freeswitch conneciton timeout in seconds")
)

// Small client that will first make sure all events are returned as JSON and second, will originate
func main() {
	flag.Parse()
	client, err := NewClient(*fshost, *fsport, *password, time.Duration(*timeout)*time.Second)
	if err != nil {
		Error("Error while creating new client: %s", err)
		return
	}

	// Apparently all is good... Let us now handle connection :)
	// We don't want this to be inside of new connection as who knows where it my lead us.
	// Remember that this is crutial part in handling incoming messages :)
	go client.Handle()
	client.Send("events json ALL")
	client.BgApi(fmt.Sprintf("originate %s %s", "sofia/internal/1001@127.0.0.1", "&socket(192.168.1.2:8084 async full)"))

	go func() {
		for {
			msg, err := client.ReadMessage()
			if err != nil {
				// If it contains EOF, we really dont care...
				if !strings.Contains(err.Error(), "EOF") && err.Error() != "unexpected end of JSON input" {
					Error("Error while reading Freeswitch message: %s", err)
				}
				break
			}
			Debug("%s", msg)
		}
	}()
	sigutil.Trap(func(s sigutil.Signal) {
		Notice("closing upon receiving %s", s)
		if err := client.Close(); err != nil {
			Error("client close: %s", err)
		}
	}, sigutil.SIGINT, sigutil.SIGTERM)
}
