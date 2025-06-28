package storage

import "context"

type Storage interface {
	Write(ctx context.Context, input1, input2, result int, operation string) error
	GetHistory(ctx context.Context) ([]*HistoryRecord, error)
}
