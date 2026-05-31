package audit

type Subscriber struct {
	channel chan Event
}

func NewSubscriber(bufferSize int) *Subscriber {
	return &Subscriber{channel: make(chan Event, bufferSize)}
}

func (s *Subscriber) Read() <-chan Event {
	return s.channel
}
