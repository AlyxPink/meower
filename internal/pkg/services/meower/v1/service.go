package meower

import (
	v1 "github.com/AlyxPink/meower/internal/pkg/api/meower/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

type meower struct {
	v1.UnimplementedMeowerSvcServer
	db *pgxpool.Pool
}

func Meower(db *pgxpool.Pool) v1.MeowerSvcServer {
	return &meower{db: db}
}
