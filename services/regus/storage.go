package regus

import (
	"context"

	"github.com/bakaoh/lavato/plugins/binance"
	"github.com/jinzhu/gorm"
	// include mysql OR sqlite
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Storage ...
type Storage struct {
	db *gorm.DB
}

// NewStorage ...
func NewStorage(dbFile string) (*Storage, error) {
	db, err := gorm.Open("sqlite3", dbFile)
	if err != nil {
		return &Storage{}, err
	}
	db.AutoMigrate(&binance.GetOrderResponse{})
	db.AutoMigrate(&Battle{})
	db.AutoMigrate(&Paladin{})

	return &Storage{
		db: db,
	}, nil
}

// SaveOrder ...
func (db *Storage) SaveOrder(ctx context.Context, item *binance.GetOrderResponse) error {
	return db.db.Save(&item).Error
}

// LoadOrder ...
func (db *Storage) LoadOrder(ctx context.Context, id int64) (*binance.GetOrderResponse, error) {
	var item binance.GetOrderResponse
	if err := db.db.First(&item, id).Error; err != nil {
		return &item, err
	}

	return &item, nil
}

// LoadAllOrders ...
func (db *Storage) LoadAllOrders(ctx context.Context) ([]binance.GetOrderResponse, error) {
	var items []binance.GetOrderResponse
	if err := db.db.Find(&items).Error; err != nil {
		return items, err
	}

	return items, nil
}

// GetMapOrders ...
func (db *Storage) GetMapOrders(ctx context.Context) (map[int64]binance.GetOrderResponse, error) {
	items, err := db.LoadAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	rs := make(map[int64]binance.GetOrderResponse)
	for _, item := range items {
		rs[item.OrderID] = item
	}

	return rs, nil
}

// SaveBattle ...
func (db *Storage) SaveBattle(ctx context.Context, item *Battle) error {
	return db.db.Save(&item).Error
}

// LoadBattle ...
func (db *Storage) LoadBattle(ctx context.Context, id int64) (*Battle, error) {
	var item Battle
	if err := db.db.First(&item, id).Error; err != nil {
		return &item, err
	}

	return &item, nil
}

// LoadAllBattles ...
func (db *Storage) LoadAllBattles(ctx context.Context) ([]Battle, error) {
	var items []Battle
	if err := db.db.Find(&items).Error; err != nil {
		return items, err
	}

	return items, nil
}

// SavePaladin ...
func (db *Storage) SavePaladin(ctx context.Context, item *Paladin) error {
	return db.db.Save(&item).Error
}

// LoadPaladin ...
func (db *Storage) LoadPaladin(ctx context.Context, id string) (*Paladin, error) {
	var item Paladin
	if err := db.db.First(&item, id).Error; err != nil {
		return &item, err
	}

	return &item, nil
}

// LoadAllPaladins ...
func (db *Storage) LoadAllPaladins(ctx context.Context) ([]Paladin, error) {
	var items []Paladin
	if err := db.db.Find(&items).Error; err != nil {
		return items, err
	}

	return items, nil
}

// Close ...
func (db *Storage) Close() {
	db.db.Close()
}
