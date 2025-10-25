package main

import (
	"fmt"
	// "os"

	"cloudhub/internal/docker"
)

func main()  {
	fmt.Println("🏠 Cloudhub - Docker Monitor")
	fmt.Println("Attempting to connect to Docker...\n")

	client, err := docker.NewClient()

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("💡 Make sure Docker is running!")
		return
	}
	defer client.Close()

	fmt.Println("✅ Successfully connected to Docker!")

	//pinging docker
	fmt.Println("Pinging docker daemon...")
	err = client.Ping()
	if err != nil {
		fmt.Printf("❌ Ping failed: %v\n", err)
		return
	}

	fmt.Println("✅ Docker is responding!")
}
