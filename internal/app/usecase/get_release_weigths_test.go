package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"avg_weights_fed_ml_itmo/internal/app/usecase/mocks"
)

func TestGetter_GetNewWeights(t *testing.T) {
	t.Parallel()
	type fields struct {
		getWeightsRepo func(ctrl *gomock.Controller) GetWeightsRepo
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr string
	}{
		{
			name: "lol test",
			fields: fields{
				getWeightsRepo: func(ctrl *gomock.Controller) GetWeightsRepo {
					m := mocks.NewMockGetWeightsRepo(ctrl)
					m.EXPECT().GetDownloadURL(gomock.Any()).Return("aboba", errors.New("aboba213"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:    "aboba",
			wantErr: "aboba213",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			g := &Getter{
				getWeightsRepo: tt.fields.getWeightsRepo(ctrl),
			}
			got, err := g.GetNewWeights(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err.Error())
			assert.Equal(t, tt.want, got)
		})
	}
}
