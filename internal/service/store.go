package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"qris-payment/internal/model"
)

type Store struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]*model.Payment
}

func NewStore(filePath string) (*Store, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	s := &Store{
		filePath: filePath,
		data:     make(map[string]*model.Payment),
	}

	if _, err := os.Stat(filePath); err == nil {
		if err := s.load(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Store) Save(p *model.Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[p.OrderID] = p
	return s.flush()
}

func (s *Store) Get(orderID string) (*model.Payment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.data[orderID]
	return p, ok
}

func (s *Store) UpdateStatus(orderID, transactionID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.data[orderID]
	if !ok {
		return nil
	}

	switch status {
	case "settlement", "capture":
		p.Status = model.StatusSuccess
	case "deny", "cancel", "failure":
		p.Status = model.StatusFailed
	case "expire":
		p.Status = model.StatusExpired
	default:
		p.Status = model.StatusPending
	}

	if transactionID != "" {
		p.TransactionID = transactionID
	}
	p.UpdatedAt = time.Now()

	return s.flush()
}

func (s *Store) All() []*model.Payment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*model.Payment, 0, len(s.data))
	for _, p := range s.data {
		list = append(list, p)
	}
	return list
}

func (s *Store) flush() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.data)
}
