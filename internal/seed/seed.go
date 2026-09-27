package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, username, full_name, role, is_verified, is_active)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'admin', true, true)
		ON CONFLICT (email) DO UPDATE SET role='admin', is_active=true`,
		"admin@thriftin.local", string(hash), "admin", "Thriftin Admin")
	if err != nil {
		return err
	}
	log.Println("seed: admin@thriftin.local / Admin123! (role=admin)")
 
	// demo categories
	cats := []string{"Jackets", "Denim", "Streetwear", "Knitwear", "Shoes", "Bags", "Accessories", "Vintage", "Tops", "Bottoms", "Others"}
	for _, n := range cats {
		_, _ = pool.Exec(ctx, `INSERT INTO categories(id,name,slug) VALUES(gen_random_uuid(),$1,$2) ON CONFLICT (slug) DO NOTHING`, n, slug(n))
	}
	brands := []string{"Levi's", "Nike", "Adidas", "Uniqlo", "Vintage"}
	for _, n := range brands {
		_, _ = pool.Exec(ctx, `INSERT INTO brands(id,name,slug) VALUES(gen_random_uuid(),$1,$2) ON CONFLICT (slug) DO NOTHING`, n, slug(n))
	}

	// demo seller + products if empty
	var cnt int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE status='ACTIVE'`).Scan(&cnt)
	if cnt == 0 {
		var sellerID string
		sh, _ := bcrypt.GenerateFromPassword([]byte("Seller123!"), bcrypt.DefaultCost)
		_ = pool.QueryRow(ctx, `INSERT INTO users(id,email,password_hash,username,full_name,role,is_verified,is_active)
			VALUES(gen_random_uuid(),'seller@thriftin.local',$1,'seller','Demo Seller','user',true,true)
			ON CONFLICT (email) DO UPDATE SET is_active=true RETURNING id`, string(sh)).Scan(&sellerID)
		if sellerID == "" {
			_ = pool.QueryRow(ctx, `SELECT id FROM users WHERE email='seller@thriftin.local'`).Scan(&sellerID)
		}
		demo := []struct{ title, desc string; price int64 }{
			{"Vintage Levi's Denim Jacket", "Size L, good condition, 90s", 349000},
			{"Nike Windbreaker Second", "Size M, like new", 275000},
			{"Knitwear Uniqlo Wool", "Size M, fair, no defect", 99000},
			{"Adidas Samba Vintage", "Size 42, good", 450000},
			{"Corduroy Jacket Retro", "Size L, like new", 299000},
			{"Thrift Hoodie Oversize", "Size XL, good", 149000},
		}
		for i, d := range demo {
			var pid string
			_ = pool.QueryRow(ctx, `INSERT INTO products(id,seller_id,title,description,price,condition,status)
				VALUES(gen_random_uuid(),$1,$2,$3,$4,'good','ACTIVE') RETURNING id`, sellerID, d.title, d.desc, d.price).Scan(&pid)
			if pid != "" {
				_, _ = pool.Exec(ctx, `INSERT INTO product_images(id,product_id,url) VALUES(gen_random_uuid(),$1,$2)`, pid, "https://picsum.photos/seed/thriftin"+itoa(i)+"/600/600")
			}
		}
		log.Println("seed: demo seller seller@thriftin.local / Seller123! + 6 products")
	}
	return nil
}

func slug(s string) string {
	out := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 32
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out += string(r)
		} else if r == ' ' || r == '\'' || r == '_' {
			out += "-"
		}
	}
	return out
}
func itoa(i int) string { return string(rune('0' + i)) + "-demo" }
