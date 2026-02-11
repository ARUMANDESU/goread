package dbx

import (
	"context"
)

// Session aims at facilitating business transactions while abstracting the underlying mechanism,
// be it a database transaction or another transaction mechanism. This allows services to execute
// multiple business use-cases and easily rollback changes in case of error, without creating a
// dependency to the database layer.
//
// Sessions should be constituted of a root session created with a "New"-type constructor and allow
// the creation of child sessions with `Begin()` and `Transaction()`. Nested transactions should be supported
// as well.
//
// link: https://stackoverflow.com/a/78104408/16499540
type Session interface {
	// Begin returns a new session with the given context and a started transaction.
	// Using the returned session should have no side-effect on the parent session.
	// The underlying transaction mechanism is injected as a value into the new session's context.
	Begin(ctx context.Context) (Session, error)

	// Transaction executes a transaction. If the given function returns an error, the transaction
	// is rolled back. Otherwise it is automatically committed before `Transaction()` returns.
	// The underlying transaction mechanism is injected into the context as a value.
	Transaction(ctx context.Context, f func(context.Context) error) error

	// Rollback the changes in the transaction. This action is final.
	Rollback() error

	// Commit the changes in the transaction. This action is final.
	Commit() error

	// Context returns the session's context. If it's the root session, `context.Background()` is returned.
	// If it's a child session started with `Begin()`, then the context will contain the associated
	// transaction mechanism as a value.
	Context() context.Context
}

// type PgxDB interface {
// 	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
// 	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
// 	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
// 	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
// }
//
// // Pgx retrieves the pgx.Tx from the context if it exists, otherwise it returns the provided fallback pgxpool.Pool.
// func Pgx(ctx context.Context, fallback *pgxpool.Pool) PgxDB {
// 	tx, err := ctxs.TxFromCtx[pgx.Tx](ctx)
// 	if err != nil {
// 		return fallback
// 	}
// 	return tx
// }
//
// // PgxTx retrieves the pgx.Tx from the context if it exists, otherwise it starts a new transaction
// // using the provided fallback pgxpool.Pool. The provided function fn is executed within the transaction.
// // If fn returns an error, the transaction is rolled back. Otherwise, it is committed.
// func PgxTx(ctx context.Context, fallback *pgxpool.Pool, fn func(pgx.Tx) error) error {
// 	const op = "db.PgxTx"
// 	tx, err := ctxs.TxFromCtx[pgx.Tx](ctx)
// 	if err == nil {
// 		return fn(tx)
// 	}
//
// 	tx, err = fallback.Begin(ctx)
// 	if err != nil {
// 		return errorx.Wrap(err, op)
// 	}
// 	defer func() {
// 		if p := recover(); p != nil {
// 			_ = tx.Rollback(ctx)
// 			panic(p)
// 		}
// 	}()
//
// 	err = fn(tx)
// 	if err != nil {
// 		rbErr := tx.Rollback(ctx)
// 		if rbErr != nil {
// 			return fmt.Errorf("%s: failed to rollback transaction after error: %v: original error: %w", op, rbErr, err)
// 		}
// 		return errorx.Wrap(err, op)
// 	}
//
// 	return errorx.Wrap(tx.Commit(ctx), op)
// }
//
// type PgxSession struct {
// 	TxOptions pgx.TxOptions
// 	pgxPool   *pgxpool.Pool
// 	pgxTx     pgx.Tx
// 	ctx       context.Context
// }
//
// func NewPgxSession(pgxPool *pgxpool.Pool, txOptions pgx.TxOptions, ctx context.Context) *PgxSession {
// 	return &PgxSession{
// 		pgxPool:   pgxPool,
// 		TxOptions: txOptions,
// 		ctx:       ctx,
// 	}
// }
//
// func (s *PgxSession) Begin(ctx context.Context) (Session, error) {
// 	const op = "db.PgxSession.Begin"
// 	tx, err := s.pgxPool.BeginTx(ctx, s.TxOptions)
// 	if err != nil {
// 		return nil, errorx.Wrap(err, op)
// 	}
//
// 	return &PgxSession{
// 		TxOptions: s.TxOptions,
// 		pgxPool:   s.pgxPool,
// 		pgxTx:     tx,
// 		ctx:       ctx,
// 	}, nil
// }
//
// func (s *PgxSession) Rollback() error {
// 	const op = "db.PgxSession.Rollback"
// 	if s.pgxTx == nil {
// 		return errorx.NewWithOp("no transaction to rollback", op)
// 	}
// 	return errorx.Wrap(s.pgxTx.Rollback(s.ctx), op)
// }
//
// func (s *PgxSession) Commit() error {
// 	const op = "db.PgxSession.Commit"
// 	if s.pgxTx == nil {
// 		return errorx.NewWithOp("no transaction to commit", op)
// 	}
// 	return errorx.Wrap(s.pgxTx.Commit(s.ctx), op)
// }
//
// func (s *PgxSession) Context() context.Context {
// 	return s.ctx
// }
//
// func (s *PgxSession) Transaction(ctx context.Context, f func(context.Context) error) error {
// 	const op = "db.PgxSession.Transaction"
// 	tx, err := s.pgxPool.BeginTx(ctx, s.TxOptions)
// 	if err != nil {
// 		return errorx.Wrap(err, op)
// 	}
//
// 	txCtx := ctxs.WithTx(ctx, tx)
// 	err = f(txCtx)
// 	if err != nil {
// 		rbErr := tx.Rollback(txCtx)
// 		if rbErr != nil {
// 			return fmt.Errorf("%s: failed to rollback transaction after error: %v: original error: %w", op, rbErr, err)
// 		}
// 		return errorx.Wrap(err, op)
// 	}
//
// 	return errorx.Wrap(tx.Commit(txCtx), op)
// }
