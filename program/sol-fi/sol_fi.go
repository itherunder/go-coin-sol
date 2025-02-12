package sol_fi

import (
	"encoding/hex"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	sol_fi_constant "github.com/pefish/go-coin-sol/program/sol-fi/constant"
	sol_fi_type "github.com/pefish/go-coin-sol/program/sol-fi/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
)

func ParseSwapTxByParsedTx(
	network rpc.Cluster,
	meta *rpc.ParsedTransactionMeta,
	transaction *rpc.ParsedTransaction,
) (*type_.SwapTxDataType, error) {
	swaps := make([]*type_.SwapDataType, 0)

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

		transfer1Data, err := util.DecodeTransferInstruction(allInstructions[index+1])
		if err != nil {
			return nil, err
		}
		transfer2Data, err := util.DecodeTransferInstruction(allInstructions[index+2])
		if err != nil {
			return nil, err
		}
		inputAmountWithDecimals := transfer1Data.AmountWithDecimals
		outputAmountWithDecimals := transfer2Data.AmountWithDecimals

		var mintA solana.PublicKey
		var mintB solana.PublicKey
		for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
				transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(vaultA) {
				mintA = tokenBalanceInfo_.Mint
			}
			if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
				transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(vaultB) {
				mintB = tokenBalanceInfo_.Mint
			}
		}

		var inputVault solana.PublicKey
		var outputVault solana.PublicKey
		var inputAddress solana.PublicKey
		var outputAddress solana.PublicKey
		if transfer1Data.Destination.Equals(vaultA) {
			// a is input
			inputVault = vaultA
			outputVault = vaultB
			inputAddress = mintA
			outputAddress = mintB
		} else {
			outputVault = vaultA
			inputVault = vaultB
			inputAddress = mintB
			outputAddress = mintA
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
			OutputAmountWithDecimals: outputAmountWithDecimals,
			UserAddress:              userAddress,
			ParsedKeys: &sol_fi_type.SwapKeys{
				PairAddress: pairAddress,
				VaultA:      vaultA,
				VaultB:      vaultB,
				MintA:       mintA,
				MintB:       mintB,
				Vaults: map[solana.PublicKey]solana.PublicKey{
					mintA: vaultA,
					mintB: vaultB,
				},
			},
			PairAddress:               pairAddress,
			InputVault:                inputVault,
			OutputVault:               outputVault,
			ReserveInputWithDecimals:  reserveInputWithDecimals,
			ReserveOutputWithDecimals: reserveOutputWithDecimals,
			Keys:                      instruction.Accounts,
			MethodId:                  methodId,
			Program:                   sol_fi_constant.SolFiProgram[network],
		})
	}

	feeInfo, err := util.GetFeeInfoFromParsedTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	return &type_.SwapTxDataType{
		TxId:    transaction.Signatures[0].String(),
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}
