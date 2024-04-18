package main

import "go_gin_chat_fuke/server"

func main() {
	server := server.NewServer()
	server.Run()
}
