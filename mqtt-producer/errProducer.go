package mqtt_producer

import (
	"context"
)

type errProducer struct {
	err error
}

func (e errProducer) Send(ctx context.Context, msg *ProducerMessage) error {
	return e.err
}

func (e errProducer) SendAsync(ctx context.Context, msg *ProducerMessage, callback func(error)) {
	callback(e.err)
}

func newErrProducer(err error) Client {
	return errProducer{err}
}
