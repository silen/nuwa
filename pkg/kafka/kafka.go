package kafka

import (
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func KafkaReader(kafkaURL, topic, groupID string, HeartbeatInterval time.Duration) *kafka.Reader {
	brokers := strings.Split(kafkaURL, "|")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
		//MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		//MaxWait:  time.Second,
		MaxWait:           time.Duration(1) * time.Second,
		HeartbeatInterval: HeartbeatInterval * time.Millisecond,
	})
}

var (
	WriterMap = make(map[string]*kafka.Writer)
)

func KafkaWriter(kafkaURL, topic string) *kafka.Writer {
	brokers := strings.Split(kafkaURL, "|")

	if WriterMap[topic] == nil {
		WriterMap[topic] = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Balancer: &kafka.LeastBytes{},
			Topic:    topic,
			// /	Async:    true,
		})
	}

	return WriterMap[topic]

}
