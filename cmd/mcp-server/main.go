package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chennqqi/godnslog/internal/mcp"
)

func main() {
	// Get configuration
	apiURL := os.Getenv("GODNSLOG_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080/api/v2"
	}

	apiKey := os.Getenv("GODNSLOG_API_KEY")
	if apiKey == "" {
		log.Fatal("GODNSLOG_API_KEY environment variable is required")
	}

	// Create MCP server
	server := mcp.NewServer(apiURL, apiKey)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down MCP server...")
		cancel()
	}()

	log.Println("Starting GODNSLOG MCP Server...")
	if err := server.Run(ctx); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}

	log.Println("MCP server stopped")
}
