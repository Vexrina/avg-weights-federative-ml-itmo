package model

import (
	"github.com/google/uuid"

	"avg_weights_fed_ml_itmo/pkg/serverside"
)

type AddMyWeightsDomainReq struct {
	ClientID    uuid.UUID
	Weights     []byte
	NumExamples uint64
}

func MapAddMyWeightsToDomain(request *serverside.AddMyWeightsRequest) AddMyWeightsDomainReq {
	return AddMyWeightsDomainReq{
		ClientID:    uuid.MustParse(request.ClientId),
		Weights:     request.Weights,
		NumExamples: request.NumExamples,
	}
}
