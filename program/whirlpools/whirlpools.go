package whirlpools

import (
	"encoding/hex"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/itherunder/go-coin-sol/discriminator"
	whirlpools_constant "github.com/itherunder/go-coin-sol/program/whirlpools/constant"
	whirlpools_type "github.com/itherunder/go-coin-sol/program/whirlpools/type"
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
		if !instruction.ProgramId.Equals(whirlpools_constant.WhirlpoolsProgram[network]) {
			continue
		}
		var userAddress solana.PublicKey
		var pairAddress solana.PublicKey
		var parsedKeys interface{}
		var inputAmountWithDecimals uint64
		var outputAmountWithDecimals uint64
		var inputDecimals uint64
		var outputDecimals uint64
		var inputAddress solana.PublicKey
		var outputAddress solana.PublicKey
		var inputVault solana.PublicKey
		var outputVault solana.PublicKey
		var mintADecimals uint64
		var mintBDecimals uint64

		methodId := hex.EncodeToString(instruction.Data)[:16]
		if methodId == hex.EncodeToString(discriminator.GetDiscriminator("global", "swap_v2")) {
			userAddress = instruction.Accounts[3]
			pairAddress = instruction.Accounts[4]

			parsedKeys = &whirlpools_type.SwapV2Keys{
				PairAddress: pairAddress,
				Oracle:      instruction.Accounts[14],
				MintA:       instruction.Accounts[5],
				MintB:       instruction.Accounts[6],
				TickArray0:  instruction.Accounts[11],
				TickArray1:  instruction.Accounts[12],
				TickArray2:  instruction.Accounts[13],
				Vaults: map[solana.PublicKey]solana.PublicKey{
					instruction.Accounts[5]: instruction.Accounts[8],
					instruction.Accounts[6]: instruction.Accounts[10],
				},
			}
			transferDatas, err := util.FindNextTwoTransferCheckedDatas(index+1, allInstructions)
			if err != nil {
				return nil, errors.Wrapf(err, "<txid: %s>", txId)
			}
			inputAmountWithDecimals = transferDatas[0].AmountWithDecimals
			outputAmountWithDecimals = transferDatas[1].AmountWithDecimals
			inputDecimals = transferDatas[0].Decimals
			outputDecimals = transferDatas[1].Decimals
			if transferDatas[0].Destination.Equals(instruction.Accounts[8]) {
				// a is input
				inputAddress = instruction.Accounts[5]
				outputAddress = instruction.Accounts[6]
				inputVault = instruction.Accounts[8]
				outputVault = instruction.Accounts[10]
			} else {
				inputAddress = instruction.Accounts[6]
				outputAddress = instruction.Accounts[5]
				inputVault = instruction.Accounts[10]
				outputVault = instruction.Accounts[8]
			}
		} else if methodId == hex.EncodeToString(discriminator.GetDiscriminator("global", "swap")) {
			userAddress = instruction.Accounts[1]
			pairAddress = instruction.Accounts[2]

			vaultA := instruction.Accounts[4]
			vaultB := instruction.Accounts[6]
			var aMint solana.PublicKey
			var bMint solana.PublicKey
			for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
				if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
					transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(vaultA) {
					aMint = tokenBalanceInfo_.Mint
					mintADecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
				}
				if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
					transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(vaultB) {
					bMint = tokenBalanceInfo_.Mint
					mintBDecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
				}
			}

			parsedKeys = &whirlpools_type.SwapKeys{
				PairAddress: pairAddress,
				Oracle:      instruction.Accounts[8],
				TickArray0:  instruction.Accounts[7],
				TickArray1:  instruction.Accounts[8],
				TickArray2:  instruction.Accounts[9],

				MintA: aMint,
				MintB: bMint,
				Vaults: map[solana.PublicKey]solana.PublicKey{
					aMint: vaultA,
					bMint: vaultB,
				},
			}
			transferDatas, err := util.FindNextTwoTransferDatas(index+1, allInstructions)
			if err != nil {
				return nil, errors.Wrapf(err, "<txid: %s>", txId)
			}

			inputAmountWithDecimals = transferDatas[0].AmountWithDecimals
			outputAmountWithDecimals = transferDatas[1].AmountWithDecimals
			if transferDatas[0].Destination.Equals(vaultA) {
				// a is input
				inputAddress = aMint
				inputDecimals = mintADecimals
				outputAddress = bMint
				outputDecimals = mintBDecimals
				inputVault = vaultA
				outputVault = vaultB
			} else {
				inputAddress = bMint
				inputDecimals = mintBDecimals
				outputAddress = aMint
				outputDecimals = mintADecimals
				inputVault = vaultB
				outputVault = vaultA
			}
		} else {
			continue
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
			PairAddress:              pairAddress,
			InputVault:               inputVault,
			OutputVault:              outputVault,
			ParsedKeys:               parsedKeys,
			ExtraDatas: &whirlpools_type.ExtraDatasType{
				ReserveInputWithDecimals:  reserveInputWithDecimals,
				ReserveOutputWithDecimals: reserveOutputWithDecimals,
			},
			Keys:     instruction.Accounts,
			AllKeys:  transaction.Message.AccountKeys,
			MethodId: methodId,
			Program:  whirlpools_constant.WhirlpoolsProgram[network],
		})
	}

	return &type_.SwapTxDataType{
		TxId:    txId,
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}
