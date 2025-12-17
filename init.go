package main

import (
	"context"
	"log"

	"github.com/nats-io/nats.go/jetstream"
)

func initStreams(ctx context.Context, js jetstream.JetStream) {
	// Create TASKS stream
	_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "TASKS",
		Subjects: []string{"tasks.*"},
	})
	if err != nil {
		log.Fatal("Failed to create TASKS stream:", err)
	}

	// Create LOGS stream with direct access for reading
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        "LOGS",
		Subjects:    []string{"logs.*"},
		AllowDirect: true,
	})
	if err != nil {
		log.Fatal("Failed to create LOGS stream:", err)
	}
}

func initConsumer(ctx context.Context, js jetstream.JetStream) jetstream.Consumer {
	consumer, err := js.CreateOrUpdateConsumer(ctx, "TASKS", jetstream.ConsumerConfig{
		Durable:   "runner",
		AckPolicy: jetstream.AckNonePolicy,
	})
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	return consumer
}
