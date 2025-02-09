package instruction

import (
	"bytes"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	raydium_proxy_constant "github.com/pefish/go-coin-sol/program/raydium-amm-proxy/constant"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	raydium_type "github.com/pefish/go-coin-sol/program/raydium-amm/type"
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
	userWSOLAssociatedAccount solana.PublicKey,
	userTokenAssociatedAccount solana.PublicKey,
	tokenAmountWithDecimals uint64,
	minReceiveSOLAmountWithDecimals uint64,
	raydiumSwapKeys raydium_type.RaydiumSwapKeys,
	coinIsSOL bool,
) (*SellInstruction, error) {
	params := new(bytes.Buffer)
	err := bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals uint64
		MinReceiveSOLAmount     uint64
	}{
		TokenAmountWithDecimals: tokenAmountWithDecimals,
		MinReceiveSOLAmount:     minReceiveSOLAmountWithDecimals,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	var coinAccount solana.PublicKey
	var pcAccount solana.PublicKey
	if coinIsSOL {
		coinAccount = solana.SolMint
		pcAccount = tokenAddress
	} else {
		coinAccount = tokenAddress
		pcAccount = solana.SolMint
	}
	return &SellInstruction{
		accounts: []*solana.AccountMeta{
			{
				PublicKey:  raydium_constant.Raydium_Liquidity_Pool_V4[network],
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  raydiumSwapKeys.AmmAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  raydium_constant.Raydium_Authority_V4[network],
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  raydiumSwapKeys.PoolCoinTokenAccountAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  raydiumSwapKeys.PoolPcTokenAccountAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  coinAccount,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  pcAccount,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  userTokenAssociatedAccount,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  userWSOLAssociatedAccount,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  userAddress,
				IsSigner:   true,
				IsWritable: false,
			},
			{
				PublicKey:  solana.TokenProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
		},
		data: append([]byte{
			250,
			174,
			212,
			217,
			47,
			84,
			212,
			231,
		}, params.Bytes()...),
		programID: raydium_proxy_constant.Proxy[network],
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
