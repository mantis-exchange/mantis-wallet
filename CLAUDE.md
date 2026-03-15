# mantis-wallet

Mantis Exchange wallet service — deposit/withdrawal, address generation, chain scanning.

## Architecture

- `internal/chain/ethereum.go` — ETH address generation, deposit scanning, transaction sending (placeholder)
- `internal/model/wallet.go` — Address, Deposit, Withdrawal models + PostgreSQL CRUD
- `internal/service/wallet.go` — GetOrCreateDepositAddress, RequestWithdrawal, ProcessPendingWithdrawals
- `internal/handler/handler.go` — REST API handlers

## API

- `GET /api/v1/wallet/deposit-address?user_id=...&chain=ETH`
- `POST /api/v1/wallet/withdraw` — `{user_id, chain, asset, address, amount}`

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `50054` | HTTP server port |
| `DB_URL` | `postgres://...` | PostgreSQL |
| `ETH_NODE` | `http://localhost:8545` | Ethereum JSON-RPC |
| `BTC_NODE` | `http://localhost:18332` | Bitcoin JSON-RPC |
