package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"gochat/api"
)

const helpStr = `Commands
1. /users :- Get list of live users
2. @{user} message :- send message to specified user
3. /exit :- Exit the Chat
4. /all :- Send message to all the users [TODO]`

var (
	name      = flag.String("name", "", "The name you want to chat as")
	port      = flag.Int("port", 12345, "Port that your server will run on.")
	host      = flag.String("host", "", "Host IP that your server is running on.")
	stdReader = bufio.NewReader(os.Stdin)
)

var MyHandle api.Handle
var USERS = PeerHandleMapSync{
	PeerHandleMap: make(map[string]api.Handle),
}

func main() {
	// Parse flags for host, port and name
	flag.Parse()

	if *name == "" || *host == "" {
		fmt.Println("Usage: gochat --name <name> --host <IP Address> --port <port>")
		os.Exit(1)
	}

	MyHandle = api.Handle{
		Name: *name,
		Host: *host,
		Port: int32(*port),
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// Broadcast for is-alive on 33333 with own UserHandle.
	go broadcastOwnHandle(&wg)

	// Listener for is-alive broadcasts from other hosts. Listening on 33333
	go listenAndRegisterUsers(&wg)

	// gRPC listener
	go startServer(&wg)

	for {
		fmt.Printf("> ")
		textInput, _ := stdReader.ReadString('\n')
		// convert CRLF to LF
		textInput = strings.Replace(textInput, "\n", "", -1)
		parseAndExecInput(textInput)
	}

	wg.Wait()
}

// Handle the input chat messages as well as help commands
func parseAndExecInput(input string) {
	// Split the line into 2 tokens (cmd and message)
	tokens := strings.SplitN(input, " ", 2)
	cmd := tokens[0]

	switch {
	case cmd == "":
		break
	case cmd == "?":
		fmt.Printf(helpStr)
		break
	case strings.ToLower(cmd) == "/users":
		fmt.Println(USERS)
		break
	case strings.ToLower(cmd) == "/exit":
		os.Exit(1)
		break
	case cmd[0] == '@':
		message := "hi" //default
		if len(tokens) > 1 {
			message = tokens[1]
		}
		userName := cmd[1:]
		// send message to particular user
		if h, ok := USERS.Get(userName); ok {
			sendChat(h, message)
		} else {
			fmt.Println("No user: ", cmd)
		}

		break
	case strings.ToLower(cmd) == "/help":
		fmt.Println(helpStr)
	default:
		fmt.Println(helpStr)
	}
}
