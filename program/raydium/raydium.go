package raydium

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

func GetSwapInstructions(
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmount string,
	isCloseUserAssociatedTokenAddress bool,
	virtualSolReserve string,
	virtualTokenReserve string,
	slippage uint64,
) ([]solana.Instruction, error) {
	if slippage == 0 {
		slippage = 50 // 0.5%
	}
	instructions := make([]solana.Instruction, 0)
	// TODO

	return instructions, nil
}
