package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"avg_weights_fed_ml_itmo/internal/app/mocks"
	"avg_weights_fed_ml_itmo/internal/app/model"
	"avg_weights_fed_ml_itmo/pkg/serverside"
)

func TestService_AddMyWeights(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req *serverside.AddMyWeightsRequest
	}
	type fields struct {
		usecase func(ctrl *gomock.Controller, a *args) UpsertWeightsUsecase
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		errText string
	}{
		{
			name: "success",
			fields: fields{
				usecase: func(ctrl *gomock.Controller, a *args) UpsertWeightsUsecase {
					m := mocks.NewMockUpsertWeightsUsecase(ctrl)
					m.EXPECT().UpsertWeights(gomock.Any(), model.AddMyWeightsDomainReq{
						ClientID:    uuid.MustParse(a.req.ClientId),
						Weights:     a.req.Weights,
						NumExamples: a.req.NumExamples,
					}).Return(nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &serverside.AddMyWeightsRequest{
					ClientId:    uuid.NewString(),
					Weights:     []byte("absdfg"),
					NumExamples: 2,
				},
			},
			errText: "",
		},
		{
			name: "usecase error",
			fields: fields{
				usecase: func(ctrl *gomock.Controller, a *args) UpsertWeightsUsecase {
					m := mocks.NewMockUpsertWeightsUsecase(ctrl)
					m.EXPECT().UpsertWeights(gomock.Any(), model.AddMyWeightsDomainReq{
						ClientID:    uuid.MustParse(a.req.ClientId),
						Weights:     a.req.Weights,
						NumExamples: a.req.NumExamples,
					}).Return(errors.New("some error"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &serverside.AddMyWeightsRequest{
					ClientId:    uuid.NewString(),
					Weights:     []byte("absdfg"),
					NumExamples: uint64(2),
				},
			},
			errText: "some error",
		},
		{
			name: "validation error",
			fields: fields{
				usecase: func(ctrl *gomock.Controller, _ *args) UpsertWeightsUsecase {
					m := mocks.NewMockUpsertWeightsUsecase(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &serverside.AddMyWeightsRequest{
					ClientId:    "bad uuid",
					Weights:     []byte("absdfg"),
					NumExamples: 0,
				},
			},
			errText: "client_id: must be a valid UUID; num_examples: must be greater than 0.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			s := &Service{
				usecase: tt.fields.usecase(ctrl, &tt.args),
			}
			_, err := s.AddMyWeights(tt.args.ctx, tt.args.req)
			if tt.errText != "" {
				require.Error(t, err)
				assert.Equal(t, tt.errText, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}
