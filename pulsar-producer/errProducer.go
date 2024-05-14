package pulsar_producer

import (
	"context"
)

type errProducer struct {
	err error
}

func (e errProducer) Topic() string { return "errTopic" }

func (e errProducer) Name() string { return "errProducer" }

func (e errProducer) Send(ctx context.Context, message *ProducerMessage) (MessageID, error) {
	return nil, e.err
}

func (e errProducer) SendAsync(ctx context.Context, message *ProducerMessage, f func(MessageID, *ProducerMessage, error)) {
	f(nil, nil, e.err)
}

func (e errProducer) Flush() error { return e.err }

func newErrProducer(err error) IPulsarProducer {
	return errProducer{err}
}
