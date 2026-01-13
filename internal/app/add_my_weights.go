package app

import (
	"context"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"avg_weights_fed_ml_itmo/internal/app/model"
	"avg_weights_fed_ml_itmo/pkg/serverside"
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
		validation.Field(&req.NumExamples, validation.By(func(v interface{}) error {
			num, ok := v.(uint64)
			if !ok {
				return fmt.Errorf("value must be uint64")
			}
			if num <= 0 {
				return fmt.Errorf("must be greater than 0")
			}
			return nil
		})),
	)
}
