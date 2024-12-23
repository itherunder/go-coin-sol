package type_

import "github.com/gagliardetto/solana-go"

type RaydiumSwapKeys struct {
	AmmAddress                   solana.PublicKey
	AmmOpenOrdersAddress         *solana.PublicKey
	AmmTargetOrdersAddress       *solana.PublicKey
	PoolCoinTokenAccountAddress  solana.PublicKey
	PoolPcTokenAccountAddress    solana.PublicKey
	SerumProgramAddress          *solana.PublicKey
	SerumMarketAddress           *solana.PublicKey
	SerumBidsAddress             *solana.PublicKey
	SerumAsksAddress             *solana.PublicKey
	SerumEventQueueAddress       *solana.PublicKey
	SerumCoinVaultAccountAddress *solana.PublicKey
	SerumPcVaultAccountAddress   *solana.PublicKey
	SerumVaultSignerAddress      *solana.PublicKey
}
