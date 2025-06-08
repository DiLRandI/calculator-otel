package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := http.Client{
		Timeout: time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/ping", nil)
	if err != nil {
		os.Stderr.WriteString("Failed to create request: " + err.Error() + "\n")
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		os.Stderr.WriteString("Failed to send request: " + err.Error() + "\n")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Stderr.WriteString("Unexpected status code: " + resp.Status + "\n")
		return
	}
	os.Stdout.WriteString("Ping successful: " + resp.Status + "\n")

	<-ctx.Done()
	os.Stdout.WriteString("Received shutdown signal, exiting...\n")
}
