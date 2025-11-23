package model

import (
	"github.com/google/uuid"

	"avg_weights_fed_ml_itmo/pkg/serverside"
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
