package sol_fi

import (
	"encoding/hex"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	sol_fi_constant "github.com/itherunder/go-coin-sol/program/sol-fi/constant"
	sol_fi_type "github.com/itherunder/go-coin-sol/program/sol-fi/type"
	type_ "github.com/itherunder/go-coin-sol/type"
	"github.com/itherunder/go-coin-sol/util"
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
		if !instruction.ProgramId.Equals(sol_fi_constant.SolFiProgram[network]) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:2]
		if methodId != "07" {
			continue
		}

		userAddress := instruction.Accounts[0]
		pairAddress := instruction.Accounts[1]
		vaultA := instruction.Accounts[2]
		vaultB := instruction.Accounts[3]

		transferDatas, err := util.FindNextTwoTransferDatas(index+1, allInstructions)
		if err != nil {
			return nil, errors.Wrapf(err, "<txid: %s>", txId)
		}
		inputAmountWithDecimals := transferDatas[0].AmountWithDecimals
		outputAmountWithDecimals := transferDatas[1].AmountWithDecimals

		var mintA solana.PublicKey
		var mintB solana.PublicKey
		var mintADecimals uint64
		var mintBDecimals uint64
		for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
				transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(vaultA) {
				mintA = tokenBalanceInfo_.Mint
				mintADecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
			}
			if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
				transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(vaultB) {
				mintB = tokenBalanceInfo_.Mint
				mintBDecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
			}
		}

		var inputVault solana.PublicKey
		var outputVault solana.PublicKey
		var inputAddress solana.PublicKey
		var outputAddress solana.PublicKey
		var inputDecimals uint64
		var outputDecimals uint64
		if transferDatas[0].Destination.Equals(vaultA) {
			// a is input
			inputVault = vaultA
			outputVault = vaultB
			inputAddress = mintA
			outputAddress = mintB
			inputDecimals = mintADecimals
			outputDecimals = mintBDecimals
		} else {
			outputVault = vaultA
			inputVault = vaultB
			inputAddress = mintB
			outputAddress = mintA
			inputDecimals = mintBDecimals
			outputDecimals = mintADecimals
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
			InputAddress:             inputAddress,
			OutputAddress:            outputAddress,
			InputAmountWithDecimals:  inputAmountWithDecimals,
			InputDecimals:            inputDecimals,
			OutputAmountWithDecimals: outputAmountWithDecimals,
			OutputDecimals:           outputDecimals,
			UserAddress:              userAddress,
			ParsedKeys: &sol_fi_type.SwapKeys{
				PairAddress: pairAddress,
				MintA:       mintA,
				MintB:       mintB,
				Vaults: map[solana.PublicKey]solana.PublicKey{
					mintA: vaultA,
					mintB: vaultB,
				},
			},
			ExtraDatas: &sol_fi_type.ExtraDatasType{
				ReserveInputWithDecimals:  reserveInputWithDecimals,
				ReserveOutputWithDecimals: reserveOutputWithDecimals,
			},
			PairAddress: pairAddress,
			InputVault:  inputVault,
			OutputVault: outputVault,
			Keys:        instruction.Accounts,
			AllKeys:     transaction.Message.AccountKeys,
			MethodId:    methodId,
			Program:     sol_fi_constant.SolFiProgram[network],
		})
	}

	return &type_.SwapTxDataType{
		TxId:    txId,
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}
