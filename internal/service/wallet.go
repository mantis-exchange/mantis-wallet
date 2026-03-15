package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mantis-exchange/mantis-wallet/internal/chain"
	"github.com/mantis-exchange/mantis-wallet/internal/model"
)

type WalletService struct {
	repo *model.WalletRepo
	eth  *chain.EthereumClient
}

func NewWalletService(repo *model.WalletRepo, eth *chain.EthereumClient) *WalletService {
	return &WalletService{repo: repo, eth: eth}
}

func (s *WalletService) GetOrCreateDepositAddress(ctx context.Context, userID uuid.UUID, chainName string) (*model.Address, error) {
	addr, err := s.repo.GetDepositAddress(ctx, userID, chainName)
	if err == nil {
		return addr, nil
	}

	// Generate new address
	var address, _ string
	switch chainName {
	case "ETH", "ERC20":
		address, _, err = s.eth.GenerateAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ETH address: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}

	addr = &model.Address{
		ID:        uuid.New(),
		UserID:    userID,
		Chain:     chainName,
		Address:   address,
		Type:      model.AddressDeposit,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateAddress(ctx, addr); err != nil {
		return nil, err
	}

	return addr, nil
}

func (s *WalletService) RequestWithdrawal(ctx context.Context, userID uuid.UUID, chainName, asset, toAddress, amount, fee string) (*model.Withdrawal, error) {
	now := time.Now()
	w := &model.Withdrawal{
		ID:        uuid.New(),
		UserID:    userID,
		Chain:     chainName,
		Asset:     asset,
		ToAddress: toAddress,
		Amount:    amount,
		Fee:       fee,
		Status:    model.WithdrawalPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateWithdrawal(ctx, w); err != nil {
		return nil, err
	}

	// TODO: Deduct balance via account service
	return w, nil
}

func (s *WalletService) ProcessPendingWithdrawals(ctx context.Context) error {
	withdrawals, err := s.repo.ListWithdrawalsByStatus(ctx, model.WithdrawalApproved, 10)
	if err != nil {
		return err
	}

	for _, w := range withdrawals {
		var txHash string
		switch w.Chain {
		case "ETH", "ERC20":
			txHash, err = s.eth.SendTransaction(w.ToAddress, w.Amount)
		default:
			continue
		}

		if err != nil {
			_ = s.repo.UpdateWithdrawalStatus(ctx, w.ID, model.WithdrawalFailed, "")
			continue
		}

		_ = s.repo.UpdateWithdrawalStatus(ctx, w.ID, model.WithdrawalSubmitted, txHash)
	}

	return nil
}
