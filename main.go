package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	pb "github.com/gautamrege/gochat/api"
)

var (
	name = flag.String("name", "", "The name you want to chat as")
	port = flag.Int("port", 12345, "Port that your server will run on.")
	host = flag.String("host", "", "Host IP that your server is running on.")
)

func main() {
	// Parse flags for host, port and name
	flag.Parse()

	if *name == "" || *host == "" {
		fmt.Println("Usage: gochat --name <name> --host <IP Address> --port <port>")
		os.Exit(1)
	}
	// Create your own Global Handle ME
	var wg sync.WaitGroup
	wg.Add(4)

	HANDLES.HandleMap = make(map[string]Handle)

	// exit channel is a buffered channel for 5 exit patterns
	exit := make(chan bool, 5)

	// Listener for is-alive broadcasts from other hosts. Listening on 33333
	go registerHandle(&wg, exit)

	// Broadcast for is-alive on 33333 with own Handle.
	go isAlive(&wg, exit)

	// Cleanup Dead Handles
	go cleanupDeadHandles(&wg, exit)

	// gRPC listener
	go listen(&wg, exit)

	ME = pb.Handle{
		Name: *name,
		Host: *host,
		Port: int32(*port),
	}

	var input string
	for {
		// Accept chat input
		fmt.Printf("> ")
		input = readInput()

		parseAndExecInput(input)
		// Loop indefinitely and render Term
		// When we need to exit, send true 3 times on exit channel!
	}

	// exit cleanly on waitgroup
	wg.Wait()
	close(exit)
}

// Handle the input chat messages as well as help commands
func parseAndExecInput(input string) {
	helpStr := `/users :- Get list of live users
@{user} message :- send message to specified user
/exit :- Exit the Chat
/all :- Send message to all the users [TODO]`
	// Split the line into 2 tokens (cmd and message)
	tokens := strings.SplitN(input, " ", 2)
	cmd := tokens[0]
	switch {
	case cmd == "":
		break
	case cmd == "?":
		fmt.Printf(`/users : List all users
/exit  : Exit chat
@<user> Type some message. e.g. @joe This works!
\n`)
		break
	case strings.ToLower(cmd) == "/users":
		fmt.Println(HANDLES)
		break
	case strings.ToLower(cmd) == "/exit":
		os.Exit(1)
		break
	case cmd[0] == '@':
		message := "hi" // default
		if len(tokens) > 1 {
			message = tokens[1]
		}

		// send message to particular user
		if h, ok := HANDLES.Get(cmd[1:]); ok {
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

// cleanup Dead Handlers
func cleanupDeadHandles(wg *sync.WaitGroup, exit chan bool) {
	defer wg.Done()
	// wait for DEAD_HANDLE_INTERVAL seconds before removing them from chatrooms and handle list
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)

	return text
}
