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
	Report       *handler.ReportHandler
	Admin        *handler.AdminHandler
	Seller       *handler.SellerHandler
}
 
func Setup(r *gin.Engine, h *Handlers, jwtSecret string) {
	r.Use(middleware.CORS(), middleware.Logger())
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/google", h.Auth.Google)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.POST("/logout", h.Auth.Logout)
		auth.GET("/me", middleware.Auth(jwtSecret), h.Auth.Me)
	}
	v1.GET("/products", h.Product.List)
	v1.GET("/products/:id", h.Product.Get)
	v1.GET("/categories", h.Category.List)
	v1.GET("/brands", h.Category.ListBrands)
	v1.GET("/reviews/:productId", h.Review.List)
	v1.GET("/banned-words", h.Seller.BannedList)
	v1.POST("/webhooks/payment", h.Order.Webhook)

	authed := v1.Group("", middleware.Auth(jwtSecret))
	{
		authed.GET("/products/mine", h.Product.Mine)
		authed.POST("/products", h.Product.Create)
		authed.PUT("/products/:id", h.Product.Update)
		authed.DELETE("/products/:id", h.Product.Delete)
		authed.POST("/products/:id/images", h.Product.AddImage)
		authed.POST("/products/:id/reserve", h.Product.Reserve)

		authed.POST("/categories", h.Category.Create)
		authed.POST("/brands", h.Category.CreateBrand)

		authed.GET("/cart", h.Cart.List)
		authed.POST("/cart/items", h.Cart.Add)
		authed.DELETE("/cart/items/:id", h.Cart.Remove)

		authed.POST("/orders/checkout", h.Order.Checkout)
		authed.GET("/orders", h.Order.List)
		authed.GET("/orders/:id", h.Order.Get)
		authed.POST("/orders/:id/pay", h.Order.Pay)

		authed.GET("/wishlist", h.Wishlist.List)
		authed.POST("/wishlist/:id", h.Wishlist.Toggle)

		authed.POST("/reviews", h.Review.Create)

		authed.POST("/follow", h.Follow.Follow)
		authed.DELETE("/follow/:id", h.Follow.Unfollow)

		authed.GET("/chat/rooms", h.Chat.Rooms)
		authed.POST("/chat/rooms", h.Chat.CreateRoom)
		authed.POST("/chat/messages", h.Chat.Send)
		authed.GET("/chat/rooms/:id/messages", h.Chat.Messages)

		authed.GET("/notifications", h.Notification.List)
		authed.POST("/notifications/:id/read", h.Notification.MarkRead)

		authed.POST("/reports", h.Report.Create)

		authed.POST("/uploads", handler.Upload)

		authed.POST("/seller/apply", h.Seller.Apply)
		authed.GET("/seller/status", h.Seller.MyStatus)
	}
	admin := v1.Group("/admin", middleware.Auth(jwtSecret), middleware.RequireRole("admin"))
	{
		admin.GET("/stats", h.Admin.Stats)
		admin.GET("/reports", h.Report.List)
		admin.GET("/sellers/pending", h.Seller.ListPending)
		admin.POST("/sellers/:id/decide", h.Seller.Decide)
		admin.POST("/banned-words", h.Seller.BannedAdd)
		admin.DELETE("/banned-words/:word", h.Seller.BannedRemove)
	}
}
