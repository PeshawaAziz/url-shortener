package workers

import (
	"context"
	"log"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"
)

type ExpirationWorker struct {
	svc *url.LifecycleService
}

func NewExpirationWorker(svc *url.LifecycleService) *ExpirationWorker {
	return &ExpirationWorker{svc: svc}
}

func (w *ExpirationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("Expiration worker started. Sweeping every 5 minutes.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Expiration worker stopping.")
			return
		case <-ticker.C:
			expired, err := w.svc.SweepExpired(ctx, 1000)
			if err != nil {
				log.Printf("Error during expiration sweep: %v\n", err)
				continue
			}
			if expired > 0 {
				log.Printf("Sweep complete. Expired %d links.\n", expired)
			}
		}
	}
}
