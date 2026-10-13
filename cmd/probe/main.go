package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func probe(url string) error {
	client := &http.Client{
		Timeout:       2 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: probe URL")
		os.Exit(2)
	}
	url := os.Args[1]
	if url == "llama" || url == "whisper" {
		key, port, path := "LLAMA_ARG_PORT", "9668", "/health"
		if url == "whisper" {
			key, port, path = "WHISPER_PORT", "9670", "/health"
		}
		if value := os.Getenv(key); value != "" {
			port = value
		}
		url = "http://" + net.JoinHostPort("127.0.0.1", port) + path
	}
	if err := probe(url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
