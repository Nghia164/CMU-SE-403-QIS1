package ai

import (
	"context"
	"time"

	"masm/backend/internal/store"
)

// Runner nhận job, đẩy vào hàng đợi và để các worker xử lý bất đồng bộ.
type Runner struct {
	repo  store.Repo
	ai    Client
	delay time.Duration
	queue Queue
}

// NewRunner chạy `workers` worker cho tới khi ctx bị huỷ.
func NewRunner(ctx context.Context, repo store.Repo, c Client, q Queue, stepDelay time.Duration, workers int) *Runner {
	r := &Runner{repo: repo, ai: c, delay: stepDelay, queue: q}
	for i := 0; i < workers; i++ {
		go r.work(ctx)
	}
	return r
}

// Enqueue tạo job, trả ngay (không chờ AI).
func (r *Runner) Enqueue(sellerID string, in Input) (store.AIJob, error) {
	j := store.AIJob{
		ID: store.NewID("job_"), SellerID: sellerID, Status: "QUEUED", CreatedAt: time.Now(),
		Steps: []store.AIStep{
			{Label: "Nhận diện sản phẩm", Status: "pending"},
			{Label: "Tìm giá tham khảo", Status: "pending"},
			{Label: "Định giá", Status: "pending"},
			{Label: "Soạn nội dung tin", Status: "pending"},
		},
	}
	r.repo.SaveJob(j)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := r.queue.Push(ctx, task{j.ID, in}); err != nil {
		j.Status, j.Error = "FAILED", "Hàng đợi AI đang bận."
		r.repo.SaveJob(j)
		return j, err
	}
	return j, nil
}

func (r *Runner) work(ctx context.Context) {
	for {
		t, err := r.queue.Pop(ctx)
		if err != nil {
			return
		}
		r.run(t)
	}
}
func (r *Runner) run(t task) {
	j, err := r.repo.Job(t.JobID)
	if err != nil {
		return
	}
	j.Status = "RUNNING"
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	type out struct {
		res *store.AIResult
		err error
	}
	ch := make(chan out, 1)
	go func() { res, err := r.ai.Suggest(ctx, t.In); ch <- out{res, err} }()

	for i := range j.Steps {
		j.Steps[i].Status = "running"
		r.repo.SaveJob(j)
		time.Sleep(r.delay)
		j.Steps[i].Status = "done"
		j.Progress = (i + 1) * 100 / len(j.Steps)
		r.repo.SaveJob(j)
	}
	o := <-ch
	if o.err != nil {
		j.Status, j.Error = "FAILED", "Không tạo được gợi ý. Bạn có thể đăng tin thủ công."
		for i := range j.Steps {
			if j.Steps[i].Status != "done" {
				j.Steps[i].Status = "error"
			}
		}
	} else {
		j.Status, j.Result, j.Progress = "SUCCEEDED", o.res, 100
	}
	r.repo.SaveJob(j)
	if j.Status == "SUCCEEDED" {
		r.repo.AddNotification(store.Notification{UserID: j.SellerID, Kind: "ai_done", Title: "AI đã soạn xong gợi ý", Body: "Bạn xem lại và chỉnh sửa trước khi đăng.", Link: "/sell?job=" + j.ID})
	}
}
