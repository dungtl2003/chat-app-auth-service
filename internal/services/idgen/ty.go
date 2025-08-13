package idgen

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/services"
)

type IdGeneratorService interface {
	services.Service
	GenerateId(ctx context.Context) (int64, error)
}
