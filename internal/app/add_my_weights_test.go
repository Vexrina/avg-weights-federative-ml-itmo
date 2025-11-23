package app

import (
	"avg_weights_fed_ml_itmo/internal/app/mocks"
	"avg_weights_fed_ml_itmo/internal/app/model"
	"avg_weights_fed_ml_itmo/pkg/serverside"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
						ClientID: uuid.MustParse(a.req.ClientId),
						Layers: map[string][]float64{
							"123": {1, 3, 5},
							"231": {3, 5, 1},
							"321": {5, 3, 1},
							"2":   {5},
							"1":   {5},
						},
					}).Return(nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &serverside.AddMyWeightsRequest{
					ClientId: uuid.NewString(),
					Layers: map[string]*serverside.AddMyWeightsRequest_Layer{
						"123": {Weights: []float64{1, 3, 5}},
						"231": {Weights: []float64{3, 5, 1}},
						"321": {Weights: []float64{5, 3, 1}},
						"2":   {Weights: []float64{5}},
						"1":   {Weights: []float64{5}},
					},
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
						ClientID: uuid.MustParse(a.req.ClientId),
						Layers: map[string][]float64{
							"123": {1, 3, 5},
							"231": {3, 5, 1},
							"321": {5, 3, 1},
							"2":   {5},
							"1":   {5},
						},
					}).Return(errors.New("some error"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &serverside.AddMyWeightsRequest{
					ClientId: uuid.NewString(),
					Layers: map[string]*serverside.AddMyWeightsRequest_Layer{
						"123": {Weights: []float64{1, 3, 5}},
						"231": {Weights: []float64{3, 5, 1}},
						"321": {Weights: []float64{5, 3, 1}},
						"2":   {Weights: []float64{5}},
						"1":   {Weights: []float64{5}},
					},
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
					ClientId: uuid.NewString(),
					Layers: map[string]*serverside.AddMyWeightsRequest_Layer{
						"123": {Weights: []float64{1, 3, 5}},
						"231": {Weights: []float64{3, 5, 1}},
					},
				},
			},
			errText: "layers: the length must be exactly 5.",
		},
		{
			name: "another validation error",
			fields: fields{
				usecase: func(ctrl *gomock.Controller, _ *args) UpsertWeightsUsecase {
					m := mocks.NewMockUpsertWeightsUsecase(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &serverside.AddMyWeightsRequest{
					ClientId: "asd",
				},
			},
			errText: "client_id: must be a valid UUID; layers: cannot be blank.",
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
