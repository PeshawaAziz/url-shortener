package analytics

import (
	"context"
	"log"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"
)

type AsyncClickReporter struct {
	clickChan chan clickJob
}

type clickJob struct {
	url      *url.URL
	metadata url.ClickMetadata
}

func NewAsyncClickReporter(bufferSize int) *AsyncClickReporter {
	r := &AsyncClickReporter{
		clickChan: make(chan clickJob, bufferSize),
	}
	go r.process()
	return r
}

func (r *AsyncClickReporter) RecordClick(ctx context.Context, u *url.URL, meta url.ClickMetadata) {
	job := clickJob{url: u, metadata: meta}
	select {
	case r.clickChan <- job:
	default:
		log.Println("WARNING: Click analytics channel full. Dropping event.")
	}
}

func (r *AsyncClickReporter) process() {
	for job := range r.clickChan {
		// In production:
		// 1. Redis INCR for total click count
		// 2. Push to Kafka/NATS stream for downstream analytics processing
		log.Printf("Async Recorded Click for Slug: %s, IP: %s\n", job.url.Slug, job.metadata.IPAddress)
	}
}
