package model

// так хотелось бы, чтобы был кодген...

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"
)

type HistoryLayers struct {
	ch.CHModel `ch:"table:history_of_layers_by_client_id"`

	ClientID  uuid.UUID `ch:"clientID"`
	CreatedAt time.Time `ch:"createdAt"`
	Layer1    []float64 `ch:"layer1"`
	Layer2    []float64 `ch:"layer2"`
	Layer3    []float64 `ch:"layer3"`
	Layer4    []float64 `ch:"layer4"`
	Layer5    []float64 `ch:"layer5"`
}
