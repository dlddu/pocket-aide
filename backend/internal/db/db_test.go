package db_test

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/db"
)

func TestOpenKeepsOneConnectionSoInProcessReadersAndWritersNeverSeeSQLiteBusy(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "busy.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if got := conn.Stats().MaxOpenConnections; got != 1 {
		t.Fatalf("MaxOpenConnections = %d, want 1", got)
	}
	if _, err := conn.Exec(`CREATE TABLE busy_probe (id INTEGER PRIMARY KEY, worker INTEGER NOT NULL, n INTEGER NOT NULL)`); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(time.Second)
	errs := make(chan error, 1)
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for n := 0; time.Now().Before(deadline); n++ {
				var count int
				if err := conn.QueryRow(`SELECT count(*) FROM busy_probe WHERE worker = ?`, w).Scan(&count); err != nil {
					select {
					case errs <- fmt.Errorf("worker %d read: %w", w, err):
					default:
					}
					return
				}
				if _, err := conn.Exec(`INSERT INTO busy_probe (worker, n) VALUES (?, ?)`, w, n); err != nil {
					select {
					case errs <- fmt.Errorf("worker %d write: %w", w, err):
					default:
					}
					return
				}
			}
		}(w)
	}
	wg.Wait()
	select {
	case err := <-errs:
		t.Fatal(err)
	default:
	}
}
