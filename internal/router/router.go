package router

import (
	"github.com/gin-gonic/gin"
	"github.com/thriftin/api/internal/handler"
	"github.com/thriftin/api/internal/middleware"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	Product      *handler.ProductHandler
	Category     *handler.CategoryHandler
	Cart         *handler.CartHandler
	Order        *handler.OrderHandler
	Wishlist     *handler.WishlistHandler
	Review       *handler.ReviewHandler
	Follow       *handler.FollowHandler
	Chat         *handler.ChatHandler
	Notification *handler.NotificationHandler
	Admin        *handler.AdminHandler
}

func Setup(r *gin.Engine, h *Handlers, jwtSecret string) {
	r.Use(middleware.CORS(), middleware.Logger())
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.GET("/me", middleware.Auth(jwtSecret), h.Auth.Me)
	}
	v1.GET("/products", h.Product.List)
	v1.GET("/products/:id", h.Product.Get)
	v1.GET("/categories", h.Category.List)
	v1.GET("/brands", h.Category.List)
	// reviews public
	v1.GET("/reviews/:productId", h.Review.List)

	authed := v1.Group("", middleware.Auth(jwtSecret))
	{
		authed.POST("/products", h.Product.Create)
		authed.PUT("/products/:id", h.Product.Update)
		authed.DELETE("/products/:id", h.Product.Delete)
		authed.POST("/products/:id/reserve", h.Product.Reserve)

		authed.POST("/categories", h.Category.Create)
		authed.POST("/brands", h.Category.Create)

		authed.GET("/cart", h.Cart.List)
		authed.POST("/cart/items", h.Cart.Add)

		authed.POST("/orders/checkout", h.Order.Checkout)
		authed.GET("/orders", h.Order.List)
		authed.GET("/orders/:id", h.Order.Get)

		authed.GET("/wishlist", h.Wishlist.List)
		authed.POST("/wishlist/:id", h.Wishlist.Toggle)

		authed.POST("/reviews", h.Review.Create)

		authed.POST("/follow", h.Follow.Follow)
		authed.DELETE("/follow/:id", h.Follow.Unfollow)

		authed.GET("/chat/rooms", h.Chat.Rooms)
		authed.POST("/chat/messages", h.Chat.Send)
		authed.GET("/chat/rooms/:id/messages", h.Chat.Messages)

		authed.GET("/notifications", h.Notification.List)
	}
	admin := v1.Group("/admin", middleware.Auth(jwtSecret), middleware.RequireRole("admin"))
	{
		admin.GET("/stats", h.Admin.Stats)
	}
}
