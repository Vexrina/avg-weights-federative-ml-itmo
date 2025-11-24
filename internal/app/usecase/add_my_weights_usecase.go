package usecase

import (
	"context"
	"time"

	"avg_weights_fed_ml_itmo/internal/app/model"
	dbModel "avg_weights_fed_ml_itmo/internal/clickhouse_repo/model"
)

type (
	WeightsRepo interface {
		UpsertActualWeights(ctx context.Context, dbReq dbModel.ActualLayersByClientId) error
		AppendHistoryWeights(ctx context.Context, dbReq dbModel.HistoryLayers) error
	}
	Upserter struct {
		clickhouseRepo WeightsRepo
	}
)

func NewUpserter(clickhouseRepo WeightsRepo) *Upserter {
	return &Upserter{clickhouseRepo: clickhouseRepo}
}

func (u *Upserter) UpsertWeights(ctx context.Context, domainRequest model.AddMyWeightsDomainReq) error {
	dbReq := dbModel.MapDomainReqToDB(domainRequest, time.Now())
	if err := u.clickhouseRepo.UpsertActualWeights(ctx, dbReq); err != nil {
		return err
	}

	dbHistoryReq := dbModel.HistoryLayers{
		ClientID:  dbReq.ClientID,
		CreatedAt: dbReq.UpdatedAt,
		Layer1:    dbReq.Layer1,
		Layer2:    dbReq.Layer2,
		Layer3:    dbReq.Layer3,
		Layer4:    dbReq.Layer4,
		Layer5:    dbReq.Layer5,
	}
	if err := u.clickhouseRepo.AppendHistoryWeights(ctx, dbHistoryReq); err != nil {
		return err
	}

	return nil
}
