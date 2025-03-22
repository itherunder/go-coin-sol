package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	pumpfun_amm_constant "github.com/itherunder/go-coin-sol/program/pumpfun-amm/constant"
	pumpfun_amm_type "github.com/itherunder/go-coin-sol/program/pumpfun-amm/type"
	"github.com/pkg/errors"
)

type BuyInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewBuyBaseOutInstruction(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	maxCostSolAmountWithDecimals uint64,
	swapKeys pumpfun_amm_type.SwapKeys,
) (*BuyInstruction, error) {
	methodBytes, err := hex.DecodeString("66063d1201daebea")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals      uint64
		MaxCostSolAmountWithDecimals uint64
	}{
		TokenAmountWithDecimals:      tokenAmountWithDecimals,
		MaxCostSolAmountWithDecimals: maxCostSolAmountWithDecimals,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	accounts, err := swapKeys.ToSwapAccounts(
		network,
		userAddress,
	)
	if err != nil {
		return nil, err
	}
	return &BuyInstruction{
		accounts:  accounts,
		data:      append(methodBytes, params.Bytes()...),
		programID: pumpfun_amm_constant.Pumpfun_AMM_Program[network],
	}, nil
}

func (t *BuyInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *BuyInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *BuyInstruction) Data() ([]byte, error) {
	return t.data, nil
}
