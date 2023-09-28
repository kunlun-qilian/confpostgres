package confpostgres

import (
	"context"

	"github.com/kunlun-qilian/sqlx/v2"
)

type contextKeyDBExecutor int

func WithDB(db sqlx.DBExecutor) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyDBExecutor(0), db)
	}
}

func FromContext(ctx context.Context) sqlx.DBExecutor {
	return ctx.Value(contextKeyDBExecutor(0)).(sqlx.DBExecutor).WithContext(ctx)
}

func SlaveFromContext(ctx context.Context) sqlx.DBExecutor {
	return SwitchSlave(ctx.Value(contextKeyDBExecutor(0)).(sqlx.DBExecutor)).WithContext(ctx)
}
