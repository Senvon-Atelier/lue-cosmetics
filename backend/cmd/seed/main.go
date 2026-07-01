package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

type SeedData struct {
	Categories []Category
	Brands     []Brand
	Users      []User
	Products   []Product
}

type Category struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type Brand struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type Product struct {
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Description  string   `json:"description"`
	CategorySlug string   `json:"category_slug"`
	BrandSlug    string   `json:"brand_slug"`
	PriceGHS     float64  `json:"price_ghs"`
	StockQty     int      `json:"stock_qty"`
	ImageURL     string   `json:"image_url"`
	Tags         []string `json:"tags"`
}

func main() {
	dbURL := flag.String("db-url", "", "Database connection URL")
	dataDir := flag.String("data-dir", "./data", "Directory containing seed data JSON files")
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			log.Fatal("DATABASE_URL environment variable or --db-url flag required")
		}
	}

	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatalf("Failed to resolve data directory: %v", err)
	}

	seed, err := loadSeedData(absDataDir)
	if err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	q := sqlc.New(pool)

	if err := seedCategories(context.Background(), q, seed.Categories); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}
	fmt.Printf("✅ Seeded %d categories\n", len(seed.Categories))

	if err := seedBrands(context.Background(), q, seed.Brands); err != nil {
		log.Fatalf("Failed to seed brands: %v", err)
	}
	fmt.Printf("✅ Seeded %d brands\n", len(seed.Brands))

	if err := seedUsers(context.Background(), q, seed.Users); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}
	fmt.Printf("✅ Seeded %d users\n", len(seed.Users))

	if err := seedProducts(context.Background(), pool, seed.Products); err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}
	fmt.Printf("✅ Seeded %d products\n", len(seed.Products))

	fmt.Println("🌱 Seed complete!")
}

func loadSeedData(dataDir string) (*SeedData, error) {
	seed := &SeedData{}

	if err := loadJSON(filepath.Join(dataDir, "categories.json"), &seed.Categories); err != nil {
		return nil, fmt.Errorf("load categories: %w", err)
	}

	if err := loadJSON(filepath.Join(dataDir, "brands.json"), &seed.Brands); err != nil {
		return nil, fmt.Errorf("load brands: %w", err)
	}

	if err := loadJSON(filepath.Join(dataDir, "users.json"), &seed.Users); err != nil {
		return nil, fmt.Errorf("load users: %w", err)
	}

	if err := loadJSON(filepath.Join(dataDir, "products.json"), &seed.Products); err != nil {
		return nil, fmt.Errorf("load products: %w", err)
	}

	return seed, nil
}

func loadJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func seedCategories(ctx context.Context, q *sqlc.Queries, categories []Category) error {
	for _, cat := range categories {
		_, err := q.CreateCategory(ctx, sqlc.CreateCategoryParams{
			Slug:  cat.Slug,
			Label: cat.Name,
		})
		if err != nil {
			return fmt.Errorf("create category %s: %w", cat.Slug, err)
		}
	}
	return nil
}

func seedBrands(ctx context.Context, q *sqlc.Queries, brands []Brand) error {
	for _, brand := range brands {
		_, err := q.CreateBrand(ctx, sqlc.CreateBrandParams{
			Slug: brand.Slug,
			Name: brand.Name,
		})
		if err != nil {
			return fmt.Errorf("create brand %s: %w", brand.Slug, err)
		}
	}
	return nil
}

func seedUsers(ctx context.Context, q *sqlc.Queries, users []User) error {
	for _, user := range users {
		hashedPassword, err := auth.Hash(user.Password, auth.DefaultParams)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", user.Email, err)
		}

		userRecord, err := q.CreateUser(ctx, sqlc.CreateUserParams{
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: true,
		})
		if err != nil {
			return fmt.Errorf("create user %s: %w", user.Email, err)
		}

		if err := q.UpsertPasswordCredential(ctx, sqlc.UpsertPasswordCredentialParams{
			UserID:       userRecord.ID,
			PasswordHash: hashedPassword,
		}); err != nil {
			return fmt.Errorf("upsert password for %s: %w", user.Email, err)
		}

		if err := q.AddUserRole(ctx, sqlc.AddUserRoleParams{
			UserID: userRecord.ID,
			Role:   user.Role,
		}); err != nil {
			return fmt.Errorf("assign role %s to user %s: %w", user.Role, user.Email, err)
		}
	}
	return nil
}

func seedProducts(ctx context.Context, pool *pgxpool.Pool, products []Product) error {
	q := sqlc.New(pool)

	// Build slug->ID maps for categories and brands
	categories, err := q.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}
	categoryIDMap := make(map[string]uuid.UUID)
	for _, cat := range categories {
		categoryIDMap[cat.Slug] = cat.ID
	}

	brands, err := q.ListBrands(ctx)
	if err != nil {
		return fmt.Errorf("list brands: %w", err)
	}
	brandIDMap := make(map[string]uuid.UUID)
	for _, brand := range brands {
		brandIDMap[brand.Slug] = brand.ID
	}

	// Create products
	for _, prod := range products {
		brandID, ok := brandIDMap[prod.BrandSlug]
		if !ok {
			return fmt.Errorf("brand not found: %s", prod.BrandSlug)
		}

		categoryID, ok := categoryIDMap[prod.CategorySlug]
		if !ok {
			return fmt.Errorf("category not found: %s", prod.CategorySlug)
		}

		priceMinor := int64(prod.PriceGHS * 100)

		_, err := q.CreateProduct(ctx, sqlc.CreateProductParams{
			Slug:          prod.Slug,
			Name:          prod.Name,
			BrandID:       brandID,
			CategoryID:    categoryID,
			PriceGhsMinor: priceMinor,
			Tags:          prod.Tags,
			ImagePath:     prod.ImageURL,
		})
		if err != nil {
			return fmt.Errorf("create product %s: %w", prod.Slug, err)
		}
	}
	return nil
}
