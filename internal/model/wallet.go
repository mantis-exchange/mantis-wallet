package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AddressType string

const (
	AddressDeposit  AddressType = "DEPOSIT"
	AddressHotWallet AddressType = "HOT"
	AddressColdWallet AddressType = "COLD"
)

type Address struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	Chain     string      `json:"chain"`
	Address   string      `json:"address"`
	Type      AddressType `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
}

type DepositStatus string

const (
	DepositPending   DepositStatus = "PENDING"
	DepositConfirmed DepositStatus = "CONFIRMED"
	DepositCredited  DepositStatus = "CREDITED"
)

type Deposit struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	Chain         string        `json:"chain"`
	Asset         string        `json:"asset"`
	TxHash        string        `json:"tx_hash"`
	Amount        string        `json:"amount"`
	Confirmations int           `json:"confirmations"`
	Status        DepositStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type WithdrawalStatus string

const (
	WithdrawalPending   WithdrawalStatus = "PENDING"
	WithdrawalApproved  WithdrawalStatus = "APPROVED"
	WithdrawalSubmitted WithdrawalStatus = "SUBMITTED"
	WithdrawalCompleted WithdrawalStatus = "COMPLETED"
	WithdrawalFailed    WithdrawalStatus = "FAILED"
	WithdrawalRejected  WithdrawalStatus = "REJECTED"
)

type Withdrawal struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	Chain     string           `json:"chain"`
	Asset     string           `json:"asset"`
	ToAddress string           `json:"to_address"`
	Amount    string           `json:"amount"`
	Fee       string           `json:"fee"`
	TxHash    string           `json:"tx_hash"`
	Status    WithdrawalStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

func (r *WalletRepo) CreateAddress(ctx context.Context, a *Address) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO addresses (id, user_id, chain, address, type, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		a.ID, a.UserID, a.Chain, a.Address, a.Type, a.CreatedAt,
	)
	return err
}

func (r *WalletRepo) GetDepositAddress(ctx context.Context, userID uuid.UUID, chain string) (*Address, error) {
	a := &Address{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, chain, address, type, created_at
		 FROM addresses WHERE user_id = $1 AND chain = $2 AND type = 'DEPOSIT'`, userID, chain,
	).Scan(&a.ID, &a.UserID, &a.Chain, &a.Address, &a.Type, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *WalletRepo) CreateDeposit(ctx context.Context, d *Deposit) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO deposits (id, user_id, chain, asset, tx_hash, amount, confirmations, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (tx_hash) DO NOTHING`,
		d.ID, d.UserID, d.Chain, d.Asset, d.TxHash, d.Amount, d.Confirmations, d.Status, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *WalletRepo) UpdateDepositStatus(ctx context.Context, id uuid.UUID, status DepositStatus, confirmations int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE deposits SET status = $1, confirmations = $2, updated_at = $3 WHERE id = $4`,
		status, confirmations, time.Now(), id,
	)
	return err
}

func (r *WalletRepo) CreateWithdrawal(ctx context.Context, w *Withdrawal) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO withdrawals (id, user_id, chain, asset, to_address, amount, fee, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		w.ID, w.UserID, w.Chain, w.Asset, w.ToAddress, w.Amount, w.Fee, w.Status, w.CreatedAt, w.UpdatedAt,
	)
	return err
}

func (r *WalletRepo) UpdateWithdrawalStatus(ctx context.Context, id uuid.UUID, status WithdrawalStatus, txHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE withdrawals SET status = $1, tx_hash = $2, updated_at = $3 WHERE id = $4`,
		status, txHash, time.Now(), id,
	)
	return err
}

func (r *WalletRepo) ListWithdrawalsByStatus(ctx context.Context, status WithdrawalStatus, limit int) ([]Withdrawal, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, chain, asset, to_address, amount, fee, tx_hash, status, created_at, updated_at
		 FROM withdrawals WHERE status = $1 ORDER BY created_at LIMIT $2`, status, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ws []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.Chain, &w.Asset, &w.ToAddress, &w.Amount, &w.Fee, &w.TxHash, &w.Status, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, nil
}
