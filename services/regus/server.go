package regus

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bakaoh/lavato/assets"
	bin "github.com/bakaoh/lavato/plugins/binance"
	bfn "github.com/bakaoh/lavato/plugins/bitfinex"
	"github.com/bakaoh/lavato/private"
	"github.com/bakaoh/lavato/services/doll"
	"github.com/bakaoh/sqlite-gobroem/gobroem"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

// Server implements Regus service
type Server struct {
	upgrader websocket.Upgrader
	binance  *bin.Client
	provider *Provider
	storage  *Storage
	barrack  *Barrack
	crawler  *doll.Crawler
}

// NewServer creates a new ws.Server
func NewServer() *Server {
	binance, err := bin.NewClient(
		private.BinanceApiKey,
		private.BinanceSecretKey,
	)
	if err != nil {
		log.Fatal("can not connect Binance: ", err)
	}
	storage, err := NewStorage(viper.GetString("regus.db"))
	if err != nil {
		log.Fatal("can not open storage: ", err)
	}

	provider := NewProvider(binance)
	barrack := NewBarrack(binance, provider, storage)

	bitfinex, err := bfn.NewClient("", "")
	if err != nil {
		log.Fatal("can not connect Binance: ", err)
	}
	crawler := doll.NewCrawler(bitfinex, "ethusd")

	return &Server{
		binance:  binance,
		provider: provider,
		storage:  storage,
		barrack:  barrack,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		crawler: crawler,
	}
}

// Run starts the storage server
func (s *Server) Run(addr string) error {
	fileServer := http.FileServer(&assets.AssetFS{
		Asset:    assets.Asset,
		AssetDir: assets.AssetDir,
		Prefix:   "assets/public_html",
	})
	http.HandleFunc("/regus/paladins", s.paladins)
	http.HandleFunc("/regus/action", s.action)
	http.HandleFunc("/regus/action/info", s.actionInfo)
	http.HandleFunc("/regus/ws", s.ws)
	http.Handle("/lavato/", http.StripPrefix("/lavato/", fileServer))

	browser, err := gobroem.NewAPI(viper.GetString("regus.db"))
	if err == nil {
		http.Handle("/regus/browser/", browser.Handler("/regus/browser/", "/regus/browser/"))
	}

	go s.provider.Run()
	go s.crawler.Run()
	defer s.crawler.Stop()
	defer s.provider.Stop()
	defer s.storage.Close()

	return http.ListenAndServe(addr, nil)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func responseJSON(w *http.ResponseWriter, data interface{}) {
	rs, err := json.Marshal(data)
	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		return
	}
	(*w).WriteHeader(http.StatusOK)
	(*w).Write(rs)
}

func (s *Server) paladins(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	responseJSON(&w, s.barrack.GetPaladins(context.Background()))
}

func (s *Server) action(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	symbol := r.URL.Query().Get("symbol")
	act := r.URL.Query().Get("act")
	if err := s.barrack.Action(context.Background(), id, symbol, act); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		responseJSON(&w, s.barrack.GetPaladins(context.Background()))
	}
}

func (s *Server) actionInfo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	symbol := r.URL.Query().Get("symbol")
	act := r.URL.Query().Get("act")
	if info, err := s.barrack.ActionInfo(context.Background(), id, symbol, act); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		responseJSON(&w, info)
	}
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	s.barrack.ShouldFullUpdate()
	s.onTick(c)
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			err := s.onTick(c)
			if err != nil {
				return
			}
		}
	}
}

func (s *Server) onTick(c *websocket.Conn) error {
	data, err := json.Marshal(s.barrack.OnTick())
	err = c.WriteMessage(1, data)
	if err != nil {
		log.Println("write:", err)
		return err
	}
	return nil
}
