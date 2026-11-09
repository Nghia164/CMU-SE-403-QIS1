// Package storage tạo presigned URL để client tải ảnh thẳng lên Cloudflare R2 (API tương thích S3),
// nên backend không phải nhận dữ liệu ảnh.
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Presigner trả URL PUT có thời hạn cho một object key. Kích thước và Content-Type được ký vào URL,
// nên client không thể tải lên file khác loại hoặc lớn hơn đã khai báo.
type Presigner interface {
	PresignPut(ctx context.Context, key, contentType string, size int64, ttl time.Duration) (url string, err error)
}

type R2 struct {
	client *s3.PresignClient
	bucket string
}

type R2Config struct {
	AccountID, AccessKeyID, SecretAccessKey, Bucket string
	Endpoint string // tuỳ chọn: ghi đè endpoint (MinIO khi test); rỗng = R2 theo AccountID
}

func NewR2(c R2Config) (*R2, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}
	cl := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.AccountID))
		if c.Endpoint != "" {
			o.BaseEndpoint = aws.String(c.Endpoint)
		}
		o.UsePathStyle = true
	})
	return &R2{client: s3.NewPresignClient(cl), bucket: c.Bucket}, nil
}

func (r *R2) PresignPut(ctx context.Context, key, contentType string, size int64, ttl time.Duration) (string, error) {
	req, err := r.client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(key),
		ContentType: aws.String(contentType), ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// Fake dùng khi chưa cấu hình R2 (dev/test): trả URL giả, không lưu gì.
type Fake struct{ BaseURL string }

func (f Fake) PresignPut(_ context.Context, key, _ string, _ int64, _ time.Duration) (string, error) {
	return f.BaseURL + "/" + key + "?fake=1", nil
}
