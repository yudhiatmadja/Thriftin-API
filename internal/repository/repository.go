package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	Report       *ReportRepo
	Admin        *AdminRepo
	Seller       *SellerRepo
	Banned       *BannedRepo
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
		Report: &ReportRepo{pool: pool}, Admin: &AdminRepo{pool: pool},
		Seller: &SellerRepo{pool: pool}, Banned: &BannedRepo{pool: pool},
		Pool: pool, Redis: rdb,
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// --- User + refresh tokens ---
type UserRepo struct{ pool *pgxpool.Pool }

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO users(id,email,password_hash,username,full_name,role) VALUES($1,$2,$3,$4,$5,$6)`, u.ID, u.Email, u.PasswordHash, u.Username, u.FullName, u.Role)
	return err
}
func (r *UserRepo) ByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx, `SELECT id,email,password_hash,username,full_name,role,is_verified,COALESCE(seller_verified,false),is_active,created_at FROM users WHERE email=$1`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Username, &u.FullName, &u.Role, &u.IsVerified, &u.SellerVerified, &u.IsActive, &u.CreatedAt)
	return u, err
}
func (r *UserRepo) UsernameTaken(ctx context.Context, username string) (bool, error) {
	var taken bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, username).Scan(&taken)
	return taken, err
}
func (r *UserRepo) ByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx, `SELECT id,email,password_hash,username,full_name,role,is_verified,COALESCE(seller_verified,false),is_active,created_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Username, &u.FullName, &u.Role, &u.IsVerified, &u.SellerVerified, &u.IsActive, &u.CreatedAt)
	return u, err
}
func (r *UserRepo) StoreRefresh(ctx context.Context, userID, token string, exp time.Time) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO refresh_tokens(id,user_id,token,expires_at) VALUES($1,$2,$3,$4)`, uuid.NewString(), userID, token, exp)
	return err
}
func (r *UserRepo) FindRefresh(ctx context.Context, token string) (string, error) {
	var uid string
	var exp time.Time
	var revoked bool
	err := r.pool.QueryRow(ctx, `SELECT user_id,expires_at,revoked FROM refresh_tokens WHERE token=$1`, token).Scan(&uid, &exp, &revoked)
	if err != nil {
		return "", err
	}
	if revoked || time.Now().After(exp) {
		return "", errors.New("refresh token expired or revoked")
	}
	return uid, nil
}
func (r *UserRepo) RevokeRefresh(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked=true WHERE token=$1`, token)
	return err
}

// --- Product ---
type ProductRepo struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO products(id,seller_id,category_id,brand_id,title,description,price,condition,size,color,material,location,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		p.ID, p.SellerID, p.CategoryID, p.BrandID, p.Title, p.Description, p.Price, p.Condition, p.Size, p.Color, p.Material, p.Location, p.Status)
	return err
}
func (r *ProductRepo) AddImage(ctx context.Context, productID, url string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO product_images(id,product_id,url) VALUES($1,$2,$3)`, uuid.NewString(), productID, url)
	return err
}
func (r *ProductRepo) images(ctx context.Context, pid string) []string {
	rows, err := r.pool.Query(ctx, `SELECT url FROM product_images WHERE product_id=$1`, pid)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u string
		rows.Scan(&u)
		out = append(out, u)
	}
	return out
}

func (r *ProductRepo) Get(ctx context.Context, id string) (*model.Product, error) {
	p := &model.Product{}
	err := r.pool.QueryRow(ctx, `SELECT p.id,p.seller_id,p.category_id,p.brand_id,c.name,br.name,p.title,p.description,p.price,p.condition,p.size,p.color,p.material,p.location,p.status,p.views,p.created_at
		FROM products p LEFT JOIN categories c ON c.id=p.category_id LEFT JOIN brands br ON br.id=p.brand_id WHERE p.id=$1`, id).
		Scan(&p.ID, &p.SellerID, &p.CategoryID, &p.BrandID, &p.CategoryName, &p.BrandName, &p.Title, &p.Description, &p.Price, &p.Condition, &p.Size, &p.Color, &p.Material, &p.Location, &p.Status, &p.Views, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.Images = r.images(ctx, p.ID)
	_, _ = r.pool.Exec(ctx, `UPDATE products SET views=views+1 WHERE id=$1`, id)
	return p, nil
}

func (r *ProductRepo) List(ctx context.Context, f model.ProductFilter) ([]model.Product, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	q := `SELECT p.id,p.seller_id,p.category_id,p.brand_id,c.name,br.name,p.title,p.description,p.price,p.condition,p.status,p.views,p.created_at
		FROM products p LEFT JOIN categories c ON c.id=p.category_id LEFT JOIN brands br ON br.id=p.brand_id
		WHERE p.status='ACTIVE'
		AND ($1='' OR p.search_vector @@ plainto_tsquery('simple',$1) OR p.title ILIKE '%'||$1||'%')
		AND ($2='' OR p.category_id::text=$2)
		AND ($3='' OR p.brand_id::text=$3)
		AND ($4='' OR p.condition=$4)
		AND ($5=0 OR p.price>=$5)
		AND ($6=0 OR p.price<=$6)
		ORDER BY p.created_at DESC LIMIT $7 OFFSET $8`
	rows, err := r.pool.Query(ctx, q, f.Q, f.CategoryID, f.BrandID, f.Condition, f.MinPrice, f.MaxPrice, f.Limit, f.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Product{}
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.SellerID, &p.CategoryID, &p.BrandID, &p.CategoryName, &p.BrandName, &p.Title, &p.Description, &p.Price, &p.Condition, &p.Status, &p.Views, &p.CreatedAt); err != nil {
			continue
		}
		p.Images = r.images(ctx, p.ID)
		out = append(out, p)
	}
	if out == nil {
		out = []model.Product{}
	}
	return out, nil
}

func (r *ProductRepo) MyProducts(ctx context.Context, sellerID string) ([]model.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,title,price,status,views,created_at FROM products WHERE seller_id=$1 ORDER BY created_at DESC`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Product{}
	for rows.Next() {
		var p model.Product
		rows.Scan(&p.ID, &p.Title, &p.Price, &p.Status, &p.Views, &p.CreatedAt)
		p.SellerID = sellerID
		p.Images = r.images(ctx, p.ID)
		out = append(out, p)
	}
	if out == nil {
		out = []model.Product{}
	}
	return out, nil
}

func (r *ProductRepo) Update(ctx context.Context, id, sellerID string, title, desc *string, price *int64, status *string) error {
	var owner string
	if err := r.pool.QueryRow(ctx, `SELECT seller_id FROM products WHERE id=$1`, id).Scan(&owner); err != nil {
		return err
	}
	if owner != sellerID {
		return errors.New("forbidden: not your product")
	}
	_, err := r.pool.Exec(ctx, `UPDATE products SET title=COALESCE($2,title), description=COALESCE($3,description), price=COALESCE($4,price), status=COALESCE($5,status) WHERE id=$1`, id, title, desc, price, status)
	return err
}
func (r *ProductRepo) Delete(ctx context.Context, id, sellerID, role string) error {
	var owner string
	if err := r.pool.QueryRow(ctx, `SELECT seller_id FROM products WHERE id=$1`, id).Scan(&owner); err != nil {
		return err
	}
	if owner != sellerID && role != "admin" {
		return errors.New("forbidden: not your product")
	}
	_, err := r.pool.Exec(ctx, `UPDATE products SET status='ARCHIVED' WHERE id=$1`, id)
	return err
}
func (r *ProductRepo) ReserveTx(ctx context.Context, productID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM products WHERE id=$1 FOR UPDATE`, productID).Scan(&status); err != nil {
		return err
	}
	if status != "ACTIVE" {
		return errors.New("product not available")
	}
	_, err = tx.Exec(ctx, `UPDATE products SET status='RESERVED' WHERE id=$1`, productID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// --- Category / Brand ---
type CategoryRepo struct{ pool *pgxpool.Pool }

func (r *CategoryRepo) List(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Category{}
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Slug)
		out = append(out, c)
	}
	if out == nil {
		out = []model.Category{}
	}
	return out, nil
}
func (r *CategoryRepo) Create(ctx context.Context, c model.Category) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.Slug == "" {
		c.Slug = slugify(c.Name)
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO categories(id,name,slug) VALUES($1,$2,$3)`, c.ID, c.Name, c.Slug)
	return err
}
func (r *CategoryRepo) ListBrands(ctx context.Context) ([]model.Brand, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug FROM brands ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Brand{}
	for rows.Next() {
		var b model.Brand
		rows.Scan(&b.ID, &b.Name, &b.Slug)
		out = append(out, b)
	}
	if out == nil {
		out = []model.Brand{}
	}
	return out, nil
}
func (r *CategoryRepo) CreateBrand(ctx context.Context, b model.Brand) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	if b.Slug == "" {
		b.Slug = slugify(b.Name)
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO brands(id,name,slug) VALUES($1,$2,$3)`, b.ID, b.Name, b.Slug)
	return err
}

// --- Cart ---
type CartRepo struct{ pool *pgxpool.Pool }

func (r *CartRepo) ensureCart(ctx context.Context, userID string) (string, error) {
	var cartID string
	err := r.pool.QueryRow(ctx, `SELECT id FROM carts WHERE user_id=$1`, userID).Scan(&cartID)
	if err == nil {
		return cartID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	cartID = uuid.NewString()
	_, err = r.pool.Exec(ctx, `INSERT INTO carts(id,user_id) VALUES($1,$2)`, cartID, userID)
	return cartID, err
}
func (r *CartRepo) Add(ctx context.Context, userID, productID string, qty int) error {
	if qty <= 0 {
		qty = 1
	}
	var status string
	if err := r.pool.QueryRow(ctx, `SELECT status FROM products WHERE id=$1`, productID).Scan(&status); err != nil {
		return errors.New("product not found")
	}
	if status != "ACTIVE" {
		return errors.New("product not available")
	}
	cartID, err := r.ensureCart(ctx, userID)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO cart_items(id,cart_id,product_id,quantity) VALUES($1,$2,$3,$4)
		ON CONFLICT (cart_id,product_id) DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity`, uuid.NewString(), cartID, productID, qty)
	return err
}
func (r *CartRepo) Remove(ctx context.Context, userID, productID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=(SELECT id FROM carts WHERE user_id=$1) AND product_id=$2`, userID, productID)
	return err
}
func (r *CartRepo) List(ctx context.Context, userID string) ([]model.CartItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT ci.id,ci.cart_id,ci.product_id,ci.quantity,p.title,p.price,p.status,
		(SELECT url FROM product_images WHERE product_id=p.id LIMIT 1)
		FROM cart_items ci JOIN carts c ON c.id=ci.cart_id JOIN products p ON p.id=ci.product_id WHERE c.user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.CartItem{}
	for rows.Next() {
		var ci model.CartItem
		rows.Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.Quantity, &ci.Title, &ci.Price, &ci.Status, &ci.Image)
		out = append(out, ci)
	}
	if out == nil {
		out = []model.CartItem{}
	}
	return out, nil
}

// --- Order ---
type OrderRepo struct{ pool *pgxpool.Pool }

func (r *OrderRepo) Checkout(ctx context.Context, buyerID, addr string) (*model.Order, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("shipping_address required")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var cartID string
	if err := tx.QueryRow(ctx, `SELECT id FROM carts WHERE user_id=$1`, buyerID).Scan(&cartID); err != nil {
		return nil, errors.New("cart empty")
	}
	rows, err := tx.Query(ctx, `SELECT ci.product_id, ci.quantity FROM cart_items ci WHERE ci.cart_id=$1`, cartID)
	if err != nil {
		return nil, err
	}
	type item struct {
		pid string
		qty int
	}
	var items []item
	for rows.Next() {
		var it item
		rows.Scan(&it.pid, &it.qty)
		items = append(items, it)
	}
	rows.Close()
	if len(items) == 0 {
		return nil, errors.New("cart empty")
	}
	for _, it := range items {
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM products WHERE id=$1 FOR UPDATE`, it.pid).Scan(&status); err != nil {
			return nil, errors.New("product not found: " + it.pid)
		}
		if status != "ACTIVE" {
			return nil, errors.New("product not available: " + it.pid)
		}
	}
	var total int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(p.price*ci.quantity),0) FROM cart_items ci JOIN products p ON p.id=ci.product_id WHERE ci.cart_id=$1`, cartID).Scan(&total); err != nil {
		return nil, err
	}
	var oid string
	if err := tx.QueryRow(ctx, `INSERT INTO orders(id,buyer_id,status,total,shipping_address) VALUES(gen_random_uuid(),$1,'PENDING_PAYMENT',$2,$3) RETURNING id`, buyerID, total, addr).Scan(&oid); err != nil {
		return nil, err
	}
	for _, it := range items {
		var price int64
		tx.QueryRow(ctx, `SELECT price FROM products WHERE id=$1`, it.pid).Scan(&price)
		if _, err := tx.Exec(ctx, `INSERT INTO order_items(id,order_id,product_id,price) VALUES(gen_random_uuid(),$1,$2,$3)`, oid, it.pid, price*int64(it.qty)); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET status='RESERVED' WHERE id=$1`, it.pid); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=$1`, cartID); err != nil {
		return nil, err
	}
	// notify buyer
	_, _ = tx.Exec(ctx, `INSERT INTO notifications(id,user_id,title,body) VALUES(gen_random_uuid(),$1,'Order created','Your order is pending payment')`, buyerID)
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &model.Order{ID: oid, BuyerID: buyerID, Status: "PENDING_PAYMENT", Total: total, ShippingAddress: addr}, nil
}
func (r *OrderRepo) List(ctx context.Context, buyerID string) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,buyer_id,status,total,shipping_address,created_at FROM orders WHERE buyer_id=$1 ORDER BY created_at DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Order{}
	for rows.Next() {
		var o model.Order
		rows.Scan(&o.ID, &o.BuyerID, &o.Status, &o.Total, &o.ShippingAddress, &o.CreatedAt)
		out = append(out, o)
	}
	if out == nil {
		out = []model.Order{}
	}
	return out, nil
}
func (r *OrderRepo) Get(ctx context.Context, buyerID, orderID string) (*model.Order, error) {
	o := &model.Order{}
	err := r.pool.QueryRow(ctx, `SELECT id,buyer_id,status,total,shipping_address,created_at FROM orders WHERE id=$1 AND buyer_id=$2`, orderID, buyerID).Scan(&o.ID, &o.BuyerID, &o.Status, &o.Total, &o.ShippingAddress, &o.CreatedAt)
	if err != nil {
		return nil, errors.New("order not found")
	}
	rows, _ := r.pool.Query(ctx, `SELECT oi.id,oi.order_id,oi.product_id,oi.price,p.title FROM order_items oi LEFT JOIN products p ON p.id=oi.product_id WHERE oi.order_id=$1`, orderID)
	defer rows.Close()
	for rows.Next() {
		var it model.OrderItem
		rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Price, &it.Title)
		o.Items = append(o.Items, it)
	}
	return o, nil
}
// Pay simulates Midtrans/Xendit: PENDING_PAYMENT -> PAID, products RESERVED -> SOLD
func (r *OrderRepo) Pay(ctx context.Context, buyerID, orderID string) (*model.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM orders WHERE id=$1 AND buyer_id=$2`, orderID, buyerID).Scan(&status); err != nil {
		return nil, errors.New("order not found")
	}
	if status != "PENDING_PAYMENT" {
		return nil, errors.New("order cannot be paid, status: " + status)
	}
	if _, err := tx.Exec(ctx, `UPDATE orders SET status='PAID' WHERE id=$1`, orderID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE products SET status='SOLD' WHERE id IN (SELECT product_id FROM order_items WHERE order_id=$1)`, orderID); err != nil {
		return nil, err
	}
	_, _ = tx.Exec(ctx, `INSERT INTO notifications(id,user_id,title,body) VALUES(gen_random_uuid(),$1,'Payment success','Order paid, seller is preparing shipment')`, buyerID)
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.Get(ctx, buyerID, orderID)
}

// --- Wishlist ---
type WishlistRepo struct{ pool *pgxpool.Pool }

func (r *WishlistRepo) Toggle(ctx context.Context, uid, pid string) (bool, error) {
	var exists bool
	_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id=$1 AND product_id=$2)`, uid, pid).Scan(&exists)
	if exists {
		_, err := r.pool.Exec(ctx, `DELETE FROM wishlists WHERE user_id=$1 AND product_id=$2`, uid, pid)
		return false, err
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO wishlists(id,user_id,product_id) VALUES(gen_random_uuid(),$1,$2)`, uid, pid)
	return true, err
}
func (r *WishlistRepo) List(ctx context.Context, uid string) ([]model.Wishlist, error) {
	rows, err := r.pool.Query(ctx, `SELECT w.id,w.user_id,w.product_id,p.title,p.price,(SELECT url FROM product_images WHERE product_id=p.id LIMIT 1)
		FROM wishlists w LEFT JOIN products p ON p.id=w.product_id WHERE w.user_id=$1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Wishlist{}
	for rows.Next() {
		var w model.Wishlist
		rows.Scan(&w.ID, &w.UserID, &w.ProductID, &w.Title, &w.Price, &w.Image)
		out = append(out, w)
	}
	if out == nil {
		out = []model.Wishlist{}
	}
	return out, nil
}

// --- Review ---
type ReviewRepo struct{ pool *pgxpool.Pool }

func (r *ReviewRepo) Create(ctx context.Context, rev model.Review) error {
	rev.ID = uuid.NewString()
	_, err := r.pool.Exec(ctx, `INSERT INTO reviews(id,product_id,user_id,rating,comment) VALUES($1,$2,$3,$4,$5)`, rev.ID, rev.ProductID, rev.UserID, rev.Rating, rev.Comment)
	return err
}
func (r *ReviewRepo) List(ctx context.Context, pid string) ([]model.Review, error) {
	rows, err := r.pool.Query(ctx, `SELECT r.id,r.product_id,r.user_id,u.username,r.rating,r.comment FROM reviews r LEFT JOIN users u ON u.id=r.user_id WHERE r.product_id=$1 ORDER BY r.created_at DESC`, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Review{}
	for rows.Next() {
		var rev model.Review
		rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Username, &rev.Rating, &rev.Comment)
		out = append(out, rev)
	}
	if out == nil {
		out = []model.Review{}
	}
	return out, nil
}

// --- Follow ---
type FollowRepo struct{ pool *pgxpool.Pool }

func (r *FollowRepo) Follow(ctx context.Context, a, b string) error {
	if a == b {
		return errors.New("cannot follow yourself")
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO follows(follower_id,following_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, a, b)
	return err
}
func (r *FollowRepo) Unfollow(ctx context.Context, a, b string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM follows WHERE follower_id=$1 AND following_id=$2`, a, b)
	return err
}

// --- Chat ---
type ChatRepo struct{ pool *pgxpool.Pool }

func (r *ChatRepo) CreateRoom(ctx context.Context, buyerID, sellerID, productID string) (*model.ChatRoom, error) {
	var room model.ChatRoom
	err := r.pool.QueryRow(ctx, `SELECT id,buyer_id,seller_id,product_id FROM chat_rooms WHERE buyer_id=$1 AND seller_id=$2 AND product_id=$3`, buyerID, sellerID, productID).Scan(&room.ID, &room.BuyerID, &room.SellerID, &room.ProductID)
	if err == nil {
		return &room, nil
	}
	room = model.ChatRoom{ID: uuid.NewString(), BuyerID: buyerID, SellerID: sellerID, ProductID: productID}
	_, err = r.pool.Exec(ctx, `INSERT INTO chat_rooms(id,buyer_id,seller_id,product_id) VALUES($1,$2,$3,$4)`, room.ID, room.BuyerID, room.SellerID, room.ProductID)
	return &room, err
}
func (r *ChatRepo) Rooms(ctx context.Context, uid string) ([]model.ChatRoom, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,buyer_id,seller_id,product_id FROM chat_rooms WHERE buyer_id=$1 OR seller_id=$1 ORDER BY id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ChatRoom{}
	for rows.Next() {
		var cr model.ChatRoom
		rows.Scan(&cr.ID, &cr.BuyerID, &cr.SellerID, &cr.ProductID)
		out = append(out, cr)
	}
	if out == nil {
		out = []model.ChatRoom{}
	}
	return out, nil
}
func (r *ChatRepo) isMember(ctx context.Context, roomID, uid string) bool {
	var ok bool
	_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM chat_rooms WHERE id=$1 AND (buyer_id=$2 OR seller_id=$2))`, roomID, uid).Scan(&ok)
	return ok
}
func (r *ChatRepo) Send(ctx context.Context, m model.ChatMessage) error {
	if !r.isMember(ctx, m.RoomID, m.SenderID) {
		return errors.New("forbidden: not room member")
	}
	m.ID = uuid.NewString()
	_, err := r.pool.Exec(ctx, `INSERT INTO chat_messages(id,room_id,sender_id,content) VALUES($1,$2,$3,$4)`, m.ID, m.RoomID, m.SenderID, m.Content)
	return err
}
func (r *ChatRepo) Messages(ctx context.Context, uid, roomID string) ([]model.ChatMessage, error) {
	if !r.isMember(ctx, roomID, uid) {
		return nil, errors.New("forbidden: not room member")
	}
	rows, err := r.pool.Query(ctx, `SELECT id,room_id,sender_id,content,created_at FROM chat_messages WHERE room_id=$1 ORDER BY created_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ChatMessage{}
	for rows.Next() {
		var m model.ChatMessage
		rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &m.Content, &m.CreatedAt)
		out = append(out, m)
	}
	if out == nil {
		out = []model.ChatMessage{}
	}
	return out, nil
}

// --- Notification ---
type NotificationRepo struct{ pool *pgxpool.Pool }

func (r *NotificationRepo) Create(ctx context.Context, n model.Notification) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO notifications(id,user_id,title,body) VALUES($1,$2,$3,$4)`, uuid.NewString(), n.UserID, n.Title, n.Body)
	return err
}
func (r *NotificationRepo) List(ctx context.Context, uid string) ([]model.Notification, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,title,body,is_read,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt)
		out = append(out, n)
	}
	if out == nil {
		out = []model.Notification{}
	}
	return out, nil
}
func (r *NotificationRepo) MarkRead(ctx context.Context, uid, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET is_read=true WHERE id=$1 AND user_id=$2`, id, uid)
	return err
}

// --- Report ---
type ReportRepo struct{ pool *pgxpool.Pool }

func (r *ReportRepo) Create(ctx context.Context, rep model.Report) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO reports(id,reporter_id,target_id,reason) VALUES($1,$2,$3,$4)`, uuid.NewString(), rep.ReporterID, rep.TargetID, rep.Reason)
	return err
}
func (r *ReportRepo) List(ctx context.Context) ([]model.Report, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,reporter_id,target_id,reason FROM reports ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Report{}
	for rows.Next() {
		var x model.Report
		rows.Scan(&x.ID, &x.ReporterID, &x.TargetID, &x.Reason)
		out = append(out, x)
	}
	if out == nil {
		out = []model.Report{}
	}
	return out, nil
}

// --- Seller verification ---
type SellerRepo struct{ pool *pgxpool.Pool }

func (r *SellerRepo) Apply(ctx context.Context, a model.SellerApplication) error {
	a.ID = uuid.NewString()
	_, err := r.pool.Exec(ctx, `INSERT INTO seller_applications(id,user_id,store_name,phone,address,id_number,product_types,description) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.UserID, a.StoreName, a.Phone, a.Address, a.IDNumber, a.ProductTypes, a.Description)
	return err
}
func (r *SellerRepo) MyStatus(ctx context.Context, uid string) (*model.SellerApplication, error) {
	a := &model.SellerApplication{}
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,store_name,phone,address,id_number,product_types,description,status,admin_note FROM seller_applications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1`, uid).
		Scan(&a.ID, &a.UserID, &a.StoreName, &a.Phone, &a.Address, &a.IDNumber, &a.ProductTypes, &a.Description, &a.Status, &a.AdminNote)
	if err != nil {
		return nil, err
	}
	return a, nil
}
func (r *SellerRepo) IsVerified(ctx context.Context, uid string) bool {
	var v bool
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(seller_verified,false) FROM users WHERE id=$1`, uid).Scan(&v)
	return v
}
func (r *SellerRepo) ListPending(ctx context.Context) ([]model.SellerApplication, error) {
	rows, err := r.pool.Query(ctx, `SELECT a.id,a.user_id,u.email,u.username,a.store_name,a.phone,a.address,a.id_number,a.product_types,a.description,a.status,a.admin_note
		FROM seller_applications a LEFT JOIN users u ON u.id=a.user_id WHERE a.status='PENDING' ORDER BY a.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.SellerApplication{}
	for rows.Next() {
		var a model.SellerApplication
		rows.Scan(&a.ID, &a.UserID, &a.UserEmail, &a.Username, &a.StoreName, &a.Phone, &a.Address, &a.IDNumber, &a.ProductTypes, &a.Description, &a.Status, &a.AdminNote)
		out = append(out, a)
	}
	if out == nil {
		out = []model.SellerApplication{}
	}
	return out, nil
}
func (r *SellerRepo) Decide(ctx context.Context, id, status, note string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var uid string
	if err := tx.QueryRow(ctx, `SELECT user_id FROM seller_applications WHERE id=$1 AND status='PENDING'`, id).Scan(&uid); err != nil {
		return errors.New("application not found or already decided")
	}
	if _, err := tx.Exec(ctx, `UPDATE seller_applications SET status=$1, admin_note=$2, decided_at=now() WHERE id=$3`, status, note, id); err != nil {
		return err
	}
	if status == "APPROVED" {
		if _, err := tx.Exec(ctx, `UPDATE users SET seller_verified=true WHERE id=$1`, uid); err != nil {
			return err
		}
		_, _ = tx.Exec(ctx, `INSERT INTO notifications(id,user_id,title,body) VALUES(gen_random_uuid(),$1,'Seller approved','Toko kamu terverifikasi, silakan jual barang')`, uid)
	} else {
		_, _ = tx.Exec(ctx, `INSERT INTO notifications(id,user_id,title,body) VALUES(gen_random_uuid(),$1,'Seller rejected',$2)`, uid, "Pengajuan ditolak: "+note)
	}
	return tx.Commit(ctx)
}

// --- Banned words ---
type BannedRepo struct{ pool *pgxpool.Pool }

func (r *BannedRepo) List(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT word FROM banned_words ORDER BY word`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var w string
		rows.Scan(&w)
		out = append(out, w)
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}
func (r *BannedRepo) Add(ctx context.Context, word string) error {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return errors.New("word required")
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO banned_words(word) VALUES($1) ON CONFLICT DO NOTHING`, word)
	return err
}
func (r *BannedRepo) Remove(ctx context.Context, word string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM banned_words WHERE word=$1`, strings.ToLower(word))
	return err
}
func (r *BannedRepo) FindViolation(text string) (string, error) {
	words, err := r.List(context.Background())
	if err != nil {
		return "", err
	}
	low := strings.ToLower(text)
	for _, w := range words {
		if w != "" && strings.Contains(low, strings.ToLower(w)) {
			return w, nil
		}
	}
	return "", nil
}

// --- Admin ---
type AdminRepo struct{ pool *pgxpool.Pool }

func (r *AdminRepo) Stats(ctx context.Context) (map[string]any, error) {
	var users, products, orders int
	var revenue int64
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&users)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE status != 'ARCHIVED'`).Scan(&products)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&orders)
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total),0) FROM orders WHERE status='PAID'`).Scan(&revenue)
	return map[string]any{"users": users, "products": products, "orders": orders, "revenue": revenue}, nil
}
