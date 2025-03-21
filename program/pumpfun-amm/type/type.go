package type_

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	pumpfun_amm_constant "github.com/pefish/go-coin-sol/program/pumpfun-amm/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pkg/errors"
)

type SwapKeys struct {
	AmmAddress         solana.PublicKey
	BaseTokenAddress   solana.PublicKey
	QuoteTokenAddress  solana.PublicKey
	BaseTokenDecimals  uint64
	QuoteTokenDecimals uint64
}

func (t *SwapKeys) BaseVault() (solana.PublicKey, error) {
	poolBaseTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		t.AmmAddress,
		t.BaseTokenAddress,
	)
	if err != nil {
		return solana.PublicKey{}, errors.Wrapf(err, "")
	}

	return poolBaseTokenAccount, nil
}

func (t *SwapKeys) QuoteVault() (solana.PublicKey, error) {
	poolQuoteTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		t.AmmAddress,
		t.QuoteTokenAddress,
	)
	if err != nil {
		return solana.PublicKey{}, errors.Wrapf(err, "")
	}

	return poolQuoteTokenAccount, nil
}

func (t *SwapKeys) ToSwapAccounts(
	network rpc.Cluster,
	userAddress solana.PublicKey,
) ([]*solana.AccountMeta, error) {
	userBaseTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		t.BaseTokenAddress,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, t.BaseTokenAddress)
	}

	userQuoteTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		t.QuoteTokenAddress,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, t.QuoteTokenAddress)
	}

	baseVault, err := t.BaseVault()
	if err != nil {
		return nil, err
	}
	quoteVault, err := t.QuoteVault()
	if err != nil {
		return nil, err
	}

	return []*solana.AccountMeta{
		{
			PublicKey:  t.AmmAddress,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  userAddress,
			IsSigner:   true,
			IsWritable: true,
		},
		{
			PublicKey:  pumpfun_amm_constant.Global_Config[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  t.BaseTokenAddress,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  t.QuoteTokenAddress,
			IsSigner:   false,
			IsWritable: false,
		},

		{
			PublicKey:  userBaseTokenAccount,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  userQuoteTokenAccount,
			IsSigner:   false,
			IsWritable: true,
		},

		{
			PublicKey:  baseVault,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  quoteVault,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  pumpfun_amm_constant.Protocol_Fee_Recipient[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  pumpfun_amm_constant.Protocol_Fee_Recipient_WSOL_Token_Account[network],
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.SystemProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.SPLAssociatedTokenAccountProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  pumpfun_amm_constant.Event_Authority[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  pumpfun_amm_constant.Pumpfun_AMM_Program[network],
			IsSigner:   false,
			IsWritable: false,
		},
	}, nil
}

type ExtraDatasType struct {
	ReserveInputWithDecimals  uint64 `json:"reserve_input_with_decimals"`
	ReserveOutputWithDecimals uint64 `json:"reserve_output_with_decimals"`
}

type AddLiqTxDataType struct {
	TxId                        string `json:"txid"`
	InitBaseAmountWithDecimals  uint64 `json:"init_base_amount_with_decimals"`
	InitQuoteAmountWithDecimals uint64 `json:"init_quote_amount_with_decimals"`

	SwapKeys

	FeeInfo *type_.FeeInfo `json:"fee_info"`
}
