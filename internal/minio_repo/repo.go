package minio_repo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (r *minioRepo) AcquireLock(ctx context.Context) (bool, error) {
	const lockKey = "locks/aggregator.lock"
	const ttl = 15 * time.Minute

	now := time.Now().UTC()

	// 1. Пытаемся посмотреть lock
	obj, err := r.minioClient.StatObject(
		ctx,
		r.bucketName,
		lockKey,
		minio.StatObjectOptions{},
	)

	if err == nil {
		// lock существует — проверяем TTL
		expStr := obj.UserMetadata["X-Amz-Meta-Expires_At"]
		if expStr == "" {
			return false, nil
		}

		exp, err := time.Parse(time.RFC3339, expStr)
		if err != nil {
			return false, nil
		}

		if now.Before(exp) {
			// lock ещё жив
			return false, nil
		}

		// lock протух — удаляем
		_ = r.minioClient.RemoveObject(
			ctx,
			r.bucketName,
			lockKey,
			minio.RemoveObjectOptions{},
		)
	}

	// 2. Пытаемся создать lock
	expiresAt := now.Add(ttl).Format(time.RFC3339)
	owner := uuid.NewString()

	opts := minio.PutObjectOptions{
		ContentType: "text/plain",
		UserMetadata: map[string]string{
			"owner":      owner,
			"expires_at": expiresAt,
		},
	}

	_, err = r.minioClient.PutObject(
		ctx,
		r.bucketName,
		lockKey,
		bytes.NewReader([]byte("lock")),
		4,
		opts,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *minioRepo) LoadLastAggregationTs(ctx context.Context) (time.Time, error) {
	const key = "meta/last_aggregation_ts"

	obj, err := r.minioClient.GetObject(ctx, r.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return time.Time{}, err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return time.Time{}, err
	}

	if len(data) == 0 {
		return time.Time{}, nil
	}

	ts, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}, err
	}

	return ts, nil
}

func (r *minioRepo) SaveLastAggregationTs(
	ctx context.Context,
	ts time.Time,
) error {
	const key = "meta/last_aggregation_ts"

	data := []byte(ts.UTC().Format(time.RFC3339))

	_, err := r.minioClient.PutObject(
		ctx,
		r.bucketName,
		key,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	return err
}

func (r *minioRepo) ListObjectAfter(
	ctx context.Context,
	from time.Time,
) ([]*WeightObject, error) {

	prefix := "weights/clients/"
	fromStr := from.UTC().Format("20060102T150405Z")

	var res []*WeightObject

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	for obj := range r.minioClient.ListObjects(ctx, r.bucketName, opts) {
		if obj.Err != nil {
			return nil, obj.Err
		}

		name := strings.TrimPrefix(obj.Key, prefix)
		if len(name) < len(fromStr) {
			continue
		}

		// лексикографическое сравнение
		if name[:len(fromStr)] <= fromStr {
			continue
		}

		wo, err := parseWeightObject(obj.Key)
		if err != nil {
			continue
		}

		data, err := r.readObject(ctx, obj.Key)
		if err != nil {
			continue
		}

		wo.Weights = data
		res = append(res, wo)
	}

	return res, nil
}

func (r *minioRepo) readObject(ctx context.Context, key string) ([]byte, error) {
	obj, err := r.minioClient.GetObject(ctx, r.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

func (r *minioRepo) SaveReleaseWeights(
	ctx context.Context,
	weights []byte,
) error {

	const key = "weights/global/latest.pt"

	_, err := r.minioClient.PutObject(
		ctx,
		r.bucketName,
		key,
		bytes.NewReader(weights),
		int64(len(weights)),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)

	return err
}

type WeightObject struct {
	Key         string
	ClientID    string
	Timestamp   time.Time
	NumExamples uint64
	Weights     []byte
}

func parseWeightObject(key string) (*WeightObject, error) {
	// weights/client_id=abc/20250901T120405Z_abc_128_uuid.pt

	parts := strings.Split(key, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid key: %s", key)
	}

	file := parts[len(parts)-1]
	fields := strings.Split(file, "_")
	if len(fields) < 4 {
		return nil, fmt.Errorf("invalid filename: %s", file)
	}

	ts, err := time.Parse("20060102T150405Z", fields[0])
	if err != nil {
		return nil, err
	}

	clientID := fields[1]

	numExamples, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, err
	}

	return &WeightObject{
		Key:         key,
		ClientID:    clientID,
		Timestamp:   ts,
		NumExamples: numExamples,
	}, nil
}
