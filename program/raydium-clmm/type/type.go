package type_

import (
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

type SwapV2Keys struct {
	AmmConfig        solana.PublicKey
	PairAddress      solana.PublicKey
	Vaults           map[solana.PublicKey]solana.PublicKey
	ObservationState solana.PublicKey
	ExBitmapAccount  solana.PublicKey
	RemainAccounts   []solana.PublicKey
}

type SwapKeys struct {
	AmmConfig        solana.PublicKey
	PairAddress      solana.PublicKey
	Vaults           map[solana.PublicKey]solana.PublicKey
	ObservationState solana.PublicKey
	TickArrayAccount solana.PublicKey
	RemainAccounts   []solana.PublicKey
}

type PoolInfo struct {
	Id                  uint64
	Bump                [1]uint8
	AmmConfig           solana.PublicKey
	Owner               solana.PublicKey
	TokenMint0          solana.PublicKey
	TokenMint1          solana.PublicKey
	TokenVault0         solana.PublicKey
	TokenVault1         solana.PublicKey
	ObservationKey      solana.PublicKey
	MintDecimals0       uint8
	MintDecimals1       uint8
	TickSpacing         uint16
	Liquidity           bin.Uint128
	SqrtPriceX64        bin.Uint128
	TickCurrent         int32
	Padding3            uint16
	Padding4            uint16
	FeeGrowthGlobal0X64 bin.Uint128
	FeeGrowthGlobal1X64 bin.Uint128
	ProtocolFeesToken0  uint64
	ProtocolFeesToken1  uint64
	SwapInAmountToken0  bin.Uint128
	SwapOutAmountToken1 bin.Uint128
	SwapInAmountToken1  bin.Uint128
	SwapOutAmountToken0 bin.Uint128
}
