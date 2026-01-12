package app

import (
	"avg_weights_fed_ml_itmo/pkg/serverside"
	"context"
)

func (s *Service) GetReleaseWeights(ctx context.Context, req *serverside.GetReleaseWeightsRequest) (*serverside.GetReleaseWeightsResponse, error) {
	_ = ctx
	_ = req
	return &serverside.GetReleaseWeightsResponse{}, nil
}
