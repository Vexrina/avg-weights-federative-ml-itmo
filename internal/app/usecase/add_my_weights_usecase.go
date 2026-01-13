package usecase

import (
	"context"
	"fmt"
	"time"

	"avg_weights_fed_ml_itmo/internal/app/model"

	"github.com/google/uuid"
)

//go:generate mockgen -source=add_my_weights_usecase.go -destination=mocks/mock_weights_repo.go -package=mocks
type (
	WeightsRepo interface {
		PutNewWeights(
			ctx context.Context,
			objectKey string,
			weights []byte,
		) error
	}

	Upserter struct {
		weightRepo WeightsRepo
	}
)

func NewUpserter(minioRepo WeightsRepo) *Upserter {
	return &Upserter{weightRepo: minioRepo}
}

func (u *Upserter) UpsertWeights(ctx context.Context, domainRequest model.AddMyWeightsDomainReq) error {
	now := time.Now().UTC()
	ts := now.Format("20060102T150405Z")
	uuidStr := uuid.New().String()
	objectKey := fmt.Sprintf("%s_%s_%d_%s.pt", ts, domainRequest.ClientID, domainRequest.NumExamples, uuidStr)

	err := u.weightRepo.PutNewWeights(ctx, objectKey, domainRequest.Weights)
	if err != nil {
		return err
	}
	return nil
}
