package model

import "time"

type Product struct {
	ID           string    `json:"id"`
	SellerID     string    `json:"seller_id"`
	CategoryID   *string   `json:"category_id"`
	BrandID      *string   `json:"brand_id"`
	CategoryName *string   `json:"category_name,omitempty"`
	BrandName    *string   `json:"brand_name,omitempty"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
	Condition    string    `json:"condition"`
	Size         *string   `json:"size,omitempty"`
	Color        *string   `json:"color,omitempty"`
	Material     *string   `json:"material,omitempty"`
	Location     *string   `json:"location,omitempty"`
	Status       string    `json:"status"`
	Views        int       `json:"views"`
	Images       []string  `json:"images,omitempty"`
	SearchVector string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
type Category struct{ ID string `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"` }
type Brand struct{ ID string `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"` }

type SellerApplication struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	UserEmail    *string `json:"user_email,omitempty"`
	Username     *string `json:"username,omitempty"`
	StoreName    string  `json:"store_name"`
	Phone        string  `json:"phone"`
	Address      string  `json:"address"`
	IDNumber     string  `json:"id_number"`
	ProductTypes string  `json:"product_types"`
	Description  *string `json:"description,omitempty"`
	Status       string  `json:"status"`
	AdminNote    *string `json:"admin_note,omitempty"`
}
type BannedWord struct{ Word string `json:"word"` }

type Cart struct{ ID string `json:"id"`; UserID string `json:"user_id"` }
type CartItem struct {
	ID        string `json:"id"`
	CartID    string `json:"cart_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	// joined
	Title  *string `json:"title,omitempty"`
	Price  *int64  `json:"price,omitempty"`
	Status *string `json:"status,omitempty"`
	Image  *string `json:"image,omitempty"`
}
type Order struct {
	ID              string      `json:"id"`
	BuyerID         string      `json:"buyer_id"`
	Status          string      `json:"status"`
	Total           int64       `json:"total"`
	ShippingAddress string      `json:"shipping_address"`
	Items           []OrderItem `json:"items,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
}
type OrderItem struct {
	ID        string `json:"id"`
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Price     int64  `json:"price"`
	Title     *string `json:"title,omitempty"`
}
type Wishlist struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Title     *string `json:"title,omitempty"`
	Price     *int64  `json:"price,omitempty"`
	Image     *string `json:"image,omitempty"`
}
type Review struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	UserID    string `json:"user_id"`
	Username  *string `json:"username,omitempty"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}
type Follow struct{ FollowerID string `json:"follower_id"`; FollowingID string `json:"following_id"` }
type ChatRoom struct {
	ID        string `json:"id"`
	BuyerID   string `json:"buyer_id"`
	SellerID  string `json:"seller_id"`
	ProductID string `json:"product_id"`
}
type ChatMessage struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
type Report struct {
	ID         string `json:"id"`
	ReporterID string `json:"reporter_id"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
}
type ProductFilter struct {
	Q          string
	CategoryID string
	BrandID    string
	Condition  string
	MinPrice   int64
	MaxPrice   int64
	Limit      int
	Offset     int
}
