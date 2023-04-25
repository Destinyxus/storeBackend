package main

import "github.com/Destinyxus/storeAPI/internal/apiserver"

func main() {
	server := apiserver.NewAPIServer()
	server.Run()
}
