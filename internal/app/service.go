package app

import (
	"context"

	"avg_weights_fed_ml_itmo/internal/app/model"
	"avg_weights_fed_ml_itmo/pkg/serverside"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_upser_weights_usecase.go -package=mocks UpsertWeightsUsecase
type (
	UpsertWeightsUsecase interface {
		UpsertWeights(ctx context.Context, domainRequest model.AddMyWeightsDomainReq) error
	}
	AggregateWeights interface {
		Aggregate(ctx context.Context)
	}
)

type Service struct {
	serverside.UnimplementedAvgWeightsServer

	usecase UpsertWeightsUsecase
}

func NewService(ups UpsertWeightsUsecase) *Service {
	return &Service{
		usecase: ups,
	}
}
