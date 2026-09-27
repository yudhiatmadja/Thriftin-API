package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/thriftin/api/internal/config"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/repository"
	googlepkg "github.com/thriftin/api/pkg/google"
	jwtpkg "github.com/thriftin/api/pkg/jwt"
)

type Services struct {
	Auth         *AuthService
	Product      *ProductService
	Category     *CategoryService
	Cart         *CartService
	Order        *OrderService
	Wishlist     *WishlistService
	Review       *ReviewService
	Follow       *FollowService
	Chat         *ChatService
	Notification *NotificationService
	Report       *ReportService
	Seller       *SellerService
	Banned       *BannedService
}
 
func New(repos *repository.Repos, cfg *config.Config) *Services {
	return &Services{
		Auth: &AuthService{users: repos.User, cfg: cfg},
		Product: &ProductService{repo: repos.Product, seller: repos.Seller, banned: repos.Banned},
		Category: &CategoryService{repo: repos.Category},
		Cart: &CartService{repo: repos.Cart},
		Order: &OrderService{repo: repos.Order},
		Wishlist: &WishlistService{repo: repos.Wishlist},
		Review: &ReviewService{repo: repos.Review},
		Follow: &FollowService{repo: repos.Follow},
		Chat: &ChatService{repo: repos.Chat, products: repos.Product},
		Notification: &NotificationService{repo: repos.Notification},
		Report: &ReportService{repo: repos.Report},
		Seller: &SellerService{repo: repos.Seller},
		Banned: &BannedService{repo: repos.Banned},
	}
}

type AuthService struct {
	users interface {
		Create(ctx context.Context, u *model.User) error
		ByEmail(ctx context.Context, email string) (*model.User, error)
		ByID(ctx context.Context, id string) (*model.User, error)
		UsernameTaken(ctx context.Context, username string) (bool, error)
		StoreRefresh(ctx context.Context, userID, token string, exp time.Time) error
		FindRefresh(ctx context.Context, token string) (string, error)
		RevokeRefresh(ctx context.Context, token string) error
	}
	cfg *config.Config
}

func (s *AuthService) Register(ctx context.Context, email, pass, username, fullName string) (*model.User, error) {
	if len(pass) < 6 {
		return nil, errors.New("password min 6 chars")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
	if err != nil {
		return nil, err
	}
	u := &model.User{ID: uuid.NewString(), Email: email, PasswordHash: string(hash), Username: username, FullName: fullName, Role: "user", IsActive: true}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return u, nil
}
func (s *AuthService) Login(ctx context.Context, email, pass string) (access, refresh string, user *model.User, err error) {
	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)) != nil {
		return "", "", nil, errors.New("invalid credentials")
	}
	if !u.IsActive {
		return "", "", nil, errors.New("account deactivated")
	}
	access, _ = jwtpkg.Generate(s.cfg.JWTSecret, u.ID, u.Role, s.cfg.JWTExpiry)
	refresh, _ = jwtpkg.Generate(s.cfg.JWTSecret, u.ID, u.Role, s.cfg.RefreshExpiry)
	_ = s.users.StoreRefresh(ctx, u.ID, refresh, time.Now().Add(s.cfg.RefreshExpiry))
	u.PasswordHash = ""
	return access, refresh, u, nil
}
func (s *AuthService) GoogleLogin(ctx context.Context, idToken string) (access, refresh string, user *model.User, err error) {
	prof, err := googlepkg.VerifyIDToken(s.cfg.GoogleClientID, idToken)
	if err != nil {
		return "", "", nil, err
	}
	u, err := s.users.ByEmail(ctx, prof.Email)
	if err != nil {
		// akun belum ada: buatkan otomatis, email sudah terbukti miliknya via Google
		base := strings.ToLower(prof.Email)
		if i := strings.Index(base, "@"); i > 0 {
			base = base[:i]
		}
		base = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
				return r
			}
			return '_'
		}, base)
		if base == "" {
			base = "user"
		}
		username := base
		for i := 0; i < 5; i++ {
			taken, _ := s.users.UsernameTaken(ctx, username)
			if !taken {
				break
			}
			username = fmt.Sprintf("%s_%s", base, uuid.NewString()[:4])
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), 10)
		u = &model.User{ID: uuid.NewString(), Email: prof.Email, PasswordHash: string(hash), Username: username, FullName: prof.Name, Role: "user", IsVerified: true, IsActive: true}
		if err := s.users.Create(ctx, u); err != nil {
			return "", "", nil, err
		}
	}
	if !u.IsActive {
		return "", "", nil, errors.New("account deactivated")
	}
	access, _ = jwtpkg.Generate(s.cfg.JWTSecret, u.ID, u.Role, s.cfg.JWTExpiry)
	refresh, _ = jwtpkg.Generate(s.cfg.JWTSecret, u.ID, u.Role, s.cfg.RefreshExpiry)
	_ = s.users.StoreRefresh(ctx, u.ID, refresh, time.Now().Add(s.cfg.RefreshExpiry))
	u.PasswordHash = ""
	return access, refresh, u, nil
}
func (s *AuthService) Me(ctx context.Context, id string) (*model.User, error) {
	u, err := s.users.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return u, nil
}
func (s *AuthService) Refresh(ctx context.Context, token string) (string, string, error) {
	c, err := jwtpkg.Validate(s.cfg.JWTSecret, token)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	uid, err := s.users.FindRefresh(ctx, token)
	if err != nil {
		return "", "", errors.New("refresh token expired or revoked")
	}
	if uid != c.UserID {
		return "", "", errors.New("invalid refresh token")
	}
	_ = s.users.RevokeRefresh(ctx, token)
	a, _ := jwtpkg.Generate(s.cfg.JWTSecret, c.UserID, c.Role, s.cfg.JWTExpiry)
	r, _ := jwtpkg.Generate(s.cfg.JWTSecret, c.UserID, c.Role, s.cfg.RefreshExpiry)
	_ = s.users.StoreRefresh(ctx, c.UserID, r, time.Now().Add(s.cfg.RefreshExpiry))
	return a, r, nil
}
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.users.RevokeRefresh(ctx, token)
}

type ProductService struct {
	repo   *repository.ProductRepo
	seller interface {
		IsVerified(ctx context.Context, uid string) bool
	}
	banned interface {
		FindViolation(text string) (string, error)
	}
}

func (s *ProductService) checkBanned(title, desc string) error {
	if s.banned == nil {
		return nil
	}
	if w, _ := s.banned.FindViolation(title + " " + desc); w != "" {
		return errors.New("produk mengandung kata terlarang: " + w + ". Barang seperti rokok/narkoba/senjata/judi dilarang")
	}
	return nil
}

func (s *ProductService) Create(ctx context.Context, p *model.Product, images []string, userRole string) error {
	if userRole != "admin" && s.seller != nil && !s.seller.IsVerified(ctx, p.SellerID) {
		return errors.New("akun belum terverifikasi sebagai seller. Daftar di /become-seller dan tunggu approve admin")
	}
	if err := s.checkBanned(p.Title, p.Description); err != nil {
		return err
	}
	p.ID = uuid.NewString()
	if p.Status == "" {
		p.Status = "ACTIVE"
	}
	if p.Condition == "" {
		p.Condition = "good"
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return err
	}
	for _, u := range images {
		if u != "" {
			_ = s.repo.AddImage(ctx, p.ID, u)
		}
	}
	return nil
}
func (s *ProductService) Get(ctx context.Context, id string) (*model.Product, error) {
	return s.repo.Get(ctx, id)
}
func (s *ProductService) List(ctx context.Context, f model.ProductFilter) ([]model.Product, error) {
	return s.repo.List(ctx, f)
}
func (s *ProductService) MyProducts(ctx context.Context, seller string) ([]model.Product, error) {
	return s.repo.MyProducts(ctx, seller)
}
func (s *ProductService) Update(ctx context.Context, id, seller string, title, desc *string, price *int64, status *string) error {
	if title != nil || desc != nil {
		t, d := "", ""
		if title != nil {
			t = *title
		}
		if desc != nil {
			d = *desc
		}
		if err := s.checkBanned(t, d); err != nil {
			return err
		}
	}
	return s.repo.Update(ctx, id, seller, title, desc, price, status)
}
func (s *ProductService) Delete(ctx context.Context, id, seller, role string) error {
	return s.repo.Delete(ctx, id, seller, role)
}
func (s *ProductService) AddImage(ctx context.Context, id, seller, url string) error {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if p.SellerID != seller {
		return errors.New("forbidden: not your product")
	}
	return s.repo.AddImage(ctx, id, url)
}
func (s *ProductService) Reserve(ctx context.Context, id string) error {
	return s.repo.ReserveTx(ctx, id)
}

type CategoryService struct{ repo *repository.CategoryRepo }

func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) { return s.repo.List(ctx) }
func (s *CategoryService) Create(ctx context.Context, c model.Category) error {
	if c.Name == "" {
		return errors.New("name required")
	}
	return s.repo.Create(ctx, c)
}
func (s *CategoryService) ListBrands(ctx context.Context) ([]model.Brand, error) {
	return s.repo.ListBrands(ctx)
}
func (s *CategoryService) CreateBrand(ctx context.Context, b model.Brand) error {
	if b.Name == "" {
		return errors.New("name required")
	}
	return s.repo.CreateBrand(ctx, b)
}

type CartService struct{ repo *repository.CartRepo }

func (s *CartService) Add(ctx context.Context, uid, pid string, qty int) error {
	return s.repo.Add(ctx, uid, pid, qty)
}
func (s *CartService) Remove(ctx context.Context, uid, pid string) error {
	return s.repo.Remove(ctx, uid, pid)
}
func (s *CartService) List(ctx context.Context, uid string) ([]model.CartItem, error) {
	return s.repo.List(ctx, uid)
}

type OrderService struct{ repo *repository.OrderRepo }

func (s *OrderService) Checkout(ctx context.Context, uid, addr string) (*model.Order, error) {
	return s.repo.Checkout(ctx, uid, addr)
}
func (s *OrderService) List(ctx context.Context, uid string) ([]model.Order, error) {
	return s.repo.List(ctx, uid)
}
func (s *OrderService) Get(ctx context.Context, uid, oid string) (*model.Order, error) {
	return s.repo.Get(ctx, uid, oid)
}
func (s *OrderService) Pay(ctx context.Context, uid, oid string) (*model.Order, error) {
	return s.repo.Pay(ctx, uid, oid)
}

type WishlistService struct{ repo *repository.WishlistRepo }

func (s *WishlistService) Toggle(ctx context.Context, uid, pid string) (bool, error) {
	return s.repo.Toggle(ctx, uid, pid)
}
func (s *WishlistService) List(ctx context.Context, uid string) ([]model.Wishlist, error) {
	return s.repo.List(ctx, uid)
}

type ReviewService struct{ repo *repository.ReviewRepo }

func (s *ReviewService) Create(ctx context.Context, r model.Review) error {
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating 1-5")
	}
	return s.repo.Create(ctx, r)
}
func (s *ReviewService) List(ctx context.Context, pid string) ([]model.Review, error) {
	return s.repo.List(ctx, pid)
}

type FollowService struct{ repo *repository.FollowRepo }

func (s *FollowService) Follow(ctx context.Context, a, b string) error {
	return s.repo.Follow(ctx, a, b)
}
func (s *FollowService) Unfollow(ctx context.Context, a, b string) error {
	return s.repo.Unfollow(ctx, a, b)
}

type ChatService struct {
	repo     *repository.ChatRepo
	products *repository.ProductRepo
}

func (s *ChatService) CreateRoom(ctx context.Context, buyer, seller, product string) (*model.ChatRoom, error) {
	if product == "" {
		return nil, errors.New("product_id required")
	}
	// Seller diambil dari produk (anti-spoof); client boleh kosongkan seller_id
	p, err := s.products.Get(ctx, product)
	if err != nil {
		return nil, errors.New("product not found")
	}
	if p.SellerID == "" {
		return nil, errors.New("product has no seller")
	}
	if buyer == p.SellerID {
		return nil, errors.New("ini produk kamu sendiri. Tidak bisa chat dengan diri sendiri")
	}
	return s.repo.CreateRoom(ctx, buyer, p.SellerID, product)
}
func (s *ChatService) Rooms(ctx context.Context, uid string) ([]model.ChatRoom, error) {
	return s.repo.Rooms(ctx, uid)
}
func (s *ChatService) Send(ctx context.Context, m model.ChatMessage) error {
	return s.repo.Send(ctx, m)
}
func (s *ChatService) Messages(ctx context.Context, uid, rid string) ([]model.ChatMessage, error) {
	return s.repo.Messages(ctx, uid, rid)
}

type NotificationService struct{ repo *repository.NotificationRepo }

func (s *NotificationService) List(ctx context.Context, uid string) ([]model.Notification, error) {
	return s.repo.List(ctx, uid)
}
func (s *NotificationService) MarkRead(ctx context.Context, uid, id string) error {
	return s.repo.MarkRead(ctx, uid, id)
}

type SellerService struct{ repo *repository.SellerRepo }

func (s *SellerService) Apply(ctx context.Context, a model.SellerApplication) error {
	if a.StoreName == "" || a.Phone == "" || a.Address == "" || a.IDNumber == "" || a.ProductTypes == "" {
		return errors.New("store_name, phone, address, id_number, product_types wajib diisi")
	}
	return s.repo.Apply(ctx, a)
}
func (s *SellerService) MyStatus(ctx context.Context, uid string) (*model.SellerApplication, error) {
	return s.repo.MyStatus(ctx, uid)
}
func (s *SellerService) IsVerified(ctx context.Context, uid string) bool {
	return s.repo.IsVerified(ctx, uid)
}
func (s *SellerService) ListPending(ctx context.Context) ([]model.SellerApplication, error) {
	return s.repo.ListPending(ctx)
}
func (s *SellerService) Decide(ctx context.Context, id, status, note string) error {
	if status != "APPROVED" && status != "REJECTED" {
		return errors.New("status must be APPROVED or REJECTED")
	}
	return s.repo.Decide(ctx, id, status, note)
}

type BannedService struct{ repo *repository.BannedRepo }

func (s *BannedService) List(ctx context.Context) ([]string, error) { return s.repo.List(ctx) }
func (s *BannedService) Add(ctx context.Context, w string) error    { return s.repo.Add(ctx, w) }
func (s *BannedService) Remove(ctx context.Context, w string) error { return s.repo.Remove(ctx, w) }

type ReportService struct{ repo *repository.ReportRepo }

func (s *ReportService) Create(ctx context.Context, r model.Report) error {
	if r.Reason == "" || r.TargetID == "" {
		return errors.New("target_id and reason required")
	}
	return s.repo.Create(ctx, r)
}
func (s *ReportService) List(ctx context.Context) ([]model.Report, error) {
	return s.repo.List(ctx)
}
