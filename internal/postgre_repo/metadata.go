package postgre_repo

import (
	"avg_weights_fed_ml_itmo/generated/metadata_db/public/model"
	"avg_weights_fed_ml_itmo/generated/metadata_db/public/table"
	"context"

	"github.com/go-jet/jet/v2/postgres"
)

func (db postgreDB) UpsertMetadataByClientID(ctx context.Context, metadata model.Metadata) error {
	tbl := table.Metadata
	stmt := tbl.INSERT(
		tbl.AllColumns.Except(tbl.CreatedAt, tbl.UpdatedAt),
	).VALUES(
		metadata.ClientID,
		metadata.NumExamples,
		metadata.ObjectKey,
	).ON_CONFLICT(
		tbl.ClientID,
	).DO_UPDATE(postgres.SET(
		tbl.NumExamples.SET(
			tbl.NumExamples.ADD(
				postgres.Int(int64(metadata.NumExamples)),
			),
		),
		tbl.ObjectKey.SET(postgres.String(metadata.ObjectKey)),
		tbl.UpdatedAt.SET(
			postgres.Timestamp(
				metadata.UpdatedAt.Year(),
				metadata.UpdatedAt.Month(),
				metadata.UpdatedAt.Day(),
				metadata.UpdatedAt.Hour(),
				metadata.UpdatedAt.Minute(),
				metadata.UpdatedAt.Second(),
			),
		),
	))

	sql, args := stmt.Sql()
	_, err := db.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
