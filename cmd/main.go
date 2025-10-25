package main

import (
	"fmt"
	"os"

	"cloudhub/internal/docker"
)

func main() {

	args := os.Args

	fmt.Println("🏠 Cloudhub - Docker Monitor")
	fmt.Println("─────────────────────────────────────\n")

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

	//getting docker version
	if len(args) > 1 && args[1] == "version" {

		fmt.Println("\n📦 Docker Information:")
		ver, err := client.GetVersion()
		if err != nil {
			fmt.Printf("❌ failed to get docker version: %v\n", err)
			return
		}

		fmt.Println("─────────────────────────────────────")
		fmt.Printf("Version:        %s\n", ver.Version)
		fmt.Printf("API Version:    %s\n", ver.APIVersion)
		fmt.Printf("OS/Arch:        %s/%s\n", ver.Os, ver.Arch)
		fmt.Printf("Build Time:     %s\n", ver.BuildTime)
		fmt.Printf("Git Commit:     %s\n", ver.GitCommit)
		fmt.Println("─────────────────────────────────────")
	}

	if len(args) > 1 && args[1] == "ls" {
		fmt.Println("📦 Containers:")

		containers, err := client.ListContainers()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		if len(containers) == 0 {
			fmt.Println("   No containers running")
			return
		}

		for i, c := range containers {
			fmt.Printf("   %d. %s (%s)\n", i+1, c.Name, c.ID)
		}
		fmt.Printf("\nTotal: %d containers\n", len(containers))
	}
}
