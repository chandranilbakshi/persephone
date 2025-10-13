package main

import (
	"Persephone/internal/purrCommands"
	"log"
)

func main() {
	err := purrCommands.InitPurrDirectories(".")
	if err != nil {
		log.Fatalf("Failed to initialize git directories: %v", err)
	}
}
