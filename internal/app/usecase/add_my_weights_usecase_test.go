package usecase

import (
	"avg_weights_fed_ml_itmo/internal/app/model"
	"avg_weights_fed_ml_itmo/internal/app/usecase/mocks"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpserter_UpsertWeights(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx           context.Context
		domainRequest model.AddMyWeightsDomainReq
	}
	type fields struct {
		weightRepo func(ctrl *gomock.Controller, a *args) WeightsRepo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr string
	}{
		{
			name: "success",
			fields: fields{
				weightRepo: func(ctrl *gomock.Controller, a *args) WeightsRepo {
					m := mocks.NewMockWeightsRepo(ctrl)
					m.EXPECT().PutNewWeights(gomock.Any(), gomock.Any(), a.domainRequest.Weights).DoAndReturn(
						func(_ context.Context, ok string, _ []byte) error {
							stringArr := strings.Split(ok, "_")
							if len(stringArr) != 4 {
								return errors.New("broke object key format")
							}
							if stringArr[1] != a.domainRequest.ClientID.String() || stringArr[2] != fmt.Sprint(a.domainRequest.NumExamples) {
								return errors.New("broke object key")
							}
							return nil
						},
					)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				domainRequest: model.AddMyWeightsDomainReq{
					ClientID:    uuid.New(),
					Weights:     []byte("1"),
					NumExamples: 123,
				},
			},
		},
		{
			name: "err",
			fields: fields{
				weightRepo: func(ctrl *gomock.Controller, a *args) WeightsRepo {
					m := mocks.NewMockWeightsRepo(ctrl)
					m.EXPECT().PutNewWeights(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("aboba"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				domainRequest: model.AddMyWeightsDomainReq{
					ClientID:    uuid.New(),
					Weights:     []byte("1"),
					NumExamples: 123,
				},
			},
			wantErr: "aboba",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			u := &Upserter{
				weightRepo: tt.fields.weightRepo(ctrl, &tt.args),
			}

			err := u.UpsertWeights(tt.args.ctx, tt.args.domainRequest)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
