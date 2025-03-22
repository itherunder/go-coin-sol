package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	raydium_constant "github.com/itherunder/go-coin-sol/program/raydium-amm/constant"
	raydium_type "github.com/itherunder/go-coin-sol/program/raydium-amm/type"
	"github.com/pkg/errors"
)

type SellInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewSellBaseInInstruction(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	minReceiveSOLAmountWithDecimals uint64,
	swapKeys raydium_type.SwapKeys,
) (*SellInstruction, error) {
	methodBytes, err := hex.DecodeString("09")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals uint64
		MinReceiveSOLAmount     uint64
	}{
		TokenAmountWithDecimals: tokenAmountWithDecimals,
		MinReceiveSOLAmount:     minReceiveSOLAmountWithDecimals,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	accounts, err := swapKeys.ToSwapAccounts(
		network,
		userAddress,
		tokenAddress,
		solana.SolMint,
	)
	if err != nil {
		return nil, err
	}
	return &SellInstruction{
		accounts:  accounts,
		data:      append(methodBytes, params.Bytes()...),
		programID: raydium_constant.Raydium_AMM_Program[network],
	}, nil
}

func (t *SellInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *SellInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *SellInstruction) Data() ([]byte, error) {
	return t.data, nil
}
