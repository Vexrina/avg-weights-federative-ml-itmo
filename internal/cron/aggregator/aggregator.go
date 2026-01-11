// aggregator.go
package aggregator

import (
	"context"
	"log"
	"time"
)

type MinioRepo interface {
	AcquireLock(ctx context.Context) (bool, error)
	LoadLastAggregationTs(ctx context.Context) (time.Time, error)
	ListObjectAfter(ctx context.Context, from time.Time) ([]*WeightObject, error)
	SaveReleaseWeights(ctx context.Context, weights []byte) error
	SaveLastAggregationTs(ctx context.Context, timestamp time.Time) error
}

type (
	accumulatorEntry struct {
		weights     []byte
		numExamples uint64
	}
	aggregator struct {
		minioRepo     MinioRepo
		entries       []accumulatorEntry
		totalExamples uint64
	}
	fedAvgPayload struct {
		TotalExamples uint64
		Entries       []accumulatorEntry
	}
)

const aggInterval = time.Minute * 10

func NewAggregator(minioRepo MinioRepo) *aggregator {
	return &aggregator{minioRepo: minioRepo}
}

func (a *aggregator) Aggregate(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("[ERROR] context already done in aggregator")
			return
		default:
		}

		// 1. Пытаемся взять lock
		ok, err := a.minioRepo.AcquireLock(ctx)
		if err != nil {
			log.Printf("[ERROR] acquire lock failed: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		if !ok {
			// другой агрегатор работает
			time.Sleep(time.Minute)
			continue
		}

		lastTs, err := a.minioRepo.LoadLastAggregationTs(ctx)
		if err != nil {
			log.Printf("[ERROR] load last ts failed: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		now := time.Now().UTC()
		nextRun := lastTs.Add(aggInterval)

		if now.Before(nextRun) {
			// Нечего делать — освобождаем lock
			time.Sleep(nextRun.Sub(now))
			continue
		}

		if err := a.aggregateOnce(ctx, lastTs); err != nil {
			log.Printf("[ERROR] aggregateOnce failed: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		if err := a.minioRepo.SaveLastAggregationTs(ctx, now); err != nil {
			log.Printf("[ERROR] save watermark failed: %v", err)
			time.Sleep(time.Minute)
			continue
		}
	}
}

func (a *aggregator) aggregateOnce(
	ctx context.Context,
	fromTs time.Time,
) error {
	a.totalExamples = 0
	a.entries = a.entries[:0]

	objs, err := a.minioRepo.ListObjectAfter(ctx, fromTs)
	if err != nil {
		return err
	}

	if len(objs) == 0 {
		return nil // нечего агрегировать
	}

	for _, obj := range objs {
		if obj.NumExamples == 0 {
			continue // или return error — на твой выбор
		}

		a.accumulate(obj.Weights, obj.NumExamples)
	}

	if a.totalExamples == 0 {
		return nil
	}

	globalDelta, err := a.finalize(ctx)
	if err != nil {
		return err
	}

	return a.minioRepo.SaveReleaseWeights(ctx, globalDelta)
}

func (a *aggregator) accumulate(weights []byte, numExamples uint64) {
	a.entries = append(a.entries, accumulatorEntry{
		weights:     weights,
		numExamples: numExamples,
	})
	a.totalExamples += numExamples
}
