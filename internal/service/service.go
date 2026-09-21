package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/thriftin/api/internal/config"
	"github.com/thriftin/api/internal/model"
	"github.com/thriftin/api/internal/repository"
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
}

func New(repos *repository.Repos, cfg *config.Config) *Services {
	return &Services{
		Auth: &AuthService{users: repos.User, pool: repos.Pool, cfg: cfg},
		Product: &ProductService{repo: repos.Product},
		Category: &CategoryService{repo: repos.Category},
		Cart: &CartService{repo: repos.Cart},
		Order: &OrderService{repo: repos.Order},
		Wishlist: &WishlistService{repo: repos.Wishlist},
		Review: &ReviewService{repo: repos.Review},
		Follow: &FollowService{repo: repos.Follow},
		Chat: &ChatService{repo: repos.Chat},
		Notification: &NotificationService{repo: repos.Notification},
	}
}

type AuthService struct {
	users interface {
		Create(ctx context.Context, u *model.User) error
		ByEmail(ctx context.Context, email string) (*model.User, error)
		ByID(ctx context.Context, id string) (*model.User, error)
	}
	pool interface{}
	cfg  *config.Config
}

func (s *AuthService) Register(ctx context.Context, email, pass, username, fullName string) (*model.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), 10)
	u := &model.User{ID: uuid.NewString(), Email: email, PasswordHash: string(hash), Username: username, FullName: fullName, Role: "user", IsActive: true}
	if err := s.users.Create(ctx, u); err != nil { return nil, err }
	return u, nil
}
func (s *AuthService) Login(ctx context.Context, email, pass string) (access, refresh string, err error) {
	u, err := s.users.ByEmail(ctx, email)
	if err != nil { return "", "", err }
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)) != nil { return "", "", err }
	access, _ = jwtpkg.Generate(s.cfg.JWTSecret, u.ID, u.Role, s.cfg.JWTExpiry)
	refresh, _ = jwtpkg.Generate(s.cfg.JWTSecret, u.ID, u.Role, s.cfg.RefreshExpiry)
	return access, refresh, nil
}
func (s *AuthService) Me(ctx context.Context, id string) (*model.User, error) { return s.users.ByID(ctx, id) }
func (s *AuthService) Refresh(ctx context.Context, token string) (string, string, error) {
	c, err := jwtpkg.Validate(s.cfg.JWTSecret, token)
	if err != nil { return "", "", err }
	a, _ := jwtpkg.Generate(s.cfg.JWTSecret, c.UserID, c.Role, s.cfg.JWTExpiry)
	r, _ := jwtpkg.Generate(s.cfg.JWTSecret, c.UserID, c.Role, s.cfg.RefreshExpiry)
	return a, r, nil
}

type ProductService struct{ repo *repository.ProductRepo }
func (s *ProductService) Create(ctx context.Context, p *model.Product) error { p.ID=uuid.NewString(); if p.Status==""{p.Status="ACTIVE"}; if p.Condition==""{p.Condition="good"}; return s.repo.Create(ctx,p) }
func (s *ProductService) Get(ctx context.Context, id string) (*model.Product, error) { return s.repo.Get(ctx,id) }
func (s *ProductService) List(ctx context.Context, q string, limit, offset int) ([]model.Product, error) { return s.repo.List(ctx,q,limit,offset) }
func (s *ProductService) Reserve(ctx context.Context, id string) error { return s.repo.ReserveTx(ctx,id) }

type CategoryService struct{ repo *repository.CategoryRepo }
func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) { return s.repo.List(ctx) }
func (s *CategoryService) Create(ctx context.Context, c model.Category) error { c.ID=uuid.NewString(); return s.repo.Create(ctx,c) }

type CartService struct{ repo *repository.CartRepo }
func (s *CartService) Add(ctx context.Context, uid,pid string, qty int) error { if qty==0 {qty=1}; return s.repo.Add(ctx,uid,pid,qty) }
func (s *CartService) List(ctx context.Context, uid string) ([]model.CartItem, error) { return s.repo.List(ctx,uid) }

type OrderService struct{ repo *repository.OrderRepo }
func (s *OrderService) Checkout(ctx context.Context, uid, addr string) (*model.Order, error) { return s.repo.Checkout(ctx,uid,addr) }

type WishlistService struct{ repo *repository.WishlistRepo }
func (s *WishlistService) Toggle(ctx context.Context, uid,pid string) error { return s.repo.Toggle(ctx,uid,pid) }
func (s *WishlistService) List(ctx context.Context, uid string) ([]model.Wishlist, error) { return s.repo.List(ctx,uid) }

type ReviewService struct{ repo *repository.ReviewRepo }
func (s *ReviewService) Create(ctx context.Context, r model.Review) error { return s.repo.Create(ctx,r) }
func (s *ReviewService) List(ctx context.Context, pid string) ([]model.Review, error) { return s.repo.List(ctx,pid) }

type FollowService struct{ repo *repository.FollowRepo }
func (s *FollowService) Follow(ctx context.Context, a,b string) error { return s.repo.Follow(ctx,a,b) }
func (s *FollowService) Unfollow(ctx context.Context, a,b string) error { return s.repo.Unfollow(ctx,a,b) }

type ChatService struct{ repo *repository.ChatRepo }
func (s *ChatService) Rooms(ctx context.Context, uid string) ([]model.ChatRoom, error) { return s.repo.Rooms(ctx,uid) }
func (s *ChatService) Send(ctx context.Context, m model.ChatMessage) error { return s.repo.Send(ctx,m) }
func (s *ChatService) Messages(ctx context.Context, rid string) ([]model.ChatMessage, error) { return s.repo.Messages(ctx,rid) }

type NotificationService struct{ repo *repository.NotificationRepo }
func (s *NotificationService) List(ctx context.Context, uid string) ([]model.Notification, error) { return s.repo.List(ctx,uid) }
