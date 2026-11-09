package storage

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Chạy với MinIO (S3 tương thích như R2): đặt TEST_S3_ENDPOINT, TEST_S3_KEY, TEST_S3_SECRET, TEST_S3_BUCKET.
func TestPresignedPutWorksAgainstS3(t *testing.T) {
	ep := os.Getenv("TEST_S3_ENDPOINT")
	if ep == "" {
		t.Skip("đặt TEST_S3_ENDPOINT để chạy với MinIO")
	}
	r, err := NewR2(R2Config{AccountID: "x", AccessKeyID: os.Getenv("TEST_S3_KEY"), SecretAccessKey: os.Getenv("TEST_S3_SECRET"), Bucket: os.Getenv("TEST_S3_BUCKET"), Endpoint: ep})
	if err != nil {
		t.Fatal(err)
	}
	ensureBucket(t, ep, os.Getenv("TEST_S3_KEY"), os.Getenv("TEST_S3_SECRET"), os.Getenv("TEST_S3_BUCKET"))
	data := []byte("\xff\xd8\xff fake jpeg bytes")
	u, err := r.PresignPut(context.Background(), "uploads/u_1/test.jpg", "image/jpeg", int64(len(data)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	put := func(body []byte, ctype string) int {
		req, _ := http.NewRequest("PUT", u, bytes.NewReader(body))
		req.Header.Set("Content-Type", ctype)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}
	if c := put(data, "image/jpeg"); c != 200 {
		t.Fatalf("PUT hợp lệ: HTTP %d", c)
	}
	// Content-Type và kích thước đã được ký: tải file khác loại hoặc lớn hơn phải bị từ chối.
	if c := put(data, "text/html"); c == 200 {
		t.Fatal("khác Content-Type mà vẫn được PUT")
	}
	if c := put(append(data, make([]byte, 100)...), "image/jpeg"); c == 200 {
		t.Fatal("lớn hơn kích thước đã ký mà vẫn được PUT")
	}
}
func ensureBucket(t *testing.T, endpoint, key, secret, bucket string) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")))
	if err != nil {
		t.Fatal(err)
	}
	cl := s3.NewFromConfig(cfg, func(o *s3.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	if _, err := cl.CreateBucket(context.Background(), &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Logf("CreateBucket: %v (bỏ qua nếu bucket đã có)", err)
	}
}
