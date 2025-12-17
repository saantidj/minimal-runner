package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx := context.Background()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal("Failed to create JetStream:", err)
	}

	// Initialize streams and consumer
	initStreams(ctx, js)
	consumer := initConsumer(ctx, js)

	fmt.Println("Waiting for task...")

	// Wait for a task (1 month timeout)
	msg, err := consumer.Next(jetstream.FetchMaxWait(30 * 24 * time.Hour))
	if err != nil {
		log.Fatal("Failed to fetch task:", err)
	}

	// Extract task ID from subject (e.g., "tasks.job-001" -> "job-001")
	subject := msg.Subject()
	taskID := strings.TrimPrefix(subject, "tasks.")
	scriptPath := fmt.Sprintf("/tmp/%s.sh", taskID)
	logSubject := fmt.Sprintf("logs.%s", taskID)

	fmt.Printf("Received task: %s\n", taskID)

	// Save script to file
	err = os.WriteFile(scriptPath, msg.Data(), 0755)
	if err != nil {
		log.Fatal("Failed to write script:", err)
	}

	// Execute with 5 minute timeout
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "bash", scriptPath)

	// Get pipes BEFORE starting
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("Failed to get stdout pipe:", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal("Failed to get stderr pipe:", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to start command:", err)
	}

	// Stream stdout and stderr in real-time
	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			js.Publish(ctx, logSubject, []byte(line))
			fmt.Println(line)
		}
	}()

	// Stream stderr with ERROR:: prefix
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := "ERROR::" + scanner.Text()
			js.Publish(ctx, logSubject, []byte(line))
			fmt.Println(line)
		}
	}()

	// Wait for output streams to finish
	wg.Wait()

	// Wait for command to complete
	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Publish exit code
	exitMsg := fmt.Sprintf("EXIT:%d", exitCode)
	js.Publish(ctx, logSubject, []byte(exitMsg))
	fmt.Println(exitMsg)

	fmt.Printf("Task %s completed with exit code %d\n", taskID, exitCode)
}
