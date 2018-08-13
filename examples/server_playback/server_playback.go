// Copyright 2015 Nevio Vesic
// Please check out LICENSE file for more information about what you CAN and what you CANNOT do!
// Basically in short this is a free software for you to do whatever you want to do BUT copyright must be included!
// I didn't write all of this code so you could say it's yours.
// MIT License

package main

import (
	"flag"
	"os"
	"strings"

	. "github.com/zyxar/eventsocket"
	"github.com/zyxar/grace/sigutil"
)

var welcomeFile = flag.String("wav", "../media/welcome.wav", "set welcome wav file")

func main() {
	flag.Parse()
	if *welcomeFile == "" {
		Error("welcome wav file not fould")
		os.Exit(1)
	}
	{
		file, err := os.Open(*welcomeFile)
		if err != nil {
			Error(err.Error())
			os.Exit(1)
		}
		file.Close()
	}

	server, err := NewServer("localhost:8084")
	if err != nil {
		Error("Got error while starting Freeswitch outbound server: %s", err)
		os.Exit(1)
	}

	Notice("starting server: %v", server.Addr())
	server.Start(handle)
	sigutil.Trap(func(s sigutil.Signal) {
		Notice("shutting down server from receiving %s", s)
		if err := server.Close(); err != nil {
			Error("server close: %s", err)
		}
		server.Wait()
	}, sigutil.SIGINT, sigutil.SIGTERM)
}

// handle - Running under goroutine here to explain how to handle playback ( play to the caller )
func handle(conn *SocketConnection) {
	Notice("New incomming connection: %v", conn)
	if err := conn.Connect(); err != nil {
		Error("Got error while accepting connection: %s", err)
		return
	}
	answer, err := conn.ExecuteAnswer("", false)
	if err != nil {
		Error("Got error while executing answer: %s", err)
		return
	}

	Debug("Answer Message: %s", answer)
	Debug("Caller UUID: %s", answer.GetHeader("Caller-Unique-Id"))

	cUUID := answer.GetCallUUID()
	if sm, err := conn.Execute("playback", *welcomeFile, true); err != nil {
		Error("Got error while executing playback: %s", err)
		return
	} else {
		Debug("Playback Message: %s", sm)
	}

	if hm, err := conn.ExecuteHangup(cUUID, "", false); err != nil {
		Error("Got error while executing hangup: %s", err)
		return
	} else {
		Debug("Hangup Message: %s", hm)
	}

	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			// If it contains EOF, we really dont care...
			if !strings.Contains(err.Error(), "EOF") {
				Error("Error while reading Freeswitch message: %s", err)
			}
			break
		}
		Debug("%s", msg)
	}
}
