package usecase

import (
	"context"
	"fmt"
	"time"

	dbModel "avg_weights_fed_ml_itmo/generated/metadata_db/public/model"
	"avg_weights_fed_ml_itmo/internal/app/model"

	"github.com/google/uuid"
)

type (
	WeightsRepo interface {
		PutNewWeights(
			ctx context.Context,
			objectKey string,
			weights []byte,
		) error
	}
	MetaDataRepo interface {
		UpsertMetadataByClientID(ctx context.Context, metadata dbModel.Metadata) error
	}

	Upserter struct {
		weightRepo   WeightsRepo
		metadataRepo MetaDataRepo
	}
)

func NewUpserter(clickhouseRepo WeightsRepo) *Upserter {
	return &Upserter{weightRepo: clickhouseRepo}
}

func (u *Upserter) UpsertWeights(ctx context.Context, domainRequest model.AddMyWeightsDomainReq) error {
	// 1. Генерация ключа с timestamp и UUID
	now := time.Now().UTC()
	ts := now.Format("20060102_150405")
	uuidStr := uuid.New().String()
	objectKey := fmt.Sprintf("%s_%s_%s.pt", ts, domainRequest.ClientID, uuidStr)

	err := u.weightRepo.PutNewWeights(ctx, objectKey, domainRequest.Weights)
	if err != nil {
		return err
	}

	if err = u.metadataRepo.UpsertMetadataByClientID(ctx, dbModel.Metadata{
		ClientID:    domainRequest.ClientID,
		CreatedAt:   now,
		UpdatedAt:   now,
		NumExamples: int32(domainRequest.NumExamples),
		ObjectKey:   objectKey,
	}); err != nil {
		return err
	}
	return nil
}
