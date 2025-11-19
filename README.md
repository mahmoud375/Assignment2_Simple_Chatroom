# Docker Hub Image

**🔗 [Click here to view on Docker Hub: elgendy2003/rpc-chat](https://hub.docker.com/r/elgendy2003/rpc-chat)**

-----

# Go RPC Chat Application

**[Watch the Video Demo (Google Drive)](https://drive.google.com/file/d/15D6090AjTCR2pK33gJGZk-rvTdBKr7R1/view?usp=drive_link)**

A simple client-server chatroom application built with Go (Golang) and the `net/rpc` package. This project demonstrates the fundamentals of RPC, concurrent server handling, and state management using mutexes.

This was built as an assignment based on the requirements in `instructions.md`.

## Features

* **Client-Server Architecture:** Uses Go's `net/rpc` library.
* **Persistent Chat History:** The server maintains a complete history of all messages.
* **Multiple Clients:** The server uses `go rpc.ServeConn(conn)` to handle multiple clients concurrently.
* **Concurrency Safe:** The chat history is protected by a `sync.Mutex` to prevent race conditions.
* **Graceful Exit:** Clients can type `exit` to leave the chat.

## Technologies Used

* **Go (Golang)**
* **Standard Libraries:**
    * `net/rpc` (for remote procedure calls)
    * `net` (for TCP listener)
    * `sync` (for `Mutex`)
    * `log` (for server-side logging)
    * `bufio` (for reading full-line client input)

## Docker

**Docker Hub Image:** [elgendy2003/rpc-chat](https://hub.docker.com/r/elgendy2003/rpc-chat)

### Running with Docker

The easiest way to run the server is using the pre-built Docker image:

```bash
docker pull elgendy2003/rpc-chat:v1
docker run -p 1234:1234 elgendy2003/rpc-chat:v1
```
