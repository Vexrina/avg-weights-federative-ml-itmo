package clickhouse_repo

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"avg_weights_fed_ml_itmo/internal/clickhouse_repo/model"
)

func Test_weightRepository_UpsertActualWeights(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx   context.Context
		dbReq model.ActualLayersByClientId
	}
	tests := []struct {
		name    string
		args    args
		prepare func(wr *weightRepository, a *args)
		check   func(wr *weightRepository, a *args)
		wantErr bool
	}{
		{
			name: "success insert",
			args: args{
				ctx: ctx,
				dbReq: model.ActualLayersByClientId{
					ClientID:  uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Layer1:    nil,
					Layer2:    nil,
					Layer3:    nil,
					Layer4:    nil,
					Layer5:    nil,
				},
			},
			check: func(wr *weightRepository, a *args) {
				res := model.ActualLayersByClientId{}
				err := wr.db.NewSelect().Model(&model.ActualLayersByClientId{}).Scan(ctx, &res)
				if err != nil {
					t.Fatalf("%v", err.Error())
				}
				res.CreatedAt = res.CreatedAt.Truncate(time.Second).UTC()
				res.UpdatedAt = res.UpdatedAt.Truncate(time.Second).UTC()

				a.dbReq.CreatedAt = a.dbReq.CreatedAt.Truncate(time.Second).UTC()
				a.dbReq.UpdatedAt = a.dbReq.UpdatedAt.Truncate(time.Second).UTC()
				if !cmp.Equal(res, a.dbReq) {
					t.Fatalf("%v", cmp.Diff(res, a.dbReq))
				}
			},
			prepare: func(wr *weightRepository, a *args) {
				return
			},
		},
		{
			name: "success upsert",
			args: args{
				ctx: ctx,
				dbReq: model.ActualLayersByClientId{
					ClientID:  uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Layer1:    nil,
					Layer2:    nil,
					Layer3:    nil,
					Layer4:    nil,
					Layer5:    nil,
				},
			},
			prepare: func(wr *weightRepository, a *args) {
				inserted := model.ActualLayersByClientId{
					ClientID:  a.dbReq.ClientID,
					CreatedAt: time.Now().Add(-time.Hour),
					UpdatedAt: time.Now().Add(-time.Hour),
					Layer1:    nil,
					Layer2:    nil,
					Layer3:    nil,
					Layer4:    nil,
					Layer5:    nil,
				}
				prepErr := wr.UpsertActualWeights(ctx, inserted)
				if prepErr != nil {
					t.Fatalf("%v", prepErr.Error())
				}
			},
			check: func(wr *weightRepository, a *args) {
				res := model.ActualLayersByClientId{}
				err := wr.db.NewSelect().Model(&model.ActualLayersByClientId{}).Scan(ctx, &res)
				if err != nil {
					t.Fatalf("%v", err.Error())
				}
				res.CreatedAt = res.CreatedAt.Add(-time.Minute).Truncate(time.Second).UTC()
				res.UpdatedAt = res.UpdatedAt.Truncate(time.Second).UTC()

				a.dbReq.CreatedAt = a.dbReq.CreatedAt.Add(-time.Minute).Truncate(time.Second).UTC()
				a.dbReq.UpdatedAt = a.dbReq.UpdatedAt.Truncate(time.Second).UTC()
				if !cmp.Equal(res, a.dbReq) {
					t.Fatalf("%v", cmp.Diff(res, a.dbReq))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := NewWeightRepo(ctx)
			t.Cleanup(func() {
				_, cleanupErr := wr.db.
					NewTruncateTable().
					Table("actual_layers_by_client_id").
					Exec(ctx)
				if cleanupErr != nil {
					t.Fatalf("%v", cleanupErr.Error())
				}
			})

			tt.prepare(&wr, &tt.args)

			if err := wr.UpsertActualWeights(tt.args.ctx, tt.args.dbReq); (err != nil) != tt.wantErr {
				t.Errorf("UpsertActualWeights() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.check(&wr, &tt.args)
		})
	}
}

func Test_weightRepository_AppendHistoryWeights(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx   context.Context
		dbReq model.HistoryLayers
	}
	tests := []struct {
		name    string
		args    args
		prepare func(wr *weightRepository, a *args)
		check   func(wr *weightRepository, a *args)
		wantErr bool
	}{
		{
			name: "success insert",
			args: args{
				ctx: ctx,
				dbReq: model.HistoryLayers{
					ClientID:  uuid.New(),
					CreatedAt: time.Now(),
					Layer1:    nil,
					Layer2:    nil,
					Layer3:    nil,
					Layer4:    nil,
					Layer5:    nil,
				},
			},
			check: func(wr *weightRepository, a *args) {
				val, err := wr.db.NewSelect().
					Model(&model.HistoryLayers{}).
					Where("clientID = ?", a.dbReq.ClientID).
					Count(ctx)
				if err != nil {
					t.Fatalf("%v", err.Error())
				}
				assert.Equal(t, 1, val)
			},
			prepare: func(wr *weightRepository, a *args) {
				return
			},
		},
		{
			name: "success upsert",
			args: args{
				ctx: ctx,
				dbReq: model.HistoryLayers{
					ClientID:  uuid.New(),
					CreatedAt: time.Now(),
					Layer1:    nil,
					Layer2:    nil,
					Layer3:    nil,
					Layer4:    nil,
					Layer5:    nil,
				},
			},
			prepare: func(wr *weightRepository, a *args) {
				inserted := model.HistoryLayers{
					ClientID:  a.dbReq.ClientID,
					CreatedAt: time.Now().Add(-time.Hour),
					Layer1:    nil,
					Layer2:    nil,
					Layer3:    nil,
					Layer4:    nil,
					Layer5:    nil,
				}
				prepErr := wr.AppendHistoryWeights(ctx, inserted)
				if prepErr != nil {
					t.Fatalf("%v", prepErr.Error())
				}
			},
			check: func(wr *weightRepository, a *args) {
				val, err := wr.db.NewSelect().
					Model(&model.HistoryLayers{}).
					Where("clientID = ?", a.dbReq.ClientID).
					Count(ctx)
				if err != nil {
					t.Fatalf("%v", err.Error())
				}
				assert.Equal(t, 2, val)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := NewWeightRepo(ctx)
			t.Cleanup(func() {
				_, cleanupErr := wr.db.
					NewTruncateTable().
					Table("actual_layers_by_client_id").
					Exec(ctx)
				if cleanupErr != nil {
					t.Fatalf("%v", cleanupErr.Error())
				}
			})

			tt.prepare(&wr, &tt.args)

			if err := wr.AppendHistoryWeights(tt.args.ctx, tt.args.dbReq); (err != nil) != tt.wantErr {
				t.Errorf("UpsertActualWeights() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.check(&wr, &tt.args)
		})
	}
}
