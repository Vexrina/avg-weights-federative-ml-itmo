package app

import (
	"context"

	"avg_weights_fed_ml_itmo/internal/app/model"
	"avg_weights_fed_ml_itmo/pkg/serverside"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_usecases.go -package=mocks
type (
	UpsertWeightsUsecase interface {
		UpsertWeights(ctx context.Context, domainRequest model.AddMyWeightsDomainReq) error
	}
	GetterWeightsUsecase interface {
		GetNewWeights(ctx context.Context) (string, error)
	}
)

type Service struct {
	serverside.UnimplementedAvgWeightsServer

	usecase UpsertWeightsUsecase
	getter  GetterWeightsUsecase
}

func NewService(
	ups UpsertWeightsUsecase,
	get GetterWeightsUsecase,
) *Service {
	return &Service{
		usecase: ups,
		getter:  get,
	}
}
