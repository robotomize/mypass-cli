package manager

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrNotFound = errors.New("record not found")

type CryptFS interface {
	Open() ([]byte, error)
	Write(b []byte) error
}

type Entry struct {
	ID        string
	Title     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChangeEntry struct {
	Title    *string
	Password *string
}

func NewStore(fs CryptFS, txManager *TxManager) *store {
	return &store{fs: fs, txManager: txManager, data: make([]Entry, 0)}
}

type store struct {
	fs CryptFS

	mtx       sync.RWMutex
	data      []Entry
	txManager *TxManager
}

func (s *store) Add(e Entry) error {
	if err := s.txManager.AddTx(e); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *store) DeleteByID(id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByID(id)
	if !ok {
		return ErrNotFound
	}

	if err := s.txManager.DelTx(entry); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *store) DeleteByPos(pos int) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByPosition(pos)
	if !ok {
		return ErrNotFound
	}

	if err := s.txManager.DelTx(entry); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *store) FindByID(id string) (Entry, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.findByID(id)
}

func (s *store) FindByPosition(pos int) (Entry, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	entry, ok := s.findByPosition(pos)
	if !ok {
		return Entry{}, false
	}

	return entry, true
}

func (s *store) Change(pos int, changed ChangeEntry) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByPosition(pos)
	if !ok {
		return ErrNotFound
	}

	if err := s.change(entry.ID, changed); err != nil {
		return fmt.Errorf("change: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *store) load() error {
	b, err := s.fs.Open()
	if err != nil {
		return fmt.Errorf("fs load: %w", err)
	}

	s.txManager.Deserialize(b)

	return nil
}

func (s *store) rebuild() {
	s.data = s.data[:0]
	s.txManager.Each(func(t1 Tx) {
		if t1.Kind == TxKindAdd {
			s.data = append(s.data, t1.Payload)

			return
		}

		if t1.Kind == TxKindDel {
			for idx, entry := range s.data {
				if entry.ID == t1.Payload.ID {
					s.data = append(s.data[:idx], s.data[idx+1:]...)
				}
			}

			return
		}
	})
}

func (s *store) change(id string, changed ChangeEntry) error {
	var entry *Entry

	for _, e := range s.data {
		if e.ID == id {
			entry = &e
			break
		}
	}

	if entry == nil {
		return ErrNotFound
	}

	if changed.Title != nil {
		entry.Title = *changed.Title
	}

	if changed.Password != nil {
		entry.Password = *changed.Password
	}

	entry.UpdatedAt = time.Now().UTC()

	if err := s.txManager.DelTx(*entry); err != nil {
		return fmt.Errorf("del tx: %w", err)
	}

	if err := s.txManager.AddTx(*entry); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	return nil
}

func (s *store) findByID(id string) (Entry, bool) {
	for _, entry := range s.data {
		if entry.ID == id {
			return entry, true
		}
	}

	return Entry{}, false
}

func (s *store) findByPosition(pos int) (Entry, bool) {
	if pos > len(s.data)-1 {
		return Entry{}, false
	}

	return s.data[pos], true
}

func (s *store) sync() error {
	bytes := s.txManager.Serialize()
	if err := s.fs.Write(bytes); err != nil {
		return fmt.Errorf("fs write: %w", err)
	}

	return nil
}
