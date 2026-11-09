// Package ai: Backend KHÔNG chạy AI nặng. Nó tạo job, trả job_id, rồi AI Service xử lý bất đồng bộ.
// AI Service không truy cập CSDL trực tiếp (chỉ qua Database Access Gateway).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"masm/backend/internal/store"
)

type Input struct {
	ImageKeys []string `json:"image_keys"`
	Hint      string   `json:"hint"`
	Category  string   `json:"category"`
	Condition string   `json:"condition"`
}

// Client là cổng gọi AI Service. Mọi gợi ý chỉ là đề xuất; người bán duyệt trước khi đăng.
type Client interface {
	Suggest(ctx context.Context, in Input) (*store.AIResult, error)
}

// HTTPClient gọi AI Service (Python) qua HTTP nội bộ.
type HTTPClient struct {
	BaseURL string
	HTTP    *http.Client
}

func NewHTTPClient(url string) *HTTPClient {
	return &HTTPClient{BaseURL: url, HTTP: &http.Client{Timeout: 60 * time.Second}}
}

func (c *HTTPClient) Suggest(ctx context.Context, in Input) (*store.AIResult, error) {
	body, _ := json.Marshal(in)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/suggest", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai service trả về %d", resp.StatusCode)
	}
	var out store.AIResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stub là AI giả lập để dev khi chưa có AI Service.
type Stub struct{}

func (Stub) Suggest(_ context.Context, in Input) (*store.AIResult, error) {
	name := in.Hint
	if name == "" {
		name = "Sản phẩm đã qua sử dụng"
	}
	return &store.AIResult{
		Titles: []string{name + ", tình trạng " + lower(in.Condition), name + " giá tốt, bán nhanh"},
		Descriptions: []string{
			"Sản phẩm đã qua sử dụng, hoạt động bình thường. Vui lòng xem kỹ ảnh và hỏi thêm trước khi mua.",
		},
		Hashtags: []string{"#dođu", "#thanhly"},
		Quick:    1300000, Optimal: 1450000, Confidence: 82,
		Basis: "Dựa trên 6 tin tương tự trong 30 ngày gần nhất (dữ liệu giả lập).",
		Sources: []store.PriceSource{
			{Source: "Chợ Tốt", Title: "Sản phẩm tương tự", Price: 1300000, Condition: "Tốt"},
			{Source: "Facebook", Title: "Sản phẩm tương tự", Price: 1600000, Condition: "Như mới"},
		},
	}, nil
}

func lower(s string) string {
	if s == "" {
		return "tốt"
	}
	b := []rune(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 'a' - 'A'
	}
	return string(b)
}
