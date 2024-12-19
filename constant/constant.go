package constant

import "github.com/gagliardetto/solana-go"

const (
	SOL_Decimals = 9
)

var (
	version = uint64(0)

	MaxSupportedTransactionVersion_0 = &version
	Compute_Budget                   = solana.MustPublicKeyFromBase58("ComputeBudget111111111111111111111111111111")
	Rent                             = solana.MustPublicKeyFromBase58("SysvarRent111111111111111111111111111111111")
)
