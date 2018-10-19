package regus

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	bin "github.com/bakaoh/lavato/plugins/binance"
)

// Battle ...
type Battle struct {
	PaladinID string
	Symbol    string
	InID      int64 `gorm:"primary_key"`
	OutID     int64
}

// Paladin ...
type Paladin struct {
	ID    string `json:"id,omitempty" gorm:"primary_key"`
	Name  string `json:"name,omitempty"`
	RawHP int    `json:"rawhp,omitempty"`
	ModHP int    `json:"modhp,omitempty"`
	InID  int64  `json:"inid,omitempty"`
	OutID int64  `json:"outid,omitempty"`
	Type  string `json:"type,omitempty"`
}

// BattleResponse ...
type BattleResponse struct {
	Symbol   string  `json:"symbol,omitempty"`
	Duration int64   `json:"duration,omitempty"`
	Percent  float64 `json:"percent,omitempty"`
}

// PaladinResponse ...
type PaladinResponse struct {
	ID        string           `json:"id,omitempty"`
	Name      string           `json:"name,omitempty"`
	RawHP     int              `json:"rawhp,omitempty"`
	ModHP     int              `json:"modhp,omitempty"`
	WinLoss   string           `json:"winloss,omitempty"`
	Symbol    string           `json:"symbol,omitempty"`
	InID      int64            `json:"inid,omitempty"`
	InPrice   string           `json:"inprice,omitempty"`
	InDate    int64            `json:"indate,omitempty"`
	InStatus  string           `json:"instatus,omitempty"`
	OutID     int64            `json:"outid,omitempty"`
	OutPrice  string           `json:"outprice,omitempty"`
	OutStatus string           `json:"outstatus,omitempty"`
	Type      string           `json:"type,omitempty"`
	Logs      []BattleResponse `json:"logs,omitempty"`
}

// PaladinSort implements sort.Interface for []PaladinResponse based on the ID field.
type PaladinSort []PaladinResponse

func (a PaladinSort) Len() int      { return len(a) }
func (a PaladinSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PaladinSort) Less(i, j int) bool {
	if a[i].Type != a[j].Type {
		return a[i].Type < a[j].Type
	}
	return a[i].ID < a[j].ID
}

// Barrack ...
type Barrack struct {
	binance    *bin.Client
	provider   *Provider
	storage    *Storage
	cache      []PaladinResponse
	fullUpdate bool
}

// NewBarrack ...
func NewBarrack(binance *bin.Client, provider *Provider, storage *Storage) *Barrack {
	return &Barrack{
		binance:  binance,
		provider: provider,
		storage:  storage,
	}
}

// GetPaladins ...
func (b *Barrack) GetPaladins(ctx context.Context) []PaladinResponse {
	paladins, _ := b.storage.LoadAllPaladins(ctx)
	orders, _ := b.storage.GetMapOrders(ctx)
	battles, _ := b.storage.LoadAllBattles(ctx)

	rs := []PaladinResponse{}
	for _, p := range paladins {
		paladin := PaladinResponse{
			ID:    p.ID,
			Name:  p.Name,
			RawHP: p.RawHP,
			ModHP: p.ModHP,
			Type:  p.Type,
		}
		if p.InID > 0 {
			inOrder := orders[p.InID]
			paladin.Symbol = inOrder.Symbol
			paladin.InID = inOrder.OrderID
			paladin.InPrice = inOrder.Price
			paladin.InDate = inOrder.Time
			hp, _ := strconv.ParseFloat(inOrder.CumQuoteQty, 64)
			paladin.ModHP = int(hp*1000) - paladin.RawHP
			paladin.InStatus = inOrder.Status
		}
		if p.OutID > 0 {
			outOrder := orders[p.OutID]
			if outOrder.Status == "FILLED" {
				battle := Battle{
					PaladinID: paladin.ID,
					Symbol:    paladin.Symbol,
					InID:      p.InID,
					OutID:     p.OutID,
				}
				battles = append(battles, battle)
				b.storage.SaveBattle(context.Background(), &battle)

				hp, _ := strconv.ParseFloat(outOrder.CumQuoteQty, 64)
				paladin.ModHP = int(hp*1000) - p.RawHP
				paladin.Symbol = ""
				paladin.InID = 0
				paladin.InPrice = ""
				paladin.InDate = 0
				paladin.InStatus = ""

				p.ModHP = paladin.ModHP
				p.InID = 0
				p.OutID = 0
				b.storage.SavePaladin(context.Background(), &p)
			} else {
				paladin.OutID = outOrder.OrderID
				paladin.OutPrice = outOrder.Price
				paladin.OutStatus = outOrder.Status
			}
		}

		logs, win := getBattleLogs(battles, orders, p.ID)
		paladin.WinLoss = fmt.Sprintf("%d/%d", win, len(logs))
		paladin.Logs = logs
		rs = append(rs, paladin)
	}
	sort.Sort(PaladinSort(rs))
	b.cache = rs
	return b.cache
}

func getBattleLogs(battles []Battle, orders map[int64]bin.GetOrderResponse, paladinID string) (logs []BattleResponse, win int) {
	logs = []BattleResponse{}
	win = 0
	for _, battle := range battles {
		if battle.PaladinID != paladinID {
			continue
		}
		inOrder := orders[battle.InID]
		outOrder := orders[battle.OutID]
		from, _ := strconv.ParseFloat(inOrder.CumQuoteQty, 64)
		to, _ := strconv.ParseFloat(outOrder.CumQuoteQty, 64)
		if from < to {
			win++
		}
		log := BattleResponse{
			Symbol:   inOrder.Symbol,
			Duration: outOrder.UpdateTime - inOrder.Time,
			Percent:  (to - from) * 100 / from,
		}
		logs = append([]BattleResponse{log}, logs...)
	}
	return
}
