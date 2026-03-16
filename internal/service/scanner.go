package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mantis-exchange/mantis-wallet/internal/chain"
	"github.com/mantis-exchange/mantis-wallet/internal/client"
	"github.com/mantis-exchange/mantis-wallet/internal/model"
)

const (
	requiredConfirmations = 6
	scanInterval          = 10 * time.Second
	nativeAsset           = "QFC"
)

// DepositScanner polls the chain for new deposits to known addresses
// and credits user balances via the account service.
type DepositScanner struct {
	repo      *model.WalletRepo
	eth       *chain.EthereumClient
	account   *client.AccountClient
	lastBlock uint64
}

func NewDepositScanner(repo *model.WalletRepo, eth *chain.EthereumClient, account *client.AccountClient) *DepositScanner {
	return &DepositScanner{repo: repo, eth: eth, account: account}
}

// Start begins the deposit scanning loop. Intended to run in a goroutine.
func (s *DepositScanner) Start() {
	log.Printf("deposit scanner started")

	ctx := context.Background()
	latest, err := s.eth.GetBlockNumber(ctx)
	if err != nil {
		log.Printf("deposit scanner: failed to get initial block number: %v", err)
	} else if latest > requiredConfirmations {
		s.lastBlock = latest - requiredConfirmations
	}

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.scan(ctx)
	}
}

func (s *DepositScanner) scan(ctx context.Context) {
	latest, err := s.eth.GetBlockNumber(ctx)
	if err != nil {
		log.Printf("deposit scanner: failed to get block number: %v", err)
		return
	}

	if latest <= requiredConfirmations {
		return
	}
	confirmedBlock := latest - requiredConfirmations

	if confirmedBlock <= s.lastBlock {
		return
	}

	// Build lookup of deposit addresses
	addresses, err := s.repo.ListDepositAddresses(ctx)
	if err != nil {
		log.Printf("deposit scanner: failed to list addresses: %v", err)
		return
	}
	if len(addresses) == 0 {
		s.lastBlock = confirmedBlock
		return
	}

	addrMap := make(map[string]*model.Address)
	for i := range addresses {
		addrMap[strings.ToLower(addresses[i].Address)] = &addresses[i]
	}

	// Scan new blocks
	for blockNum := s.lastBlock + 1; blockNum <= confirmedBlock; blockNum++ {
		block, err := s.eth.GetBlockByNumber(ctx, blockNum)
		if err != nil {
			log.Printf("deposit scanner: failed to get block %d: %v", blockNum, err)
			return // retry from this block next tick
		}

		for _, tx := range block.Transactions {
			if tx.To == "" || tx.Value == "" || tx.Value == "0x0" {
				continue
			}

			addr, ok := addrMap[strings.ToLower(tx.To)]
			if !ok {
				continue
			}

			amount := chain.ParseWeiToEther(tx.Value)
			if amount == "0" {
				continue
			}

			now := time.Now()
			deposit := &model.Deposit{
				ID:            uuid.New(),
				UserID:        addr.UserID,
				Chain:         addr.Chain,
				Asset:         nativeAsset,
				TxHash:        tx.Hash,
				Amount:        amount,
				Confirmations: int(latest - blockNum),
				Status:        model.DepositConfirmed,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := s.repo.CreateDeposit(ctx, deposit); err != nil {
				// ON CONFLICT DO NOTHING handles duplicates
				continue
			}

			// Credit balance via account service
			if err := s.account.CreditBalance(ctx, addr.UserID.String(), nativeAsset, amount); err != nil {
				log.Printf("deposit scanner: failed to credit %s %s to %s: %v", amount, nativeAsset, addr.UserID, err)
				continue
			}

			if err := s.repo.UpdateDepositStatus(ctx, deposit.ID, model.DepositCredited, deposit.Confirmations); err != nil {
				log.Printf("deposit scanner: failed to update deposit status: %v", err)
			}

			log.Printf("deposit credited: %s %s to user %s (tx: %s)", amount, nativeAsset, addr.UserID, tx.Hash)
		}

		s.lastBlock = blockNum
	}
}
