package tmap

import (
	"context"
	"testing"
	"time"
)

type testItem struct {
	id      string
	deleted bool
}

func (m *testItem) GetID() string {
	return m.id
}

func TestNotFound(t *testing.T) {
	e := ErrNotFound{}
	if e.Error() != errNotFound.Error() {
		t.Errorf("expected %s got %s", e.Error(), errNotFound.Error())
	}
}

func TestLoad(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	key := "item1"
	m.items.Store(key, Wrapper{item: &testItem{id: key}, ttl: time.Time{}})

	got, err := m.Load(ctx, "item1")
	if err != nil {
		t.Error("expected no error, got", err)
	}
	if got.GetID() != key {
		t.Error("expected", key, "got", got.GetID())
	}
}

func TestLoad_NotFound(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	_, err := m.Load(ctx, "notfound")
	if err != errNotFound {
		t.Error("expected ErrNotFound, got", err)
	}
}

func TestStore(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	if err := m.Store(ctx, &testItem{id: "item1"}); err != nil {
		t.Error("expected nil, got", err)
	}
}

func TestSwap(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	key := "item1"
	_ = m.Store(ctx, &testItem{id: key})

	if err := m.Swap(ctx, &testItem{id: key, deleted: true}); err != nil {
		t.Error("expected nil, got", err)
	}
}

func TestSwap_NotFound(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	if err := m.Swap(ctx, &testItem{id: "notfound"}); err != errNotFound {
		t.Error("expected ErrNotFound, got", err)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	key := "item1"
	_ = m.Store(ctx, &testItem{id: key})

	if err := m.Delete(ctx, key); err != nil {
		t.Error("expected nil, got", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	if err := m.Delete(ctx, "notfound"); err != errNotFound {
		t.Error("expected ErrNotFound, got", err)
	}
}

func TestGetItemTTL(t *testing.T) {
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	key := "item1"
	itemTTL := time.Now().Add(1 * time.Hour)
	m.items.Store(key, Wrapper{item: &testItem{id: key}, ttl: itemTTL})

	ttl := m.getItemTTL("item1")
	if !ttl.Equal(itemTTL) {
		t.Errorf("expected %v got %v", itemTTL, ttl)
	}

}

func TestGetItemTTL_Zero(t *testing.T) {
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	ttl := m.getItemTTL("item1")
	if !ttl.IsZero() {
		t.Errorf("expected zero ttl got %v", ttl)
	}

}

func TestRange(t *testing.T) {
	ctx := context.TODO()
	now := time.Now
	m := NewTMap(1*time.Hour, now)

	_ = m.Store(ctx, &testItem{id: "test1"})
	_ = m.Store(ctx, &testItem{id: "test2"})
	_ = m.Store(ctx, &testItem{id: "test3"})

	items, err := m.Range(ctx, func(_, _ any) bool { return true })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 3 {
		t.Error("expected 3, got", len(items))
	}
}

func TestTruncate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	now := time.Now()
	m := NewTMap(1*time.Second, func() time.Time { return now })

	_ = m.Store(ctx, &testItem{id: "item1"})
	_ = m.Store(ctx, &testItem{id: "item2"})
	_ = m.Store(ctx, &testItem{id: "item3"})

	items, err := m.Range(ctx, func(_, _ any) bool { return true })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3, got %d", len(items))
	}

	tick := make(chan time.Time)
	go m.Flush(ctx, &time.Ticker{C: tick})

	m.nowFn = func() time.Time { return now.Add(1 * time.Hour) }
	tick <- m.nowFn()

	cancel()
	err = m.Truncate(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	<-ctx.Done()

	items, err = m.Range(ctx, func(_, _ any) bool { return true })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after expiration, got %d", len(items))
	}
}
