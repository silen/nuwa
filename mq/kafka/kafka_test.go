package kafka

import "testing"

func TestKafkaWriterKeyIncludesKafkaURL(t *testing.T) {
	t.Parallel()

	key1 := kafkaWriterKey("host1:9092", "topic-a")
	key2 := kafkaWriterKey("host2:9092", "topic-a")

	if key1 == key2 {
		t.Fatalf("expected different keys for different kafka urls")
	}
}
