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

type SellInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewSellBaseInInstruction(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	minReceiveSOLAmountWithDecimals uint64,
	swapKeys pumpfun_amm_type.SwapKeys,
) (*SellInstruction, error) {
	methodBytes, err := hex.DecodeString("33e685a4017f83ad")
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
	)
	if err != nil {
		return nil, err
	}
	return &SellInstruction{
		accounts:  accounts,
		data:      append(methodBytes, params.Bytes()...),
		programID: pumpfun_amm_constant.Pumpfun_AMM_Program[network],
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
