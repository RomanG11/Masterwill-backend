// Package seed populates a freshly created database with the demo catalog
// used throughout development and the admin bootstrap account, so a clone
// of this repo is immediately explorable without manual data entry.
package seed

import (
	"context"
	"fmt"
	"log"

	"masterwill-backend/internal/auth"
	"masterwill-backend/internal/models"
	"masterwill-backend/internal/store"
)

type demoProduct struct {
	slug, categorySlug, name, description, ageLabel, icon, accent string
	priceCents                                                    int64
	stock                                                          int
}

var demoCategories = []models.Category{
	{Slug: "constructors", Name: "Конструктори", SortOrder: 1},
	{Slug: "plush", Name: "М'які іграшки", SortOrder: 2},
	{Slug: "board-games", Name: "Настільні ігри", SortOrder: 3},
	{Slug: "dolls", Name: "Ляльки й будиночки", SortOrder: 4},
	{Slug: "crafts", Name: "Творчість", SortOrder: 5},
	{Slug: "vehicles", Name: "Транспорт", SortOrder: 6},
}

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
func Run(ctx context.Context, s *store.Store, adminEmail, adminPassword string) error {
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
			Icon:        p.icon,
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
