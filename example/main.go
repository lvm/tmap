package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lvm/tmap"
)

type example struct {
	K string
	V string
}

func (e example) GetID() string {
	return e.K
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := tmap.NewTMap(10*time.Second, time.Now)
	t := time.NewTicker(1 * time.Second)
	go m.Flush(ctx, t)

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" || value == "" {
			http.Error(w, "key and value are required", http.StatusBadRequest)
			return
		}

		if err := m.Store(ctx, example{key, value}); err != nil {
			http.Error(w, "cant store item", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "Stored %s = %s\n", key, value)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key is required", http.StatusBadRequest)
			return
		}

		value, err := m.Load(ctx, key)
		if err != nil {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%s = %s\n", key, value)
	})

	http.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		all, err := m.Range(ctx, func(_, _ any) bool { return true })
		if err != nil {
			http.Error(w, "something went wrong retrieving items", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%+v\n", all)
	})

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
