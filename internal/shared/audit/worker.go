package audit

import "context"

func StartWorker(ctx context.Context, storage Storage, sub *Subscriber, server *AuditService) {
	for {
		select {
		case <-ctx.Done():
			server.Unsubscribe(sub)
			return
		case e, ok := <-sub.Read():
			if !ok {
				return
			}
			storage.Send(ctx, e)
		}
	}
}
