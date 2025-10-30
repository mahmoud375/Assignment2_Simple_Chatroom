package main

import (
	"log"
	"net"
	"net/rpc"
	"sync"
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

// ChatServer represents the RPC server
type ChatServer struct {
	history []string
	mu      sync.Mutex
}

// SendMessage handles new messages and returns updated history
func (s *ChatServer) SendMessage(args *MessageArgs, reply *HistoryReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Format and append the new message
	formattedMsg := args.Name + ": " + args.Message
	s.history = append(s.history, formattedMsg)

	log.Printf("Received message from %s: '%s'. History now has %d messages.", args.Name, args.Message, len(s.history))
	// --------------------------

	// Set reply with complete history
	reply.History = make([]string, len(s.history))
	copy(reply.History, s.history)

	return nil
}

// GetHistory returns the current chat history
func (s *ChatServer) GetHistory(_ *struct{}, reply *HistoryReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set reply with complete history
	reply.History = make([]string, len(s.history))
	copy(reply.History, s.history)

	return nil
}

func main() {
	// Create and register the RPC server
	server := new(ChatServer)
	rpc.Register(server)

	// Listen for incoming connections
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen error:", err)
	}

	log.Println("Chat server running on port 1234...")

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v\n", err)
			continue
		}

		go rpc.ServeConn(conn)
	}
}