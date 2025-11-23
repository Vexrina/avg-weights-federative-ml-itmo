package model

import (
	"avg_weights_fed_ml_itmo/pkg/serverside"

	"github.com/google/uuid"
)

type AddMyWeightsDomainReq struct {
	ClientID uuid.UUID
	Layers   map[string][]float64
}

func MapAddMyWeightsToDomain(request *serverside.AddMyWeightsRequest) AddMyWeightsDomainReq {
	l := make(map[string][]float64)
	for layerName, layer := range request.Layers {
		l[layerName] = layer.Weights
	}
	return AddMyWeightsDomainReq{
		ClientID: uuid.MustParse(request.ClientId),
		Layers:   l,
	}
}
