package storage

import "time"

type HistoryRecord struct {
	ID        int
	Input1    int
	Input2    int
	Result    int
	Operation string
	CreatedAt time.Time
}
