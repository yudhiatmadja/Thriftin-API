package handler
type RegisterReq struct { Email string `json:"email" binding:"required,email"`; Password string `json:"password" binding:"required,min=6"`; Username string `json:"username" binding:"required"`; FullName string `json:"full_name"` }
type LoginReq struct { Email string `json:"email" binding:"required"`; Password string `json:"password" binding:"required"` }
type RefreshReq struct { RefreshToken string `json:"refresh_token" binding:"required"` }
type ProductReq struct { Title string `json:"title" binding:"required"`; Description string `json:"description"`; Price int64 `json:"price" binding:"required"`; Condition string `json:"condition"`; CategoryID *string `json:"category_id"`; BrandID *string `json:"brand_id"` }
type CartAddReq struct { ProductID string `json:"product_id" binding:"required"`; Quantity int `json:"quantity"` }
type OrderReq struct { ShippingAddress string `json:"shipping_address" binding:"required"` }
type ReviewReq struct { ProductID string `json:"product_id" binding:"required"`; Rating int `json:"rating" binding:"required,min=1,max=5"`; Comment string `json:"comment"` }
type FollowReq struct { UserID string `json:"user_id" binding:"required"` }
type ChatSendReq struct { RoomID string `json:"room_id" binding:"required"`; Content string `json:"content" binding:"required"` }
