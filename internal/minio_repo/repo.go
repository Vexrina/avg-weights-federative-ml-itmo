package minio_repo

import (
	"bytes"
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioRepo struct {
	minioClient *minio.Client
	bucketName  string
}

func NewMinioRepo(endpoint, accessKey, secretKey, bucketName string) *minioRepo {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // без tls
	})
	if err != nil {
		panic("cant connect to minio: " + err.Error())
	}

	return &minioRepo{minioClient: minioClient, bucketName: bucketName}
}

func (r *minioRepo) PutNewWeights(
	ctx context.Context,
	objectKey string,
	weights []byte,
) error {
	_, err := r.minioClient.PutObject(
		ctx,
		r.bucketName,
		objectKey,
		bytes.NewReader(weights),
		int64(len(weights)),
		minio.PutObjectOptions{},
	)
	if err != nil {
		log.Println("[ERROR] put weights error: " + err.Error())
		return err
	}
	return nil
}
