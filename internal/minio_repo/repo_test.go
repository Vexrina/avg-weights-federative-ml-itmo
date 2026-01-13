package minio_repo

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/require"
)

func TestPutNewWeights(t *testing.T) {
	ctx := context.Background()

	// Настройка тестового MinIO (можно использовать локальный контейнер или тестовый bucket)
	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	// Табличные кейсы
	tests := []struct {
		name       string
		numObjects int
		size       int // размер payload в байтах
	}{
		{"small_payload", 10, 128},
		{"medium_payload", 5, 1024},
		{"large_payload", 3, 10240},
		{"empty_payload", 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // запускаем тесты параллельно

			for i := 0; i < tt.numObjects; i++ {
				key := "test/" + uuid.NewString() + ".pt"
				data := bytes.Repeat([]byte{byte(i)}, tt.size)

				err := repo.PutNewWeights(ctx, key, data)
				require.NoError(t, err)

				// Проверяем, что объект реально появился
				obj, err := repo.minioClient.GetObject(ctx, repo.bucketName, "weights/clients/"+key, minio.GetObjectOptions{})
				require.NoError(t, err)
				defer obj.Close()

				readData, err := io.ReadAll(obj)
				require.NoError(t, err)
				require.Equal(t, data, readData)
				_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key, minio.RemoveObjectOptions{})
			}
		})
	}
}

func TestLoadLastAggregationTs(t *testing.T) {
	ctx := context.Background()

	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	const key = "meta/last_aggregation_ts"

	tests := []struct {
		name        string
		content     []byte
		expectError bool
		expectedTs  time.Time
	}{
		{
			name:        "valid timestamp",
			content:     []byte(time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)),
			expectError: false,
			expectedTs:  time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
		},
		{
			name:        "invalid timestamp",
			content:     []byte("not-a-timestamp"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Подготавливаем объект в MinIO
			if len(tt.content) > 0 || tt.name != "empty object" {
				_, _ = repo.minioClient.PutObject(
					ctx,
					repo.bucketName,
					key,
					bytes.NewReader(tt.content),
					int64(len(tt.content)),
					minio.PutObjectOptions{},
				)
			}

			ts, err := repo.LoadLastAggregationTs(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedTs, ts)
			}

			// Удаляем объект после теста, чтобы не мешал другим
			_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key, minio.RemoveObjectOptions{})
		})
	}
}

func TestSaveLastAggregationTs(t *testing.T) {
	ctx := context.Background()

	// Создаём тестовый репозиторий MinIO
	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	const key = "meta/last_aggregation_ts"

	testTs := time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC)

	// Сохраняем timestamp
	err := repo.SaveLastAggregationTs(ctx, testTs)
	require.NoError(t, err)

	// Проверяем, что объект реально появился в MinIO
	obj, err := repo.minioClient.GetObject(ctx, repo.bucketName, key, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer obj.Close()

	readData, err := io.ReadAll(obj)
	require.NoError(t, err)
	require.Equal(t, testTs.Format(time.RFC3339), string(readData))

	// Пробуем перезаписать другим значением
	newTs := testTs.Add(1 * time.Hour)
	err = repo.SaveLastAggregationTs(ctx, newTs)
	require.NoError(t, err)

	obj2, err := repo.minioClient.GetObject(ctx, repo.bucketName, key, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer obj2.Close()

	readData2, err := io.ReadAll(obj2)
	require.NoError(t, err)
	require.Equal(t, newTs.Format(time.RFC3339), string(readData2))

	// Чистим объект после теста
	_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key, minio.RemoveObjectOptions{})
}

func TestListObjectAfter(t *testing.T) {
	ctx := context.Background()

	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	prefix := "weights/clients/"

	fromTime := time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC)

	weights1 := []byte{1, 2, 3}
	ts1 := fromTime.Add(-1 * time.Minute) // до фильтра → не попадёт
	key1 := prefix + ts1.Format("20060102T150405Z") + "_clientA_10_" + uuid.NewString() + ".pt"
	_, _ = repo.minioClient.PutObject(ctx, repo.bucketName, key1, bytes.NewReader(weights1), int64(len(weights1)), minio.PutObjectOptions{})

	weights2 := []byte{4, 5, 6}
	ts2 := fromTime.Add(1 * time.Minute) // после фильтра → попадёт
	key2 := prefix + ts2.Format("20060102T150405Z") + "_clientB_20_" + uuid.NewString() + ".pt"
	_, _ = repo.minioClient.PutObject(ctx, repo.bucketName, key2, bytes.NewReader(weights2), int64(len(weights2)), minio.PutObjectOptions{})

	objs, err := repo.ListObjectAfter(ctx, fromTime)
	require.NoError(t, err)
	require.Len(t, objs, 1)

	wo := objs[0]
	require.Equal(t, "clientB", wo.ClientID)
	require.Equal(t, uint64(20), wo.NumExamples)
	require.Equal(t, weights2, wo.Weights)

	// Чистим после теста
	_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key1, minio.RemoveObjectOptions{})
	_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key2, minio.RemoveObjectOptions{})
}

func TestSaveReleaseWeights(t *testing.T) {
	ctx := context.Background()

	// Создаём MinIO репозиторий
	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	const key = "weights/global/latest.pt"

	weights := []byte{10, 20, 30, 40, 50}

	// Сохраняем глобальные веса
	err := repo.SaveReleaseWeights(ctx, weights)
	require.NoError(t, err)

	// Проверяем, что объект реально появился и данные совпадают
	obj, err := repo.minioClient.GetObject(ctx, repo.bucketName, key, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer obj.Close()

	readData, err := io.ReadAll(obj)
	require.NoError(t, err)
	require.Equal(t, weights, readData)

	// Пробуем перезаписать другими данными
	newWeights := []byte{1, 2, 3}
	err = repo.SaveReleaseWeights(ctx, newWeights)
	require.NoError(t, err)

	obj2, err := repo.minioClient.GetObject(ctx, repo.bucketName, key, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer obj2.Close()

	readData2, err := io.ReadAll(obj2)
	require.NoError(t, err)
	require.Equal(t, newWeights, readData2)

	// Чистим после теста
	_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key, minio.RemoveObjectOptions{})
}

func TestLoadLastWeight(t *testing.T) {
	ctx := context.Background()

	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	const key = "weights/global/latest.pt"

	tests := []struct {
		name        string
		content     []byte
		expectError bool
		expected    []byte
	}{
		{
			name:        "valid content",
			content:     []byte{0x01, 0x02, 0x03, 0x04}, // пример бинарных данных весов
			expectError: false,
			expected:    []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:        "empty object",
			content:     []byte{},
			expectError: false,
			expected:    []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Подготавливаем объект в MinIO, если content != nil
			if tt.content != nil {
				_, _ = repo.minioClient.PutObject(
					ctx,
					repo.bucketName,
					key,
					bytes.NewReader(tt.content),
					int64(len(tt.content)),
					minio.PutObjectOptions{},
				)
			}

			data, err := repo.LoadLastWeight(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, data)
			}

			// Удаляем объект после теста, чтобы не мешал другим
			_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key, minio.RemoveObjectOptions{})
		})
	}
}

func TestGetDownloadURL(t *testing.T) {
	ctx := context.Background()

	repo := NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	const key = "weights/global/latest.pt"

	tests := []struct {
		name        string
		content     []byte
		expectError bool
	}{
		{
			name:        "object exists",
			content:     []byte{0x01, 0x02, 0x03},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.content != nil {
				_, _ = repo.minioClient.PutObject(
					ctx,
					repo.bucketName,
					key,
					bytes.NewReader(tt.content),
					int64(len(tt.content)),
					minio.PutObjectOptions{},
				)
			}

			url, err := repo.GetDownloadURL(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, url)
				require.Contains(t, url, repo.bucketName) // базовая проверка, что ссылка корректная
			}

			// Удаляем объект после теста
			_ = repo.minioClient.RemoveObject(ctx, repo.bucketName, key, minio.RemoveObjectOptions{})
		})
	}
}
