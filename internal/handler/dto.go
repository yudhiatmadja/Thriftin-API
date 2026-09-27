package handler

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Username string `json:"username" binding:"required"`
	FullName string `json:"full_name"`
}
type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
} 
type RefreshReq struct{ RefreshToken string `json:"refresh_token" binding:"required"` }
type GoogleReq struct{ IDToken string `json:"id_token" binding:"required"` }
type LogoutReq struct{ RefreshToken string `json:"refresh_token" binding:"required"` }
type ProductReq struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Price       int64    `json:"price" binding:"required"`
	Condition   string   `json:"condition"`
	CategoryID  *string  `json:"category_id"`
	BrandID     *string  `json:"brand_id"`
	Size        *string  `json:"size"`
	Color       *string  `json:"color"`
	Material    *string  `json:"material"`
	Location    *string  `json:"location"`
	Images      []string `json:"images"`
}
type ProductUpdateReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	Status      *string `json:"status"`
}
type ProductImageReq struct{ URL string `json:"url" binding:"required"` }
type CartAddReq struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity"`
}
type OrderReq struct{ ShippingAddress string `json:"shipping_address" binding:"required"` }
type ReviewReq struct {
	ProductID string `json:"product_id" binding:"required"`
	Rating    int    `json:"rating" binding:"required,min=1,max=5"`
	Comment   string `json:"comment"`
}
type FollowReq struct{ UserID string `json:"user_id" binding:"required"` }
type ChatRoomReq struct {
	SellerID  string `json:"seller_id"`
	ProductID string `json:"product_id" binding:"required"`
}
type ChatSendReq struct {
	RoomID  string `json:"room_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}
type ReportReq struct {
	TargetID string `json:"target_id" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}
type CategoryReq struct{ Name string `json:"name" binding:"required"` }
type BrandReq struct{ Name string `json:"name" binding:"required"` }
type SellerApplyReq struct {
	StoreName    string `json:"store_name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Address      string `json:"address" binding:"required"`
	IDNumber     string `json:"id_number" binding:"required"`
	ProductTypes string `json:"product_types" binding:"required"`
	Description  string `json:"description"`
}
type SellerDecideReq struct {
	Status    string `json:"status" binding:"required"`
	AdminNote string `json:"admin_note"`
}
type BannedWordReq struct{ Word string `json:"word" binding:"required"` }
type PayWebhookReq struct {
	OrderID string `json:"order_id" binding:"required"`
	Status  string `json:"status"`
}
