package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func main()  {
	broker := "localhost:9092"
	topic := "test-topic"
	groupID := "app-group"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic: topic,
		GroupID: groupID,
	})
	defer reader.Close()

	log.Println("Waiting for the message...")
	
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln("failed to read message: ", err.Error())
		}

		log.Printf("Got message: key=%s value=%s", msg.Key, msg.Value)
	}
}
