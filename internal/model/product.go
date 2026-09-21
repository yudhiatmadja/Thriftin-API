package model

import "time"

type Product struct {
	ID          string    `json:"id"`
	SellerID    string    `json:"seller_id"`
	CategoryID  *string   `json:"category_id"`
	BrandID     *string   `json:"brand_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Condition   string    `json:"condition"`
	Size        *string   `json:"size"`
	Color       *string   `json:"color"`
	Material    *string   `json:"material"`
	Location    *string   `json:"location"`
	Status      string    `json:"status"`
	Views       int       `json:"views"`
	SearchVector string   `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}
type Category struct { ID string `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"` }
type Brand struct { ID string `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"` }

type Cart struct { ID string `json:"id"`; UserID string `json:"user_id"` }
type CartItem struct { ID string `json:"id"`; CartID string `json:"cart_id"`; ProductID string `json:"product_id"`; Quantity int `json:"quantity"` }
type Order struct { ID string `json:"id"`; BuyerID string `json:"buyer_id"`; Status string `json:"status"`; Total int64 `json:"total"`; ShippingAddress string `json:"shipping_address"`; CreatedAt time.Time `json:"created_at"` }
type OrderItem struct { ID string `json:"id"`; OrderID string `json:"order_id"`; ProductID string `json:"product_id"`; Price int64 `json:"price"` }
type Wishlist struct { ID string `json:"id"`; UserID string `json:"user_id"`; ProductID string `json:"product_id"` }
type Review struct { ID string `json:"id"`; ProductID string `json:"product_id"`; UserID string `json:"user_id"`; Rating int `json:"rating"`; Comment string `json:"comment"` }
type Follow struct { FollowerID string `json:"follower_id"`; FollowingID string `json:"following_id"` }
type ChatRoom struct { ID string `json:"id"`; BuyerID string `json:"buyer_id"`; SellerID string `json:"seller_id"`; ProductID string `json:"product_id"` }
type ChatMessage struct { ID string `json:"id"`; RoomID string `json:"room_id"`; SenderID string `json:"sender_id"`; Content string `json:"content"`; CreatedAt time.Time `json:"created_at"` }
type Notification struct { ID string `json:"id"`; UserID string `json:"user_id"`; Title string `json:"title"`; Body string `json:"body"`; IsRead bool `json:"is_read"`; CreatedAt time.Time `json:"created_at"` }
