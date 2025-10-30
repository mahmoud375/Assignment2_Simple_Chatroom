package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
)

// MessageArgs represents the arguments for sending a message
type MessageArgs struct {
	Name    string
	Message string
}

// HistoryReply represents the response containing chat history
type HistoryReply struct {
	History []string
}

func main() {
	// Connect to the RPC server
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer client.Close()

	// Get user's name
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading name:", err)
	}
	name = strings.TrimSpace(name)

	fmt.Printf("Welcome, %s! You can start chatting.\n", name)

	// Main chat loop
	for {
		fmt.Print("Enter message (or 'exit' to quit): ")
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading message:", err)
		}
		message = strings.TrimSpace(message)

		// Check if user wants to exit
		if message == "exit" {
			break
		}

		// Prepare the message arguments and reply
		args := &MessageArgs{
			Name:    name,
			Message: message,
		}
		var reply HistoryReply

		// Send the message to the server
		err = client.Call("ChatServer.SendMessage", args, &reply)
		if err != nil {
			log.Fatal("RPC error:", err)
		}

		// Print chat history
		fmt.Println("\n--- Chat History ---")
		for _, msg := range reply.History {
			fmt.Println(msg)
		}
		fmt.Println("------------------\n")
	}

	fmt.Println("Goodbye!")
}