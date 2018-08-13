// Copyright 2015 Nevio Vesic
// Please check out LICENSE file for more information about what you CAN and what you CANNOT do!
// Basically in short this is a free software for you to do whatever you want to do BUT copyright must be included!
// I didn't write all of this code so you could say it's yours.
// MIT License

package main

import (
	"os"
	"strings"

	. "github.com/zyxar/eventsocket"
	"github.com/zyxar/grace/sigutil"
)

var (
	goeslMessage = "Hello from GoESL. Open source freeswitch event socket wrapper written in Golang!"
)

func main() {
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

// handle handles connection of tts outbound server
func handle(conn *SocketConnection) {
	Debug("New incomming connection: %v", conn.RemoteAddr())
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

	if te, err := conn.ExecuteSet("tts_engine", "flite", false); err != nil {
		Error("Got error while attempting to set tts_engine: %s", err)
	} else {
		Debug("TTS Engine Msg: %s", te)
	}

	if tv, err := conn.ExecuteSet("tts_voice", "slt", false); err != nil {
		Error("Got error while attempting to set tts_voice: %s", err)
	} else {
		Debug("TTS Voice Msg: %s", tv)
	}

	if sm, err := conn.Execute("speak", goeslMessage, true); err != nil {
		Error("Got error while executing speak: %s", err)
		return
	} else {
		Debug("Speak Message: %s", sm)
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
