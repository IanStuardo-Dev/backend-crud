package seeds

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	passwordinfra "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/security/password"
)

var ErrEmbedderUnavailable = errors.New("product embedder is not configured for seeds")

func Run(ctx context.Context, db *sql.DB, embedder productapp.Embedder) error {
	hasher := passwordinfra.NewBcryptHasher(0)
	if embedder == nil {
		return ErrEmbedderUnavailable
	}

	if err := ensureSeedUser(ctx, db, hasher, seedUser{
		Email:    getenvDefault("SEED_SUPER_ADMIN_EMAIL", "superadmin@example.com"),
		Password: getenvDefault("SEED_SUPER_ADMIN_PASSWORD", "Password123"),
		Name:     "Super Admin",
		Role:     "super_admin",
		IsActive: true,
	}); err != nil {
		return err
	}

	if err := ensureSeedUser(ctx, db, hasher, seedUser{
		CompanyID:       int64Pointer(1),
		DefaultBranchID: int64Pointer(1),
		Email:           getenvDefault("SEED_COMPANY_ADMIN_EMAIL", "admin@default-company.local"),
		Password:        getenvDefault("SEED_COMPANY_ADMIN_PASSWORD", "Password123"),
		Name:            "Company Admin",
		Role:            "company_admin",
		IsActive:        true,
	}); err != nil {
		return err
	}

	if err := ensureSeedUser(ctx, db, hasher, seedUser{
		CompanyID:       int64Pointer(1),
		DefaultBranchID: int64Pointer(1),
		Email:           getenvDefault("SEED_INVENTORY_MANAGER_EMAIL", "inventory@default-company.local"),
		Password:        getenvDefault("SEED_INVENTORY_MANAGER_PASSWORD", "Password123"),
		Name:            "Inventory Manager",
		Role:            "inventory_manager",
		IsActive:        true,
	}); err != nil {
		return err
	}

	if err := ensureSeedUser(ctx, db, hasher, seedUser{
		CompanyID:       int64Pointer(1),
		DefaultBranchID: int64Pointer(1),
		Email:           getenvDefault("SEED_SALES_USER_EMAIL", "sales@default-company.local"),
		Password:        getenvDefault("SEED_SALES_USER_PASSWORD", "Password123"),
		Name:            "Sales User",
		Role:            "sales_user",
		IsActive:        true,
	}); err != nil {
		return err
	}

	for _, product := range seedProducts() {
		if err := ensureSeedProduct(ctx, db, embedder, product); err != nil {
			return err
		}
	}

	return nil
}

type seedUser struct {
	CompanyID       *int64
	DefaultBranchID *int64
	Email           string
	Password        string
	Name            string
	Role            string
	IsActive        bool
}

type passwordHasher interface {
	Hash(password string) (string, error)
}

type seedProduct struct {
	CompanyID   int64
	BranchID    int64
	SKU         string
	Name        string
	Description string
	Category    string
	Brand       string
	PriceCents  int64
	Currency    string
	Stock       int
}

func ensureSeedUser(ctx context.Context, db *sql.DB, hasher passwordHasher, user seedUser) error {
	passwordHash, err := hasher.Hash(user.Password)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO users (company_id, name, email, role, is_active, default_branch_id, password_hash, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)
		ON CONFLICT (email)
		DO UPDATE SET
			company_id = EXCLUDED.company_id,
			name = EXCLUDED.name,
			role = EXCLUDED.role,
			is_active = EXCLUDED.is_active,
			default_branch_id = EXCLUDED.default_branch_id,
			password_hash = EXCLUDED.password_hash,
			updated_at = EXCLUDED.updated_at`,
		nullableInt64(user.CompanyID),
		user.Name,
		user.Email,
		user.Role,
		user.IsActive,
		nullableInt64(user.DefaultBranchID),
		passwordHash,
		time.Now().UTC(),
	)
	return err
}

func ensureSeedProduct(ctx context.Context, db *sql.DB, embedder productapp.Embedder, product seedProduct) error {
	embedding, err := embedder.EmbedText(ctx, productEmbeddingText(product))
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var productID int64
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO products (
			company_id,
			branch_id,
			sku,
			name,
			description,
			category,
			brand,
			price_cents,
			currency,
			stock,
			embedding,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::vector,$12,$12)
		ON CONFLICT (branch_id, sku)
		DO UPDATE SET
			company_id = EXCLUDED.company_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			category = EXCLUDED.category,
			brand = EXCLUDED.brand,
			price_cents = EXCLUDED.price_cents,
			currency = EXCLUDED.currency,
			stock = EXCLUDED.stock,
			embedding = EXCLUDED.embedding,
			updated_at = EXCLUDED.updated_at
		RETURNING id`,
		product.CompanyID,
		product.BranchID,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Brand,
		product.PriceCents,
		product.Currency,
		product.Stock,
		formatVector(embedding),
		time.Now().UTC(),
	).Scan(&productID)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO branch_inventory (
			company_id,
			branch_id,
			product_id,
			stock_on_hand,
			reserved_stock,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,0,$5,$5)
		ON CONFLICT (branch_id, product_id)
		DO UPDATE SET
			company_id = EXCLUDED.company_id,
			stock_on_hand = EXCLUDED.stock_on_hand,
			reserved_stock = EXCLUDED.reserved_stock,
			updated_at = EXCLUDED.updated_at`,
		product.CompanyID,
		product.BranchID,
		productID,
		product.Stock,
		time.Now().UTC(),
	); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func seedProducts() []seedProduct {
	return []seedProduct{
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-001",
			Name:        "Wireless Mouse",
			Description: "Mouse inalambrico ergonomico para oficina y uso diario.",
			Category:    "peripherals",
			Brand:       "Acme",
			PriceCents:  19990,
			Currency:    "CLP",
			Stock:       35,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-002",
			Name:        "Laptop Cooling Pad",
			Description: "Base de enfriamiento para laptop con ventiladores silenciosos.",
			Category:    "office",
			Brand:       "Northwind",
			PriceCents:  27990,
			Currency:    "CLP",
			Stock:       18,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-003",
			Name:        "Mechanical Keyboard",
			Description: "Teclado mecanico compacto con switches tactiles.",
			Category:    "peripherals",
			Brand:       "Northwind",
			PriceCents:  54990,
			Currency:    "CLP",
			Stock:       14,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-004",
			Name:        "USB-C Dock Station",
			Description: "Dock con HDMI, USB y carga para notebooks modernas.",
			Category:    "accessories",
			Brand:       "Northwind",
			PriceCents:  45990,
			Currency:    "CLP",
			Stock:       11,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-005",
			Name:        "Laptop Stand Aluminum",
			Description: "Soporte de aluminio para laptop con mejor ventilacion.",
			Category:    "office",
			Brand:       "Acme",
			PriceCents:  24990,
			Currency:    "CLP",
			Stock:       20,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-006",
			Name:        "Cafe Dolca Instantaneo 170g",
			Description: "Cafe soluble clasico en frasco de 170 gramos para consumo diario.",
			Category:    "beverages",
			Brand:       "Dolca",
			PriceCents:  6490,
			Currency:    "CLP",
			Stock:       28,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-007",
			Name:        "Cafe Nestle Clasico 170g",
			Description: "Cafe instantaneo tradicional en frasco de 170 gramos.",
			Category:    "beverages",
			Brand:       "Nestle",
			PriceCents:  6990,
			Currency:    "CLP",
			Stock:       24,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-008",
			Name:        "Coffee Marley Instant Blend 170g",
			Description: "Instant coffee blend de 170 gramos con sabor intenso.",
			Category:    "beverages",
			Brand:       "Marley",
			PriceCents:  7490,
			Currency:    "CLP",
			Stock:       19,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-009",
			Name:        "Bluetooth Speaker Mini",
			Description: "Parlante portatil compacto con bateria recargable.",
			Category:    "audio",
			Brand:       "SoundLab",
			PriceCents:  32990,
			Currency:    "CLP",
			Stock:       22,
		},
		{
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-010",
			Name:        "Wireless Ergonomic Mouse",
			Description: "Mouse inalambrico ergonomico con agarre comodo para largas jornadas.",
			Category:    "peripherals",
			Brand:       "Acme",
			PriceCents:  25990,
			Currency:    "CLP",
			Stock:       7,
		},
	}
}

func productEmbeddingText(product seedProduct) string {
	parts := make([]string, 0, 5)
	if value := strings.TrimSpace(product.SKU); value != "" {
		parts = append(parts, "SKU: "+value)
	}
	if value := strings.TrimSpace(product.Name); value != "" {
		parts = append(parts, "Name: "+value)
	}
	if value := strings.TrimSpace(product.Description); value != "" {
		parts = append(parts, "Description: "+value)
	}
	if value := strings.TrimSpace(product.Category); value != "" {
		parts = append(parts, "Category: "+value)
	}
	if value := strings.TrimSpace(product.Brand); value != "" {
		parts = append(parts, "Brand: "+value)
	}

	return strings.Join(parts, "\n")
}

func formatVector(embedding []float32) any {
	if len(embedding) == 0 {
		return nil
	}

	values := make([]string, 0, len(embedding))
	for _, value := range embedding {
		values = append(values, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}

	return "[" + strings.Join(values, ",") + "]"
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func int64Pointer(value int64) *int64 {
	return &value
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
