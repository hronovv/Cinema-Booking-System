package booking

import (
	"cinemabooking/internal/adapters"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

func TestConcurrentBooking(t *testing.T) {
	store := NewRedisStore(adapters.NewClient("localhost:6379"))
	service := NewService(store)

	const NumGoroutines int = 1e5
	var (
		successes atomic.Int64
		failures  atomic.Int64
		wg        sync.WaitGroup
	)
	wg.Add(NumGoroutines)
	for i := range NumGoroutines {
		go func(userNumber int) {
			defer wg.Done()
			_, err := service.Book(Booking{
				MovieID: "screen-1",
				SeatID:  "A5",
				UserID:  uuid.New().String(),
			})
			if err == nil {
				successes.Add(1)
			} else {
				failures.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if res := successes.Load(); res != 1 {
		t.Errorf("expected exactly 1 correct booking, instead got %d", res)
	}
	if res := failures.Load(); res != int64(NumGoroutines)-1 {
		t.Errorf("expected %d failures, got %d", NumGoroutines-1, res)
	}
}
