package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/response"
)

type CategoryHandler struct{ svc *service.CategoryService }
func NewCategory(s *service.CategoryService) *CategoryHandler { return &CategoryHandler{svc:s} }
// List godoc
// @Summary List categories
// @Tags categories
// @Success 200 {object} map[string]any
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context){ list,_:=h.svc.List(c.Request.Context()); response.OK(c,list) }
// Create godoc
// @Summary Create category
// @Tags categories
// @Security Bearer
// @Success 201 {object} map[string]any
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context){ var b model.Category; c.ShouldBindJSON(&b); h.svc.Create(c.Request.Context(), b); response.Created(c,b) }

type CartHandler struct{ svc *service.CartService }
func NewCart(s *service.CartService) *CartHandler { return &CartHandler{svc:s} }
// Add godoc
// @Summary Add to cart
// @Tags cart
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /cart/items [post]
func (h *CartHandler) Add(c *gin.Context){ var req CartAddReq; c.ShouldBindJSON(&req); uid,_:=c.Get("userID"); h.svc.Add(c.Request.Context(), uid.(string), req.ProductID, req.Quantity); response.OK(c, gin.H{"added":true}) }
// List godoc
// @Summary List cart
// @Tags cart
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /cart [get]
func (h *CartHandler) List(c *gin.Context){ uid,_:=c.Get("userID"); list,_:=h.svc.List(c.Request.Context(), uid.(string)); response.OK(c,list) }

type OrderHandler struct{ svc *service.OrderService }
func NewOrder(s *service.OrderService) *OrderHandler { return &OrderHandler{svc:s} }
// Checkout godoc
// @Summary Checkout
// @Tags orders
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /orders/checkout [post]
func (h *OrderHandler) Checkout(c *gin.Context){ var req OrderReq; c.ShouldBindJSON(&req); uid,_:=c.Get("userID"); o,_:=h.svc.Checkout(c.Request.Context(), uid.(string), req.ShippingAddress); response.OK(c,o) }
func (h *OrderHandler) List(c *gin.Context){ response.OK(c, []any{}) }
func (h *OrderHandler) Get(c *gin.Context){ response.OK(c, gin.H{"id":c.Param("id")}) }

type WishlistHandler struct{ svc *service.WishlistService }
func NewWishlist(s *service.WishlistService) *WishlistHandler { return &WishlistHandler{svc:s} }
// Toggle godoc
// @Summary Toggle wishlist
// @Tags wishlist
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /wishlist/{id} [post]
func (h *WishlistHandler) Toggle(c *gin.Context){ uid,_:=c.Get("userID"); h.svc.Toggle(c.Request.Context(), uid.(string), c.Param("id")); response.OK(c, gin.H{"toggled":true}) }
// List godoc
// @Summary List wishlist
// @Tags wishlist
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /wishlist [get]
func (h *WishlistHandler) List(c *gin.Context){ uid,_:=c.Get("userID"); list,_:=h.svc.List(c.Request.Context(), uid.(string)); response.OK(c,list) }

type ReviewHandler struct{ svc *service.ReviewService }
func NewReview(s *service.ReviewService) *ReviewHandler { return &ReviewHandler{svc:s} }
// Create godoc
// @Summary Create review
// @Tags reviews
// @Security Bearer
// @Success 201 {object} map[string]any
// @Router /reviews [post]
func (h *ReviewHandler) Create(c *gin.Context){ var req ReviewReq; c.ShouldBindJSON(&req); uid,_:=c.Get("userID"); h.svc.Create(c.Request.Context(), model.Review{ProductID:req.ProductID, UserID: uid.(string), Rating:req.Rating, Comment:req.Comment}); response.Created(c, gin.H{"created":true}) }
// List godoc
// @Summary List reviews
// @Tags reviews
// @Success 200 {object} map[string]any
// @Router /reviews/{productId} [get]
func (h *ReviewHandler) List(c *gin.Context){ list,_:=h.svc.List(c.Request.Context(), c.Param("productId")); response.OK(c,list) }

type FollowHandler struct{ svc *service.FollowService }
func NewFollow(s *service.FollowService) *FollowHandler { return &FollowHandler{svc:s} }
// Follow godoc
// @Summary Follow user
// @Tags follow
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /follow [post]
func (h *FollowHandler) Follow(c *gin.Context){ var req FollowReq; c.ShouldBindJSON(&req); uid,_:=c.Get("userID"); h.svc.Follow(c.Request.Context(), uid.(string), req.UserID); response.OK(c, gin.H{"followed":true}) }
// Unfollow godoc
// @Summary Unfollow
// @Tags follow
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /follow/{id} [delete]
func (h *FollowHandler) Unfollow(c *gin.Context){ uid,_:=c.Get("userID"); h.svc.Unfollow(c.Request.Context(), uid.(string), c.Param("id")); response.OK(c, gin.H{"unfollowed":true}) }

type ChatHandler struct{ svc *service.ChatService }
func NewChat(s *service.ChatService) *ChatHandler { return &ChatHandler{svc:s} }
// Rooms godoc
// @Summary List chat rooms
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/rooms [get]
func (h *ChatHandler) Rooms(c *gin.Context){ uid,_:=c.Get("userID"); list,_:=h.svc.Rooms(c.Request.Context(), uid.(string)); response.OK(c,list) }
// Send godoc
// @Summary Send message
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/messages [post]
func (h *ChatHandler) Send(c *gin.Context){ var req ChatSendReq; c.ShouldBindJSON(&req); uid,_:=c.Get("userID"); h.svc.Send(c.Request.Context(), model.ChatMessage{RoomID:req.RoomID, SenderID: uid.(string), Content:req.Content}); response.OK(c, gin.H{"sent":true}) }
// Messages godoc
// @Summary List messages
// @Tags chat
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /chat/rooms/{id}/messages [get]
func (h *ChatHandler) Messages(c *gin.Context){ list,_:=h.svc.Messages(c.Request.Context(), c.Param("id")); response.OK(c,list) }

type NotificationHandler struct{ svc *service.NotificationService }
func NewNotification(s *service.NotificationService) *NotificationHandler { return &NotificationHandler{svc:s} }
// List godoc
// @Summary List notifications
// @Tags notifications
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /notifications [get]
func (h *NotificationHandler) List(c *gin.Context){ uid,_:=c.Get("userID"); list,_:=h.svc.List(c.Request.Context(), uid.(string)); response.OK(c,list) }

type AdminHandler struct{}
func NewAdmin() *AdminHandler { return &AdminHandler{} }
// Stats godoc
// @Summary Admin stats
// @Tags admin
// @Security Bearer
// @Success 200 {object} map[string]any
// @Router /admin/stats [get]
func (h *AdminHandler) Stats(c *gin.Context){ response.OK(c, gin.H{"users":0,"products":0}) }
