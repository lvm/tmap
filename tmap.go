package tmap

import (
	"context"
	"sync"
	"time"
)

type (
	Item interface {
		GetID() string
	}

	ITMap interface {
		Load(ctx context.Context, itemID string) (Item, error)
		Store(ctx context.Context, item Item) error
		Swap(ctx context.Context, item Item) error
		Delete(ctx context.Context, itemID string) error
		Range(ctx context.Context, f func(k, v any) bool) ([]Item, error)

		Truncate(ctx context.Context) error
		Flush(ctx context.Context, ticker *time.Ticker)
	}

	ErrNotFound struct{}

	Wrapper struct {
		item Item
		ttl  time.Time
	}

	TMap struct {
		items sync.Map
		ttl   time.Duration
		nowFn func() time.Time
	}
)

var errNotFound ErrNotFound = ErrNotFound{}

func NewTMap(ttl time.Duration, nowFn func() time.Time) *TMap {
	m := &TMap{
		ttl:   ttl,
		nowFn: nowFn,
	}

	return m
}

func (e ErrNotFound) Error() string {
	return "Not found"
}

func (m *TMap) getItemTTL(itemID string) time.Time {
	if value, ok := m.items.Load(itemID); ok {
		return value.(Wrapper).ttl
	}

	return time.Time{}
}

func (m *TMap) Load(ctx context.Context, itemID string) (Item, error) {
	if value, ok := m.items.Load(itemID); ok {
		return value.(Wrapper).item, nil
	}

	return nil, errNotFound
}

func (m *TMap) Range(ctx context.Context, f func(k, v any) bool) ([]Item, error) {
	var items []Item = make([]Item, 0)
	m.items.Range(func(_, value any) bool {
		item := value.(Wrapper).item
		if f(nil, item) {
			items = append(items, item)
		}
		return true
	})

	return items, nil
}

func (m *TMap) Store(ctx context.Context, item Item) error {
	m.items.Store(item.GetID(), Wrapper{
		item: item,
		ttl:  m.nowFn().Add(m.ttl),
	})

	return nil
}

func (m *TMap) Swap(ctx context.Context, item Item) error {
	if _, err := m.Load(ctx, item.GetID()); err != nil {
		return errNotFound
	}
	m.items.Swap(item.GetID(), Wrapper{
		item: item,
		ttl:  m.nowFn().Add(m.ttl),
	})
	return nil
}

func (m *TMap) Delete(ctx context.Context, itemID string) error {
	if _, err := m.Load(ctx, itemID); err != nil {
		return errNotFound
	}
	m.items.Delete(itemID)

	return nil
}

func (m *TMap) Truncate(ctx context.Context) error {
	var err error
	m.Range(ctx, func(_, v any) bool {
		err = m.Delete(ctx, v.(Item).GetID())
		return true
	})

	return err
}

func (m *TMap) Flush(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			m.Range(ctx, func(_, v any) bool {
				key := v.(Item).GetID()
				ttl := m.getItemTTL(key)
				if !ttl.IsZero() && m.nowFn().After(ttl) {
					_ = m.Delete(ctx, key)
				}
				return true
			})
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
