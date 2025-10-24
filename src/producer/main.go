package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main()  {
	broker := "localhost:9092"
	topic := "test-topic"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic: topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	for i := 0; i < 5; i++ {
		log.Println("Writing message #", i)

		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key: []byte(fmt.Sprintf("Key-%d", i)),
			Value: []byte(fmt.Sprintf("Message #%d", i)),
		})

		if err != nil {
			log.Fatalln("failed to write message: ", err.Error())
		}
		
		time.Sleep(time.Second)
	}
}
