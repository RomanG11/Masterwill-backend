// Package seed populates a freshly created database with the demo catalog
// used throughout development and the admin bootstrap account, so a clone
// of this repo is immediately explorable without manual data entry.
package seed

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"masterwill-backend/internal/auth"
	"masterwill-backend/internal/models"
	"masterwill-backend/internal/store"
)

//go:embed photos/*.jpg
var photosFS embed.FS

type demoProduct struct {
	slug, categorySlug, name, description, ageLabel, photoKey, accent string
	priceCents                                                        int64
	stock                                                              int
}

var demoCategories = []models.Category{
	{Slug: "constructors", Name: "Конструктори", SortOrder: 1},
	{Slug: "plush", Name: "М'які іграшки", SortOrder: 2},
	{Slug: "board-games", Name: "Настільні ігри", SortOrder: 3},
	{Slug: "dolls", Name: "Ляльки й будиночки", SortOrder: 4},
	{Slug: "crafts", Name: "Творчість", SortOrder: 5},
	{Slug: "vehicles", Name: "Транспорт", SortOrder: 6},
}

// photoKey names a file under photos/<key>.jpg, copied to the uploads dir
// (as /uploads/<key>.jpg) the same way an admin-uploaded photo would be.
var demoProducts = []demoProduct{
	{"zamok-drakona", "constructors", "«Замок дракона», 240 деталей", "Конструктор для будівництва казкового замку з рухомими вежами.", "3+", "castle", "yellow", 119000, 14},
	{"vedmedyk-topa", "plush", "Плюшевий ведмедик «Тьопа», 35 см", "М'яка іграшка з гіпоалергенного плюшу, можна прати.", "0+", "bear", "pink", 59000, 22},
	{"zvirobrodylka", "board-games", "«Звіробродилка», 2–4 гравці", "Сімейна гра-ходилка з картками пригод.", "5+", "board", "teal", 45000, 18},
	{"lyalkovyi-budynochok", "dolls", "Ляльковий будиночок «Ромашка»", "Триповерховий будиночок з меблями та ліфтом.", "3+", "dollhouse", "yellow", 219000, 6},
	{"raiduzhnyi-slaim", "crafts", "Набір «Райдужний слайм», 6 кольорів", "Набір для створення слайму власними руками.", "6+", "slime", "pink", 33000, 30},
	{"mashynka-turbo", "vehicles", "Машинка на радіокеруванні «Турбо»", "Швидкість до 15 км/год, акумулятор у комплекті.", "4+", "car", "teal", 89000, 11},
}

// Run seeds categories/products on first launch and ensures the configured
// admin account exists (creating it if the admin_users table is empty).
// uploadsDir is where the demo product photos get copied to, so they're
// served by the same /uploads/ route as real admin uploads.
func Run(ctx context.Context, s *store.Store, adminEmail, adminPassword, uploadsDir string) error {
	if err := extractSeedPhotos(uploadsDir); err != nil {
		return err
	}

	n, err := s.CountProducts(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		if err := seedCatalog(ctx, s); err != nil {
			return err
		}
		log.Println("seed: demo catalog created")
	}

	count, err := s.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		hash, err := auth.HashPassword(adminPassword)
		if err != nil {
			return err
		}
		if err := s.CreateAdmin(ctx, adminEmail, hash); err != nil {
			return err
		}
		log.Printf("seed: admin account created (%s)", adminEmail)
	}
	return nil
}

// extractSeedPhotos copies the embedded demo photos into uploadsDir if
// they're not already there — cheap to re-check on every startup, and
// keeps a redeploy from clobbering a photo an admin has since replaced.
func extractSeedPhotos(uploadsDir string) error {
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return fmt.Errorf("create uploads dir: %w", err)
	}
	entries, err := fs.ReadDir(photosFS, "photos")
	if err != nil {
		return fmt.Errorf("read embedded photos: %w", err)
	}
	for _, e := range entries {
		dst := filepath.Join(uploadsDir, e.Name())
		if _, err := os.Stat(dst); err == nil {
			continue // already extracted (or an admin's own upload took this name)
		}
		data, err := photosFS.ReadFile("photos/" + e.Name())
		if err != nil {
			return fmt.Errorf("read embedded photo %s: %w", e.Name(), err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("write seed photo %s: %w", e.Name(), err)
		}
	}
	return nil
}

func seedCatalog(ctx context.Context, s *store.Store) error {
	slugToID := map[string]int64{}
	for _, c := range demoCategories {
		created, err := s.CreateCategory(ctx, c)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.Slug, err)
		}
		slugToID[c.Slug] = created.ID
	}

	for _, p := range demoProducts {
		catID, ok := slugToID[p.categorySlug]
		if !ok {
			return fmt.Errorf("seed product %s: unknown category %s", p.slug, p.categorySlug)
		}
		_, err := s.CreateProduct(ctx, models.Product{
			Slug:        p.slug,
			CategoryID:  catID,
			Name:        p.name,
			Description: p.description,
			AgeLabel:    p.ageLabel,
			PhotoURL:    "/uploads/" + p.photoKey + ".jpg",
			AccentColor: p.accent,
			PriceCents:  p.priceCents,
			Currency:    "UAH",
			StockQty:    p.stock,
			IsActive:    true,
		})
		if err != nil {
			return fmt.Errorf("seed product %s: %w", p.slug, err)
		}
	}
	return nil
}
