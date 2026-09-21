package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	// idempotent via ON CONFLICT — default admin
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, username, full_name, role, is_verified, is_active)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'admin', true, true)
		ON CONFLICT (email) DO UPDATE SET role='admin', is_active=true`,
		"admin@thriftin.local", string(hash), "admin", "Thriftin Admin")
	if err != nil {
		return err
	}
	log.Println("seed: admin@thriftin.local / Admin123! (role=admin) ✓")
	return nil
}
