package kafka_producer

import (
	"github.com/Shopify/sarama"
)

type (
	ProducerMessage = sarama.ProducerMessage
	Encoder         = sarama.Encoder
	StringEncoder   = sarama.StringEncoder
	ByteEncoder     = sarama.ByteEncoder
	RecordHeader    = sarama.RecordHeader
	Message         = sarama.Message
	ProducerError   = sarama.ProducerError
)
