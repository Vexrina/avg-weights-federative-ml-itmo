package usecase

import "context"

//go:generate mockgen -source=get_release_weigths.go -destination=mocks/mock_get_repo.go -package=mocks
type GetWeightsRepo interface {
	GetDownloadURL(ctx context.Context) (string, error)
}

type Getter struct {
	getWeightsRepo GetWeightsRepo
}

func NewGetter(getWeightsRepo GetWeightsRepo) *Getter {
	return &Getter{
		getWeightsRepo: getWeightsRepo,
	}
}

func (g *Getter) GetNewWeights(ctx context.Context) (string, error) {
	return g.getWeightsRepo.GetDownloadURL(ctx)
}
