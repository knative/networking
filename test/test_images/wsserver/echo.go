/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"knative.dev/networking/pkg/http/header"
	"knative.dev/networking/pkg/http/probe"
	"knative.dev/networking/test"
)

const suffixMessageEnv = "SUFFIX"

// Gets the message suffix from envvar. Empty by default.
func messageSuffix() string {
	value := os.Getenv(suffixMessageEnv)
	if value == "" {
		return ""
	}
	return value
}

var upgrader = ws.HTTPUpgrader{
	Negotiate: func(opt httphead.Option) (ret httphead.Option, err error) {
		return httphead.Option{}, nil
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	if header.IsKubeletProbe(r) {
		w.WriteHeader(http.StatusOK)
		return
	}
	conn, _, _, err := upgrader.Upgrade(r, w)
	if err != nil {
		log.Println("Error upgrading websocket:", err)
		return
	}
	defer conn.Close()
	log.Println("Connection upgraded to WebSocket. Entering receive loop.")
	for {
		var messages []wsutil.Message
		messages, err = wsutil.ReadMessage(conn, ws.StateServerSide, messages)
		for _, m := range messages {
			message := m.Payload
			messageType := m.OpCode
			if err != nil {
				// We close abnormally, because we're just closing the connection in the client,
				// which is okay. There's no value delaying closure of the connection unnecessarily.
				if errors.Is(err, io.ErrUnexpectedEOF) {
					log.Println("Client disconnected.")
				} else {
					log.Println("Handler exiting on error:", err)
				}
				return
			}
			if suffix := messageSuffix(); suffix != "" {
				respMes := string(message) + " " + suffix
				message = []byte(respMes)
			}

			log.Printf("Successfully received: %q", message)
			if err = wsutil.WriteClientMessage(conn, messageType, message); err != nil {
				log.Println("Failed to write message:", err)
				return
			}
			log.Printf("Successfully wrote: %q", message)
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	h := probe.NewHandler(http.HandlerFunc(handler))
	port := os.Getenv("PORT")
	if cert, key := os.Getenv("CERT"), os.Getenv("KEY"); cert != "" && key != "" {
		log.Print("Server starting on port with TLS ", port)
		test.ListenAndServeTLSGracefully(cert, key, ":"+port, h.ServeHTTP)
	} else {
		log.Print("Server starting on port ", port)
		test.ListenAndServeGracefully(":"+port, h.ServeHTTP)
	}
}
