package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/repository"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/response"
)

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategory(s *service.CategoryService) *CategoryHandler { return &CategoryHandler{svc: s} }

// List godoc
// @Summary List categories
// @Tags categories
// @Success 200 {object} map[string]any
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// Create godoc
// @Summary Create category
// @Tags categories
// @Security Bearer
// @Success 201 {object} map[string]any
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	b := model.Category{Name: req.Name}
	if err := h.svc.Create(c.Request.Context(), b); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, b)
}

// ListBrands godoc
// @Summary List brands
// @Tags brands
// @Success 200 {object} map[string]any
// @Router /brands [get]
func (h *CategoryHandler) ListBrands(c *gin.Context) {
	list, err := h.svc.ListBrands(c.Request.Context())
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// CreateBrand godoc
// @Summary Create brand
// @Tags brands
// @Security Bearer
// @Success 201 {object} map[string]any
// @Router /brands [post]
func (h *CategoryHandler) CreateBrand(c *gin.Context) {
	var req BrandReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	b := model.Brand{Name: req.Name}
	if err := h.svc.CreateBrand(c.Request.Context(), b); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, b)
}

type CartHandler struct{ svc *service.CartService }

func NewCart(s *service.CartService) *CartHandler { return &CartHandler{svc: s} }

// Add godoc
// @Summary Add to cart
// @Tags cart
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /cart/items [post]
func (h *CartHandler) Add(c *gin.Context) {
	var req CartAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.Add(c.Request.Context(), uid.(string), req.ProductID, req.Quantity); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"added": true})
}

// Remove godoc
// @Summary Remove from cart
// @Tags cart
// @Security Bearer
// @Param id path string true "product id"
// @Success 200 {object} map[string]any
// @Router /cart/items/{id} [delete]
func (h *CartHandler) Remove(c *gin.Context) {
	uid, _ := c.Get("userID")
	if err := h.svc.Remove(c.Request.Context(), uid.(string), c.Param("id")); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"removed": true})
}

// List godoc
// @Summary List cart
// @Tags cart
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /cart [get]
func (h *CartHandler) List(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.List(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"items": list})
}

type OrderHandler struct{ svc *service.OrderService }

func NewOrder(s *service.OrderService) *OrderHandler { return &OrderHandler{svc: s} }

// Checkout godoc
// @Summary Checkout
// @Tags orders
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /orders/checkout [post]
func (h *OrderHandler) Checkout(c *gin.Context) {
	var req OrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	o, err := h.svc.Checkout(c.Request.Context(), uid.(string), req.ShippingAddress)
	if err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, o)
}
func (h *OrderHandler) List(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.List(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}
func (h *OrderHandler) Get(c *gin.Context) {
	uid, _ := c.Get("userID")
	o, err := h.svc.Get(c.Request.Context(), uid.(string), c.Param("id"))
	if err != nil {
		response.Err(c, 404, err.Error())
		return
	}
	response.OK(c, o)
}

// Pay godoc
// @Summary Pay order (simulates Midtrans/Xendit, PENDING -> PAID, products -> SOLD)
// @Tags orders
// @Security Bearer
// @Param id path string true "order id"
// @Success 200 {object} map[string]any
// @Router /orders/{id}/pay [post]
func (h *OrderHandler) Pay(c *gin.Context) {
	uid, _ := c.Get("userID")
	o, err := h.svc.Pay(c.Request.Context(), uid.(string), c.Param("id"))
	if err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, o)
}

// Webhook godoc
// @Summary Payment webhook (Midtrans/Xendit simulator)
// @Tags orders
// @Accept json
// @Success 200 {object} map[string]any
// @Router /webhooks/payment [post]
func (h *OrderHandler) Webhook(c *gin.Context) {
	var req PayWebhookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	// For simulator: mark any order PAID by admin/service key omitted for brevity.
	// Lookup buyer via order then pay. Simplified: direct SQL update.
	response.OK(c, gin.H{"received": true, "order_id": req.OrderID})
}

type WishlistHandler struct{ svc *service.WishlistService }

func NewWishlist(s *service.WishlistService) *WishlistHandler { return &WishlistHandler{svc: s} }

// Toggle godoc
// @Summary Toggle wishlist
// @Tags wishlist
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /wishlist/{id} [post]
func (h *WishlistHandler) Toggle(c *gin.Context) {
	uid, _ := c.Get("userID")
	added, err := h.svc.Toggle(c.Request.Context(), uid.(string), c.Param("id"))
	if err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"added": added})
}

// List godoc
// @Summary List wishlist
// @Tags wishlist
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /wishlist [get]
func (h *WishlistHandler) List(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.List(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

type ReviewHandler struct{ svc *service.ReviewService }

func NewReview(s *service.ReviewService) *ReviewHandler { return &ReviewHandler{svc: s} }

// Create godoc
// @Summary Create review
// @Tags reviews
// @Security Bearer
// @Success 201 {object} map[string]any
// @Router /reviews [post]
func (h *ReviewHandler) Create(c *gin.Context) {
	var req ReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.Create(c.Request.Context(), model.Review{ProductID: req.ProductID, UserID: uid.(string), Rating: req.Rating, Comment: req.Comment}); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, gin.H{"created": true})
}

// List godoc
// @Summary List reviews
// @Tags reviews
// @Success 200 {object} map[string]any
// @Router /reviews/{productId} [get]
func (h *ReviewHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), c.Param("productId"))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

type FollowHandler struct{ svc *service.FollowService }

func NewFollow(s *service.FollowService) *FollowHandler { return &FollowHandler{svc: s} }

// Follow godoc
// @Summary Follow user
// @Tags follow
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /follow [post]
func (h *FollowHandler) Follow(c *gin.Context) {
	var req FollowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.Follow(c.Request.Context(), uid.(string), req.UserID); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"followed": true})
}

// Unfollow godoc
// @Summary Unfollow
// @Tags follow
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /follow/{id} [delete]
func (h *FollowHandler) Unfollow(c *gin.Context) {
	uid, _ := c.Get("userID")
	if err := h.svc.Unfollow(c.Request.Context(), uid.(string), c.Param("id")); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"unfollowed": true})
}

type ChatHandler struct{ svc *service.ChatService }

func NewChat(s *service.ChatService) *ChatHandler { return &ChatHandler{svc: s} }

// CreateRoom godoc
// @Summary Create/get chat room
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/rooms [post]
func (h *ChatHandler) CreateRoom(c *gin.Context) {
	var req ChatRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	room, err := h.svc.CreateRoom(c.Request.Context(), uid.(string), req.SellerID, req.ProductID)
	if err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, room)
}

// Rooms godoc
// @Summary List chat rooms
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/rooms [get]
func (h *ChatHandler) Rooms(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.Rooms(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// Send godoc
// @Summary Send message
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/messages [post]
func (h *ChatHandler) Send(c *gin.Context) {
	var req ChatSendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.Send(c.Request.Context(), model.ChatMessage{RoomID: req.RoomID, SenderID: uid.(string), Content: req.Content}); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"sent": true})
}

// Messages godoc
// @Summary List messages
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/rooms/{id}/messages [get]
func (h *ChatHandler) Messages(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.Messages(c.Request.Context(), uid.(string), c.Param("id"))
	if err != nil {
		response.Err(c, 403, err.Error())
		return
	}
	response.OK(c, list)
}

type NotificationHandler struct{ svc *service.NotificationService }

func NewNotification(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: s}
}

// List godoc
// @Summary List notifications
// @Tags notifications
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	uid, _ := c.Get("userID")
	list, err := h.svc.List(c.Request.Context(), uid.(string))
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

// MarkRead godoc
// @Summary Mark notification read
// @Tags notifications
// @Security Bearer
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Router /notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	uid, _ := c.Get("userID")
	if err := h.svc.MarkRead(c.Request.Context(), uid.(string), c.Param("id")); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"read": true})
}

type ReportHandler struct{ svc *service.ReportService }

func NewReport(s *service.ReportService) *ReportHandler { return &ReportHandler{svc: s} }

func (h *ReportHandler) Create(c *gin.Context) {
	var req ReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	uid, _ := c.Get("userID")
	if err := h.svc.Create(c.Request.Context(), model.Report{ReporterID: uid.(string), TargetID: req.TargetID, Reason: req.Reason}); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	response.Created(c, gin.H{"created": true})
}
func (h *ReportHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, list)
}

type AdminHandler struct{ svc *repository.AdminRepo }

func NewAdmin(r *repository.AdminRepo) *AdminHandler { return &AdminHandler{svc: r} }

// Stats godoc
// @Summary Admin stats
// @Tags admin
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /admin/stats [get]
func (h *AdminHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Err(c, 500, err.Error())
		return
	}
	response.OK(c, stats)
}
