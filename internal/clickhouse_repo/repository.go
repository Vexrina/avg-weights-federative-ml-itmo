package clickhouse_repo

import (
	"context"
	"log"
	"time"

	"github.com/uptrace/go-clickhouse/ch"

	dbModel "avg_weights_fed_ml_itmo/internal/clickhouse_repo/model"
)

type weightRepository struct {
	db *ch.DB
}

func NewWeightRepo(ctx context.Context) weightRepository {
	db := ch.Connect(
		ch.WithAddr("localhost:9002"),
		ch.WithDatabase("default"),
		ch.WithUser("root"),
		ch.WithPassword("root"),
		ch.WithConnMaxIdleTime(2*time.Second),
		ch.WithPoolSize(10),
	)

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping clickhouse: %v", err)
	}

	return weightRepository{db: db}
}

func (wr *weightRepository) UpsertActualWeights(ctx context.Context, dbReq dbModel.ActualLayersByClientId) error {
	stmt := wr.db.NewInsert().Model(&dbReq)

	if _, err := stmt.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (wr *weightRepository) AppendHistoryWeights(ctx context.Context, dbReq dbModel.HistoryLayers) error {
	stmt := wr.db.NewInsert().Model(&dbReq)
	if _, err := stmt.Exec(ctx); err != nil {
		return err
	}
	return nil
}
