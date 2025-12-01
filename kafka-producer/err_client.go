package kafka_producer

type errClient struct {
	err error
}

func (e *errClient) SendMessage(msg *ProducerMessage) (partition int32, offset int64, err error) {
	return 0, 0, e.err
}

func (e *errClient) SendMessages(msgs []*ProducerMessage) error {
	return e.err
}

func newErrClient(err error) Client {
	return &errClient{
		err: err,
	}
}
