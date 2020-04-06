package event

import (
	"fmt"

	"github.com/tylerwray/amazin/config"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Dispatcher struct{}

func NewDispatcher(cfg config.Values) Dispatcher {
	return Dispatcher{}
}

func (d Dispatcher) Send(event []byte) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			fmt.Println("Checking...")
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			default:
				fmt.Printf("IDK: %+v", ev)
			}
		}
	}()

	topic := "stripe-events"

	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          event,
	}, nil)

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
	fmt.Println("Done.")
}
