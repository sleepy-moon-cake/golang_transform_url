package audit

import (
	"context"
	"io"
	"log/slog"
)

func StartWorker(ctx context.Context, storage Storage, sub *Subscriber, server *AuditService) {
	defer func() {
		server.Unsubscribe(sub)

		if closer, ok := storage.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				slog.Error("error closing storage in worker", "err", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-sub.Read():
			if !ok {
				return
			}
			storage.Send(ctx, e)
		}
	}
}
