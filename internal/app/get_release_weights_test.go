package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"avg_weights_fed_ml_itmo/internal/app/mocks"
	"avg_weights_fed_ml_itmo/pkg/serverside"
)

func TestService_GetReleaseWeights(t *testing.T) {
	type fields struct {
		getter func(ctrl *gomock.Controller) GetterWeightsUsecase
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *serverside.GetReleaseWeightsResponse
		wantErr string
	}{
		{
			name: "success",
			fields: fields{
				getter: func(ctrl *gomock.Controller) GetterWeightsUsecase {
					m := mocks.NewMockGetterWeightsUsecase(ctrl)
					m.EXPECT().GetNewWeights(gomock.Any()).Return("aboba", nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: &serverside.GetReleaseWeightsResponse{
				LinkToMinio: "aboba",
			},
		},
		{
			name: "error",
			fields: fields{
				getter: func(ctrl *gomock.Controller) GetterWeightsUsecase {
					m := mocks.NewMockGetterWeightsUsecase(ctrl)
					m.EXPECT().GetNewWeights(gomock.Any()).Return("", errors.New("aboba"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: "aboba",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			s := &Service{
				getter: tt.fields.getter(ctrl),
			}
			got, err := s.GetReleaseWeights(tt.args.ctx, nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
