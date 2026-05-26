package audit

import (
	"context"
	"log/slog"
)

type Event struct {
	Ts     int    `json:"ts"`
	Action string `json:"action"`
	UserId string `json:"user_id"`
	URL    string `json:"url"`
}

func NewAuditService(ctx context.Context) *AuditService {
	subCh := make(chan chan<- Event)
	unsubCh := make(chan chan<- Event)
	eventCh := make(chan Event)

	Service := &AuditService{subCh: subCh, eventCh: eventCh, unsubCh: unsubCh}

	go func() {

		var subscriber []chan<- Event

		for {
			select {
			case <-ctx.Done():
				for _, ch := range subscriber {
					close(ch)
				}
				return
			case ch := <-subCh:
				subscriber = append(subscriber, ch)

			case target := <-unsubCh:
				for i, ch := range subscriber {
					if target == ch {
						subscriber = append(subscriber[:i], subscriber[i+1:]...)
						close(ch)
						break
					}
				}

			case e := <-eventCh:
				for _, ch := range subscriber {
					select {
					case ch <- e:
					default:
						slog.Warn("subscriber channel is full, skipping event", "event", e)
					}
				}
			}
		}
	}()

	return Service
}

type AuditService struct {
	unsubCh chan chan<- Event
	subCh   chan chan<- Event
	eventCh chan Event
}

func (s *AuditService) Emit(e Event) {
	s.eventCh <- e
}

func (s *AuditService) Subscribe(sub *Subscriber) {
	s.subCh <- sub.channel
}

func (s *AuditService) Unsubscribe(sub *Subscriber) {
	s.unsubCh <- sub.channel
}
