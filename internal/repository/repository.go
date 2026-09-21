package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/thriftin/api/internal/model"
)

type Repos struct {
	User         *UserRepo
	Product      *ProductRepo
	Category     *CategoryRepo
	Cart         *CartRepo
	Order        *OrderRepo
	Wishlist     *WishlistRepo
	Review       *ReviewRepo
	Follow       *FollowRepo
	Chat         *ChatRepo
	Notification *NotificationRepo
	Pool         *pgxpool.Pool
	Redis        *redis.Client
}

func New(pool *pgxpool.Pool, rdb *redis.Client) *Repos {
	return &Repos{
		User: &UserRepo{pool: pool}, Product: &ProductRepo{pool: pool, rdb: rdb},
		Category: &CategoryRepo{pool: pool}, Cart: &CartRepo{pool: pool},
		Order: &OrderRepo{pool: pool}, Wishlist: &WishlistRepo{pool: pool},
		Review: &ReviewRepo{pool: pool}, Follow: &FollowRepo{pool: pool},
		Chat: &ChatRepo{pool: pool}, Notification: &NotificationRepo{pool: pool},
		Pool: pool, Redis: rdb,
	}
}

// --- User ---
type UserRepo struct{ pool *pgxpool.Pool }
func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO users(id,email,password_hash,username,full_name,role) VALUES($1,$2,$3,$4,$5,$6)`, u.ID, u.Email, u.PasswordHash, u.Username, u.FullName, u.Role)
	return err
}
func (r *UserRepo) ByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx, `SELECT id,email,password_hash,username,full_name,role,is_verified,is_active,created_at FROM users WHERE email=$1`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Username, &u.FullName, &u.Role, &u.IsVerified, &u.IsActive, &u.CreatedAt)
	return u, err
}
func (r *UserRepo) ByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx, `SELECT id,email,password_hash,username,full_name,role,is_verified,is_active,created_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Username, &u.FullName, &u.Role, &u.IsVerified, &u.IsActive, &u.CreatedAt)
	return u, err
}

// --- Product ---
type ProductRepo struct{ pool *pgxpool.Pool; rdb *redis.Client }
func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO products(id,seller_id,category_id,brand_id,title,description,price,condition,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, p.ID, p.SellerID, p.CategoryID, p.BrandID, p.Title, p.Description, p.Price, p.Condition, p.Status)
	return err
}
func (r *ProductRepo) Get(ctx context.Context, id string) (*model.Product, error) {
	p := &model.Product{}
	err := r.pool.QueryRow(ctx, `SELECT id,seller_id,category_id,brand_id,title,description,price,condition,status,views,created_at FROM products WHERE id=$1`, id).Scan(&p.ID, &p.SellerID, &p.CategoryID, &p.BrandID, &p.Title, &p.Description, &p.Price, &p.Condition, &p.Status, &p.Views, &p.CreatedAt)
	return p, err
}
func (r *ProductRepo) List(ctx context.Context, q string, limit, offset int) ([]model.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,seller_id,title,price,status,views,created_at FROM products WHERE status='ACTIVE' AND ($1='' OR search_vector @@ plainto_tsquery($1)) ORDER BY created_at DESC LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []model.Product
	for rows.Next() { var p model.Product; rows.Scan(&p.ID, &p.SellerID, &p.Title, &p.Price, &p.Status, &p.Views, &p.CreatedAt); out = append(out, p) }
	return out, nil
}
func (r *ProductRepo) ReserveTx(ctx context.Context, productID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM products WHERE id=$1 FOR UPDATE`, productID).Scan(&status); err != nil { return err }
	if status != "ACTIVE" { return err }
	_, err = tx.Exec(ctx, `UPDATE products SET status='RESERVED' WHERE id=$1`, productID)
	if err != nil { return err }
	return tx.Commit(ctx)
}

// --- Category / Brand ---
type CategoryRepo struct{ pool *pgxpool.Pool }
func (r *CategoryRepo) List(ctx context.Context) ([]model.Category, error) {
	rows, _ := r.pool.Query(ctx, `SELECT id,name,slug FROM categories`)
	defer rows.Close()
	var out []model.Category
	for rows.Next() { var c model.Category; rows.Scan(&c.ID, &c.Name, &c.Slug); out = append(out, c) }
	return out, nil
}
func (r *CategoryRepo) Create(ctx context.Context, c model.Category) error { _, err := r.pool.Exec(ctx, `INSERT INTO categories(id,name,slug) VALUES($1,$2,$3)`, c.ID, c.Name, c.Slug); return err }

// --- Cart ---
type CartRepo struct{ pool *pgxpool.Pool }
func (r *CartRepo) Add(ctx context.Context, userID, productID string, qty int) error {
	var cartID string
	err := r.pool.QueryRow(ctx, `SELECT id FROM carts WHERE user_id=$1`, userID).Scan(&cartID)
	if err != nil {
		_, err = r.pool.Exec(ctx, `INSERT INTO carts(id,user_id) VALUES(gen_random_uuid(),$1)`, userID)
		if err != nil { return err }
		r.pool.QueryRow(ctx, `SELECT id FROM carts WHERE user_id=$1`, userID).Scan(&cartID)
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO cart_items(id,cart_id,product_id,quantity) VALUES(gen_random_uuid(),$1,$2,$3) ON CONFLICT (cart_id,product_id) DO UPDATE SET quantity=EXCLUDED.quantity`, cartID, productID, qty)
	return err
}
func (r *CartRepo) List(ctx context.Context, userID string) ([]model.CartItem, error) {
	rows, _ := r.pool.Query(ctx, `SELECT ci.id,ci.cart_id,ci.product_id,ci.quantity FROM cart_items ci JOIN carts c ON c.id=ci.cart_id WHERE c.user_id=$1`, userID)
	defer rows.Close()
	var out []model.CartItem
	for rows.Next() { var ci model.CartItem; rows.Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.Quantity); out = append(out, ci) }
	return out, nil
}

// --- Order ---
type OrderRepo struct{ pool *pgxpool.Pool }
func (r *OrderRepo) Checkout(ctx context.Context, buyerID, addr string) (*model.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	var cartID string
	if err := tx.QueryRow(ctx, `SELECT id FROM carts WHERE user_id=$1`, buyerID).Scan(&cartID); err != nil { return nil, err }
	rows, err := tx.Query(ctx, `SELECT product_id FROM cart_items WHERE cart_id=$1`, cartID)
	if err != nil { return nil, err }
	var pids []string
	for rows.Next() { var pid string; rows.Scan(&pid); pids = append(pids, pid) }
	rows.Close()
	if len(pids)==0 { return nil, err }
	for _, pid := range pids {
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM products WHERE id=$1 FOR UPDATE`, pid).Scan(&status); err != nil { return nil, err }
		if status != "ACTIVE" { return nil, err }
		tx.Exec(ctx, `UPDATE products SET status='RESERVED' WHERE id=$1`, pid)
	}
	var total int64
	tx.QueryRow(ctx, `SELECT COALESCE(SUM(p.price),0) FROM cart_items ci JOIN products p ON p.id=ci.product_id WHERE ci.cart_id=$1`, cartID).Scan(&total)
	var oid string
	tx.QueryRow(ctx, `INSERT INTO orders(id,buyer_id,status,total,shipping_address) VALUES(gen_random_uuid(),$1,'PENDING_PAYMENT',$2,$3) RETURNING id`, buyerID, total, addr).Scan(&oid)
	for _, pid := range pids {
		var price int64
		tx.QueryRow(ctx, `SELECT price FROM products WHERE id=$1`, pid).Scan(&price)
		tx.Exec(ctx, `INSERT INTO order_items(id,order_id,product_id,price) VALUES(gen_random_uuid(),$1,$2,$3)`, oid, pid, price)
	}
	tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=$1`, cartID)
	_ = time.Now()
	if err := tx.Commit(ctx); err != nil { return nil, err }
	return &model.Order{ID: oid, BuyerID: buyerID, Status: "PENDING_PAYMENT", Total: total}, nil
}

// --- Wishlist / Review / Follow / Chat / Notification ---
type WishlistRepo struct{ pool *pgxpool.Pool }
func (r *WishlistRepo) Toggle(ctx context.Context, uid, pid string) error { _, err := r.pool.Exec(ctx, `INSERT INTO wishlists(id,user_id,product_id) VALUES(gen_random_uuid(),$1,$2) ON CONFLICT DO NOTHING`, uid, pid); return err }
func (r *WishlistRepo) List(ctx context.Context, uid string) ([]model.Wishlist, error) {
	rows, _ := r.pool.Query(ctx, `SELECT id,user_id,product_id FROM wishlists WHERE user_id=$1`, uid)
	defer rows.Close(); var out []model.Wishlist
	for rows.Next() { var w model.Wishlist; rows.Scan(&w.ID, &w.UserID, &w.ProductID); out=append(out,w) }
	return out,nil
}
type ReviewRepo struct{ pool *pgxpool.Pool }
func (r *ReviewRepo) Create(ctx context.Context, rev model.Review) error { _, err := r.pool.Exec(ctx, `INSERT INTO reviews(id,product_id,user_id,rating,comment) VALUES(gen_random_uuid(),$1,$2,$3,$4)`, rev.ProductID, rev.UserID, rev.Rating, rev.Comment); return err }
func (r *ReviewRepo) List(ctx context.Context, pid string) ([]model.Review, error) {
	rows,_:=r.pool.Query(ctx,`SELECT id,product_id,user_id,rating,comment FROM reviews WHERE product_id=$1`,pid)
	defer rows.Close(); var out []model.Review
	for rows.Next(){var rev model.Review; rows.Scan(&rev.ID,&rev.ProductID,&rev.UserID,&rev.Rating,&rev.Comment); out=append(out,rev)}
	return out,nil
}
type FollowRepo struct{ pool *pgxpool.Pool }
func (r *FollowRepo) Follow(ctx context.Context, a,b string) error { _,err:=r.pool.Exec(ctx,`INSERT INTO follows(follower_id,following_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,a,b); return err }
func (r *FollowRepo) Unfollow(ctx context.Context, a,b string) error { _,err:=r.pool.Exec(ctx,`DELETE FROM follows WHERE follower_id=$1 AND following_id=$2`,a,b); return err }
type ChatRepo struct{ pool *pgxpool.Pool }
func (r *ChatRepo) Rooms(ctx context.Context, uid string) ([]model.ChatRoom, error) {
	rows,_:=r.pool.Query(ctx,`SELECT id,buyer_id,seller_id,product_id FROM chat_rooms WHERE buyer_id=$1 OR seller_id=$1`,uid)
	defer rows.Close(); var out []model.ChatRoom
	for rows.Next(){var cr model.ChatRoom; rows.Scan(&cr.ID,&cr.BuyerID,&cr.SellerID,&cr.ProductID); out=append(out,cr)}
	return out,nil
}
func (r *ChatRepo) Send(ctx context.Context, m model.ChatMessage) error { _,err:=r.pool.Exec(ctx,`INSERT INTO chat_messages(id,room_id,sender_id,content) VALUES(gen_random_uuid(),$1,$2,$3)`,m.RoomID,m.SenderID,m.Content); return err }
func (r *ChatRepo) Messages(ctx context.Context, roomID string) ([]model.ChatMessage, error) {
	rows,_:=r.pool.Query(ctx,`SELECT id,room_id,sender_id,content,created_at FROM chat_messages WHERE room_id=$1 ORDER BY created_at`,roomID)
	defer rows.Close(); var out []model.ChatMessage
	for rows.Next(){var m model.ChatMessage; rows.Scan(&m.ID,&m.RoomID,&m.SenderID,&m.Content,&m.CreatedAt); out=append(out,m)}
	return out,nil
}
type NotificationRepo struct{ pool *pgxpool.Pool }
func (r *NotificationRepo) List(ctx context.Context, uid string) ([]model.Notification, error) {
	rows,_:=r.pool.Query(ctx,`SELECT id,user_id,title,body,is_read,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC`,uid)
	defer rows.Close(); var out []model.Notification
	for rows.Next(){var n model.Notification; rows.Scan(&n.ID,&n.UserID,&n.Title,&n.Body,&n.IsRead,&n.CreatedAt); out=append(out,n)}
	return out,nil
}
