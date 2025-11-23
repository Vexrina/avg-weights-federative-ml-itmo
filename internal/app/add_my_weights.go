package app

import (
	"avg_weights_fed_ml_itmo/internal/app/model"
	"context"

	"avg_weights_fed_ml_itmo/pkg/serverside"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

const layersSize = 5

func (s *Service) AddMyWeights(ctx context.Context, req *serverside.AddMyWeightsRequest) (*serverside.AddMyWeightsResponse, error) {
	err := validateAddMyWeightsRequest(req)
	if err != nil {
		return nil, err
	}
	domainReq := model.MapAddMyWeightsToDomain(req)

	err = s.usecase.UpsertWeights(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	return &serverside.AddMyWeightsResponse{}, nil
}

func validateAddMyWeightsRequest(req *serverside.AddMyWeightsRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.ClientId, validation.Required, is.UUID),
		validation.Field(&req.Layers, validation.Required, validation.Length(layersSize, layersSize)),
	)
}
