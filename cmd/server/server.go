package main

import "bloccs-server/internal/server"

func main() {
	srv := server.New()

	err := srv.Start()

	srv.Stop()

	if err != nil {
		panic(err)
	}
}
