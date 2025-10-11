package health

import (
	"context"
)

type HealthRepository interface {
	GetAllHealth(ctx context.Context) ([]*Health, error)
	BatchUpdateHealth(ctx context.Context, statuses []Health) error
}
