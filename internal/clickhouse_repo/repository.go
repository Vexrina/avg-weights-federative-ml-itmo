package clickhouse_repo

import (
	dbModel "avg_weights_fed_ml_itmo/internal/clickhouse_repo/model"
	"context"
	"log"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
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

func (wr *weightRepository) UpsertActualWeights(ctx context.Context, dbReq dbModel.ActualLayers) error {
	stmt := wr.db.NewInsert().Model(&dbReq).Table("actual_layers_by_client_id")

	if _, err := stmt.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (wr *weightRepository) AppendHistoryWeights(ctx context.Context, dbReq dbModel.HistoryLayers) error {
	stmt := wr.db.NewInsert().Model(&dbReq).Table("history_layers_by_client_id")
	if _, err := stmt.Exec(ctx); err != nil {
		return err
	}
	return nil
}
