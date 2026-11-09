package httpapi

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"masm/backend/internal/payment"
	"masm/backend/internal/store"
)

var phoneRe = regexp.MustCompile(`^(0|\+84)\d{9}$`)

type createOrderReq struct {
	ListingID    string `json:"listing_id"`
	Method       string `json:"method"` // vnpay | momo
	ReceiverName string `json:"receiver_name"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
}

func (s *Server) createOrder(c *gin.Context) {
	var r createOrderReq
	if c.ShouldBindJSON(&r) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	if r.Method != "vnpay" && r.Method != "momo" {
		fail(c, 422, "invalid_method", "Chọn VNPay hoặc MoMo.")
		return
	}
	r.ReceiverName, r.Address = strings.TrimSpace(r.ReceiverName), strings.TrimSpace(r.Address)
	if r.ReceiverName == "" || r.Address == "" {
		fail(c, 422, "invalid_receiver", "Nhập đầy đủ họ tên và địa chỉ nhận hàng.")
		return
	}
	if !phoneRe.MatchString(strings.ReplaceAll(r.Phone, " ", "")) {
		fail(c, 422, "invalid_phone", "Số điện thoại chưa đúng, ví dụ 0912345678.")
		return
	}
	l, err := s.repo.Listing(r.ListingID)
	if err != nil || l.Status != "active" {
		fail(c, 409, "listing_unavailable", "Tin này không còn để mua.")
		return
	}
	u := cur(c)
	if l.SellerID == u.ID {
		fail(c, 422, "own_listing", "Bạn không thể mua tin của chính mình.")
		return
	}
	o := s.repo.CreateOrder(store.Order{
		ListingID: l.ID, ListingTitle: l.Title, BuyerID: u.ID, SellerID: l.SellerID,
		Amount: l.Price, // giá luôn lấy từ server, không tin client
		Method: r.Method, Status: "pending_payment",
		ReceiverName: r.ReceiverName, Phone: strings.ReplaceAll(r.Phone, " ", ""), Address: r.Address,
		Events: map[string]time.Time{"placed": time.Now()},
	})
	payURL, err := s.pay.PayURL(c.Request.Context(), r.Method, o.ID, o.Amount, c.ClientIP())
	if err != nil {
		log.Printf("tạo link thanh toán %s cho %s: %v", r.Method, o.ID, err)
		o.Status = "cancelled"
		o.Events["cancelled"] = time.Now()
		_ = s.repo.UpdateOrder(o)
		fail(c, 503, "payment_unavailable", "Chưa kết nối được cổng thanh toán. Bạn thử lại sau hoặc chọn cổng khác.")
		return
	}
	c.JSON(201, gin.H{"order": o, "pay_url": payURL})
}

func (s *Server) listOrders(c *gin.Context) {
	out := s.repo.OrdersOfUser(cur(c).ID)
	c.JSON(200, gin.H{"items": out, "total": len(out)})
}

// ownOrder chỉ cho người mua, người bán của đơn hoặc admin xem; còn lại trả 404 để không lộ đơn tồn tại.
func (s *Server) ownOrder(c *gin.Context) (store.Order, bool) {
	o, err := s.repo.Order(c.Param("id"))
	u := cur(c)
	if err != nil || (o.BuyerID != u.ID && o.SellerID != u.ID && u.Role != store.RoleAdmin) {
		fail(c, 404, "not_found", "Không tìm thấy đơn hàng.")
		return o, false
	}
	return o, true
}

func (s *Server) getOrder(c *gin.Context) {
	if o, ok := s.ownOrder(c); ok {
		c.JSON(200, o)
	}
}

func (s *Server) shipOrder(c *gin.Context) {
	o, ok := s.ownOrder(c)
	if !ok {
		return
	}
	if o.SellerID != cur(c).ID || o.Status != "paid" {
		fail(c, 409, "invalid_state", "Chỉ người bán mới giao được đơn đã thanh toán.")
		return
	}
	o.Status = "shipping"
	o.Events["shipping"] = time.Now()
	_ = s.repo.UpdateOrder(o)
	s.notify(o.BuyerID, "order_shipped", "Đơn hàng đang được giao", o.ListingTitle, "/orders/"+o.ID)
	c.JSON(200, o)
}

func (s *Server) confirmReceived(c *gin.Context) {
	o, ok := s.ownOrder(c)
	if !ok {
		return
	}
	if o.BuyerID != cur(c).ID || o.Status != "shipping" {
		fail(c, 409, "invalid_state", "Chỉ người mua xác nhận được khi đơn đang giao.")
		return
	}
	o.Status = "delivered"
	o.Events["delivered"] = time.Now()
	_ = s.repo.UpdateOrder(o)
	s.notify(o.SellerID, "order_delivered", "Người mua đã nhận hàng", o.ListingTitle, "/orders/"+o.ID)
	c.JSON(200, o)
}

type settleResult int

const (
	settled settleResult = iota
	settleNotFound
	settleAmountMismatch
	settleAlreadyDone
)

// settle áp kết quả thanh toán đã xác minh chữ ký vào đơn: đối chiếu số tiền, idempotent.
func (s *Server) settle(provider string, r payment.Result) (store.Order, settleResult) {
	o, err := s.repo.Order(r.OrderID)
	if err != nil || o.Method != provider {
		return o, settleNotFound
	}
	if r.Amount != o.Amount {
		return o, settleAmountMismatch
	}
	if o.Status != "pending_payment" { // cổng gọi lại: không xử lý lần hai
		return o, settleAlreadyDone
	}
	if !r.Success {
		o.Status = "cancelled"
		o.Events["cancelled"] = time.Now()
		_ = s.repo.UpdateOrder(o)
		return o, settled
	}
	o.PaymentRef = r.TxnID
	l, err := s.repo.Listing(o.ListingID)
	if err != nil || l.Status != "active" { // đã có người khác mua trước: hoàn tiền
		o.Status = "refunded"
		o.Events["cancelled"] = time.Now()
	} else {
		o.Status = "paid"
		o.Events["paid"] = time.Now()
		// Ghi đơn trước: ràng buộc duy nhất của CSDL chặn hai đơn cùng thanh toán một tin.
		if err := s.repo.UpdateOrder(o); err != nil {
			o.Status = "refunded"
			delete(o.Events, "paid")
			o.Events["cancelled"] = time.Now()
		} else {
			l.Status = "sold"
			_ = s.repo.UpdateListing(l)
			s.notify(o.SellerID, "order_paid", "Có đơn mới đã thanh toán", o.ListingTitle, "/orders/"+o.ID)
			s.notify(o.BuyerID, "order_paid", "Thanh toán thành công", o.ListingTitle, "/orders/"+o.ID)
		}
	}
	_ = s.repo.UpdateOrder(o)
	return o, settled
}

// vnpayIPN: VNPay gọi GET kèm tham số vnp_*; phản hồi JSON {RspCode, Message} theo đặc tả của họ.
func (s *Server) vnpayIPN(c *gin.Context) {
	res, ok := s.pay.VerifyVNPay(c.Request.URL.Query())
	if !ok {
		c.JSON(200, gin.H{"RspCode": "97", "Message": "Invalid signature"})
		return
	}
	switch _, st := s.settle("vnpay", res); st {
	case settleNotFound:
		c.JSON(200, gin.H{"RspCode": "01", "Message": "Order not found"})
	case settleAmountMismatch:
		c.JSON(200, gin.H{"RspCode": "04", "Message": "Invalid amount"})
	case settleAlreadyDone:
		c.JSON(200, gin.H{"RspCode": "02", "Message": "Order already confirmed"})
	default:
		c.JSON(200, gin.H{"RspCode": "00", "Message": "Confirm Success"})
	}
}

// momoIPN: MoMo gọi POST JSON; trả 204 khi đã nhận.
func (s *Server) momoIPN(c *gin.Context) {
	var n payment.MoMoIPN
	if c.ShouldBindJSON(&n) != nil {
		fail(c, 400, "bad_request", "Dữ liệu không hợp lệ.")
		return
	}
	res, ok := s.pay.VerifyMoMo(n)
	if !ok {
		fail(c, 401, "bad_signature", "Chữ ký không hợp lệ.")
		return
	}
	switch _, st := s.settle("momo", res); st {
	case settleNotFound, settleAmountMismatch:
		fail(c, 422, "mismatch", "Đơn hàng hoặc số tiền không khớp.")
	default: // thành công hoặc đã xử lý trước đó
		c.Status(204)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }