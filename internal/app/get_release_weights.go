package app

import (
	"context"

	"avg_weights_fed_ml_itmo/pkg/serverside"
)

func (s *Service) GetReleaseWeights(ctx context.Context, _ *serverside.GetReleaseWeightsRequest) (*serverside.GetReleaseWeightsResponse, error) {
	downloadUrl, err := s.getter.GetNewWeights(ctx)
	if err != nil {
		return nil, err
	}
	return &serverside.GetReleaseWeightsResponse{
		LinkToMinio: downloadUrl,
	}, nil
}
