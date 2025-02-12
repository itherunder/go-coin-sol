package meteora_dlmm

import (
	"encoding/hex"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/discriminator"
	meteora_dlmm_constant "github.com/pefish/go-coin-sol/program/meteora-dlmm/constant"
	meteora_dlmm_type "github.com/pefish/go-coin-sol/program/meteora-dlmm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	"github.com/pkg/errors"
)

func ParseSwapTxByParsedTx(
	network rpc.Cluster,
	meta *rpc.ParsedTransactionMeta,
	transaction *rpc.ParsedTransaction,
) (*type_.SwapTxDataType, error) {
	txId := transaction.Signatures[0].String()
	feeInfo, err := util.GetFeeInfoFromParsedTx(meta, transaction)
	if err != nil {
		return nil, errors.Wrapf(err, "<txid: %s>", txId)
	}
	swaps := make([]*type_.SwapDataType, 0)
	if meta.Err != nil {
		return &type_.SwapTxDataType{
			TxId:    txId,
			Swaps:   swaps,
			FeeInfo: feeInfo,
		}, nil
	}

	allInstructions := make([]*rpc.ParsedInstruction, 0)
	for index, instruction := range transaction.Message.Instructions {
		allInstructions = append(allInstructions, instruction)
		innerInstructions := util.FindInnerInstructionsFromParsedMeta(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		allInstructions = append(allInstructions, innerInstructions...)
	}

	for index, instruction := range allInstructions {
		if !instruction.ProgramId.Equals(meteora_dlmm_constant.Meteora_DLMM[network]) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:16]
		if methodId != discriminator.GetDiscriminator("global", "swap") {
			continue
		}

		userAddress := instruction.Accounts[10]

		pairAddress := instruction.Accounts[0]
		var parsedKeys interface{}

		parsedKeys = &meteora_dlmm_type.SwapKeys{
			PairAddress:    pairAddress,
			VaultX:         instruction.Accounts[2],
			VaultY:         instruction.Accounts[3],
			Oracle:         instruction.Accounts[8],
			XMint:          instruction.Accounts[6],
			YMint:          instruction.Accounts[7],
			RemainAccounts: instruction.Accounts[15:],
			Vaults: map[solana.PublicKey]solana.PublicKey{
				instruction.Accounts[6]: instruction.Accounts[2],
				instruction.Accounts[7]: instruction.Accounts[3],
			},
		}

		transferDatas, err := util.FindNextTwoTransferCheckedDatas(index+1, allInstructions)
		if err != nil {
			return nil, errors.Wrapf(err, "<txid: %s>", txId)
		}
		inputAddress := transferDatas[0].Mint
		outputAddress := transferDatas[1].Mint
		inputAmount := transferDatas[0].AmountWithDecimals
		outputAmount := transferDatas[1].AmountWithDecimals

		var inputVault solana.PublicKey
		var outputVault solana.PublicKey
		if inputAddress.Equals(instruction.Accounts[6]) {
			// x is input
			inputVault = instruction.Accounts[2]
			outputVault = instruction.Accounts[3]
		} else {
			inputVault = instruction.Accounts[3]
			outputVault = instruction.Accounts[2]
		}

		var reserveInputWithDecimals uint64
		var reserveOutputWithDecimals uint64
		for _, tokenBalanceInfo_ := range meta.PostTokenBalances {
			if transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(inputVault) {
				reserveInputWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
			}
			if transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(outputVault) {
				reserveOutputWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
			}
		}

		swaps = append(swaps, &type_.SwapDataType{
			InputAddress:              inputAddress,
			OutputAddress:             outputAddress,
			InputAmountWithDecimals:   inputAmount,
			OutputAmountWithDecimals:  outputAmount,
			UserAddress:               userAddress,
			PairAddress:               pairAddress,
			InputVault:                inputVault,
			OutputVault:               outputVault,
			ReserveInputWithDecimals:  reserveInputWithDecimals,
			ReserveOutputWithDecimals: reserveOutputWithDecimals,
			ParsedKeys:                parsedKeys,
			Keys:                      instruction.Accounts,
			MethodId:                  methodId,
			Program:                   meteora_dlmm_constant.Meteora_DLMM[network],
		})
	}

	return &type_.SwapTxDataType{
		TxId:    txId,
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}
