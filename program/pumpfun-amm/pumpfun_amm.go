package pumpfun_amm

import (
	"encoding/hex"
	"strconv"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	pumpfun_amm_constant "github.com/pefish/go-coin-sol/program/pumpfun-amm/constant"
	"github.com/pefish/go-coin-sol/program/pumpfun-amm/instruction"
	pumpfun_amm_type "github.com/pefish/go-coin-sol/program/pumpfun-amm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	"github.com/pkg/errors"
)

func GetSwapInstructions(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAmountWithDecimals uint64,
	swapKeys pumpfun_amm_type.SwapKeys,
	isClose bool,
	solReserveWithDecimals uint64,
	tokenReserveWithDecimals uint64,
	slippage uint64, // 0 代表不设置滑点
) ([]solana.Instruction, error) {
	instructions := make([]solana.Instruction, 0)

	userBaseTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		swapKeys.BaseTokenAddress,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s>", userAddress)
	}

	userQuoteTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		swapKeys.QuoteTokenAddress,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, swapKeys.QuoteTokenAddress)
	}

	var userWSOLTokenAccount solana.PublicKey
	var userTokenTokenAccount solana.PublicKey
	if userBaseTokenAccount.Equals(solana.SolMint) {
		userWSOLTokenAccount = userBaseTokenAccount
		userTokenTokenAccount = userQuoteTokenAccount
	} else if userQuoteTokenAccount.Equals(solana.SolMint) {
		userWSOLTokenAccount = userQuoteTokenAccount
		userTokenTokenAccount = userBaseTokenAccount
	} else {
		return nil, errors.New("base or quote both not wsol")
	}

	if swapType == type_.SwapType_Buy {
		if slippage == 0 {
			return nil, errors.New("购买必须设置滑点")
		}
		maxCostSolAmountWithDecimals := uint64(
			float64(slippage+10000) * 1.005 * float64(solReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(tokenReserveWithDecimals) / 10000,
		) // 收取 0.25% 交易手续费

		if maxCostSolAmountWithDecimals == 0 {
			return nil, errors.New("购买数量太小")
		}

		swapInstruction, err := instruction.NewBuyBaseOutInstruction(
			network,
			userAddress,
			tokenAmountWithDecimals,
			maxCostSolAmountWithDecimals,
			swapKeys,
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		instructions = append(
			instructions,
			associated_token_account_instruction.NewCreateIdempotentInstruction(
				userAddress,
				userBaseTokenAccount,
				userAddress,
				swapKeys.BaseTokenAddress,
			),
			associated_token_account_instruction.NewCreateIdempotentInstruction(
				userAddress,
				userQuoteTokenAccount,
				userAddress,
				swapKeys.QuoteTokenAddress,
			),
			system.NewTransferInstruction(
				maxCostSolAmountWithDecimals,
				userAddress,
				userWSOLTokenAccount,
			).Build(),
			token.NewSyncNativeInstruction(userWSOLTokenAccount).Build(),
			swapInstruction,
		)
	} else {
		minReceiveSolAmountWithDecimals := uint64(0)
		if slippage != 0 {
			// 应该收到的 sol 数量
			minReceiveSolAmountWithDecimals = uint64(
				0.995 * float64(10000-slippage) * float64(solReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(tokenReserveWithDecimals) / 10000,
			)
		}

		swapInstruction, err := instruction.NewSellBaseInInstruction(
			network,
			userAddress,
			tokenAmountWithDecimals,
			minReceiveSolAmountWithDecimals,
			swapKeys,
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		instructions = append(
			instructions,
			associated_token_account_instruction.NewCreateIdempotentInstruction(
				userAddress,
				userBaseTokenAccount,
				userAddress,
				swapKeys.BaseTokenAddress,
			),
			swapInstruction,
			token.NewCloseAccountInstruction(
				userWSOLTokenAccount,
				userAddress,
				userAddress,
				nil,
			).Build(),
		)

		if isClose {
			instructions = append(
				instructions,
				token.NewCloseAccountInstruction(
					userTokenTokenAccount,
					userAddress,
					userAddress,
					nil,
				).Build(),
			)
		}
	}

	return instructions, nil
}

func ParseSwapTxByParsedTx(
	network rpc.Cluster,
	meta *rpc.ParsedTransactionMeta,
	parsedTransaction *rpc.ParsedTransaction,
) (*type_.SwapTxDataType, error) {
	txId := parsedTransaction.Signatures[0].String()

	feeInfo, err := util.GetFeeInfoFromParsedTx(meta, parsedTransaction)
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
	for index, instruction := range parsedTransaction.Message.Instructions {
		allInstructions = append(allInstructions, instruction)
		innerInstructions := util.FindInnerInstructionsFromParsedMeta(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		allInstructions = append(allInstructions, innerInstructions...)
	}

	var ammAddress solana.PublicKey
	var baseTokenAddress solana.PublicKey
	var quoteTokenAddress solana.PublicKey
	var poolBaseTokenAccount solana.PublicKey
	var poolQuoteTokenAccount solana.PublicKey
	for _, instruction := range allInstructions {
		if instruction.ProgramId.Equals(pumpfun_amm_constant.Pumpfun_AMM_Program[network]) {
			dataHexString := hex.EncodeToString(instruction.Data)
			methodId := dataHexString[:16]
			if methodId == "66063d1201daebea" || methodId == "33e685a4017f83ad" {
				ammAddress = instruction.Accounts[0]
				baseTokenAddress = instruction.Accounts[3]
				quoteTokenAddress = instruction.Accounts[4]
				poolBaseTokenAccount = instruction.Accounts[7]
				poolQuoteTokenAccount = instruction.Accounts[8]
			}
		}
		if !instruction.ProgramId.Equals(pumpfun_amm_constant.Pumpfun_AMM_Program[network]) {
			continue
		}
		if len(instruction.Accounts) != 1 || !instruction.Accounts[0].Equals(pumpfun_amm_constant.Event_Authority[network]) {
			continue
		}
		dataHexString := hex.EncodeToString(instruction.Data)
		methodId := dataHexString[:16]
		if methodId != "e445a52e51cb9a1d" {
			continue
		}
		eventId := dataHexString[16:32]

		var logObj struct {
			Timestamp                        int64
			BaseAmountOutOrIn                uint64
			MaxOrMinQuoteAmountIn            uint64
			UserBaseTokenReserves            uint64
			UserQuoteTokenReserves           uint64
			PoolBaseTokenReservesBeforeSwap  uint64
			PoolQuoteTokenReservesBeforeSwap uint64
			QuoteAmountOutOrIn               uint64
			LPFeeBasisPoints                 uint64
			LPFee                            uint64
			ProtocolFeeBasisPoints           uint64
			ProtocolFee                      uint64
			QuoteAmountInOrOutWithLpFee      uint64
			UserQuoteAmountInOrOut           uint64
			Pool                             solana.PublicKey
			User                             solana.PublicKey
			UserBaseTokenAccount             solana.PublicKey
			UserQuoteTokenAccount            solana.PublicKey
			ProtocolFeeRecipient             solana.PublicKey
			ProtocolFeeRecipientTokenAccount solana.PublicKey
		}
		err := bin.NewBorshDecoder(instruction.Data[16:]).Decode(&logObj)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		var baseTokenDecimals uint64
		var quoteTokenDecimals uint64
		var reserveBaseWithDecimals uint64
		var reserveQuoteWithDecimals uint64
		for _, tokenBalanceInfo_ := range meta.PostTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(ammAddress) &&
				parsedTransaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(poolBaseTokenAccount) {
				baseTokenDecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
				reserveBaseWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
			}
			if tokenBalanceInfo_.Owner.Equals(ammAddress) &&
				parsedTransaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(poolQuoteTokenAccount) {
				quoteTokenDecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
				reserveQuoteWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
			}
		}

		var inputAddress solana.PublicKey
		var outputAddress solana.PublicKey
		var inputAmountWithDecimals uint64
		var outputAmountWithDecimals uint64
		var inputDecimals uint64
		var outputDecimals uint64
		var inputVault solana.PublicKey
		var outputVault solana.PublicKey
		var reserveInputWithDecimals uint64
		var reserveOutputWithDecimals uint64
		if eventId == "67f4521f2cf57777" {
			// buy. quote -> base
			inputAddress = quoteTokenAddress
			inputAmountWithDecimals = logObj.QuoteAmountInOrOutWithLpFee
			outputAddress = baseTokenAddress
			outputAmountWithDecimals = logObj.BaseAmountOutOrIn
			inputDecimals = quoteTokenDecimals
			outputDecimals = baseTokenDecimals
			inputVault = poolQuoteTokenAccount
			outputVault = poolBaseTokenAccount
			reserveInputWithDecimals = reserveQuoteWithDecimals
			reserveOutputWithDecimals = reserveBaseWithDecimals
		} else if eventId == "3e2f370aa503dc2a" {
			// sell. base -> quote
			inputAddress = baseTokenAddress
			inputAmountWithDecimals = logObj.BaseAmountOutOrIn
			outputAddress = quoteTokenAddress
			outputAmountWithDecimals = logObj.QuoteAmountInOrOutWithLpFee
			inputDecimals = baseTokenDecimals
			outputDecimals = quoteTokenDecimals
			inputVault = poolBaseTokenAccount
			outputVault = poolQuoteTokenAccount
			reserveInputWithDecimals = reserveBaseWithDecimals
			reserveOutputWithDecimals = reserveQuoteWithDecimals
		} else {
			continue
		}

		swaps = append(swaps, &type_.SwapDataType{
			InputAddress:             inputAddress,
			OutputAddress:            outputAddress,
			InputAmountWithDecimals:  inputAmountWithDecimals,
			InputDecimals:            inputDecimals,
			OutputAmountWithDecimals: outputAmountWithDecimals,
			OutputDecimals:           outputDecimals,
			UserAddress:              logObj.User,
			ParsedKeys: &pumpfun_amm_type.SwapKeys{
				AmmAddress:         ammAddress,
				BaseTokenAddress:   baseTokenAddress,
				QuoteTokenAddress:  quoteTokenAddress,
				BaseTokenDecimals:  baseTokenDecimals,
				QuoteTokenDecimals: quoteTokenDecimals,
			},
			ExtraDatas: &pumpfun_amm_type.ExtraDatasType{
				ReserveInputWithDecimals:  reserveInputWithDecimals,
				ReserveOutputWithDecimals: reserveOutputWithDecimals,
			},
			PairAddress: ammAddress,
			InputVault:  inputVault,
			OutputVault: outputVault,
			Keys:        instruction.Accounts,
			AllKeys:     parsedTransaction.Message.AccountKeys,
			MethodId:    methodId,
			Program:     pumpfun_amm_constant.Pumpfun_AMM_Program[network],
		})
	}

	return &type_.SwapTxDataType{
		TxId:    txId,
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}

func ParseAddLiqTxByParsedTx(
	network rpc.Cluster,
	meta *rpc.ParsedTransactionMeta,
	parsedTransaction *rpc.ParsedTransaction,
) (*pumpfun_amm_type.AddLiqTxDataType, error) {
	allInstructions := make([]*rpc.ParsedInstruction, 0)
	for index, instruction := range parsedTransaction.Message.Instructions {
		allInstructions = append(allInstructions, instruction)
		innerInstructions := util.FindInnerInstructionsFromParsedMeta(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		allInstructions = append(allInstructions, innerInstructions...)
	}

	for _, parsedInstruction := range allInstructions {
		if !parsedInstruction.ProgramId.Equals(pumpfun_amm_constant.Pumpfun_AMM_Program[network]) {
			continue
		}
		if len(parsedInstruction.Accounts) != 1 || !parsedInstruction.Accounts[0].Equals(pumpfun_amm_constant.Event_Authority[network]) {
			continue
		}
		dataHexString := hex.EncodeToString(parsedInstruction.Data)
		methodId := dataHexString[:16]
		if methodId != "e445a52e51cb9a1d" {
			continue
		}
		eventId := dataHexString[16:32]
		if eventId != "b1310cd2a076a774" {
			continue
		}

		var logObj struct {
			Timestamp             int64
			Index                 uint16
			Creator               solana.PublicKey
			BaseMint              solana.PublicKey
			QuoteMint             solana.PublicKey
			BaseMintDecimals      uint8
			QuoteMintDecimals     uint8
			BaseAmountIn          uint64
			QuoteAmountIn         uint64
			PoolBaseAmount        uint64
			PoolQuoteAmount       uint64
			MinimumLiquidity      uint64
			InitialLiquidity      uint64
			LPTokenAmountOut      uint64
			PoolBump              uint8
			Pool                  solana.PublicKey
			LPMint                solana.PublicKey
			UserBaseTokenAccount  solana.PublicKey
			UserQuoteTokenAccount solana.PublicKey
		}
		err := bin.NewBorshDecoder(parsedInstruction.Data[16:]).Decode(&logObj)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		feeInfo, err := util.GetFeeInfoFromParsedTx(meta, parsedTransaction)
		if err != nil {
			return nil, err
		}

		return &pumpfun_amm_type.AddLiqTxDataType{
			TxId: parsedTransaction.Signatures[0].String(),
			SwapKeys: pumpfun_amm_type.SwapKeys{
				AmmAddress:         logObj.Pool,
				BaseTokenAddress:   logObj.BaseMint,
				QuoteTokenAddress:  logObj.QuoteMint,
				BaseTokenDecimals:  uint64(logObj.BaseMintDecimals),
				QuoteTokenDecimals: uint64(logObj.QuoteMintDecimals),
			},
			InitBaseAmountWithDecimals:  logObj.PoolBaseAmount,
			InitQuoteAmountWithDecimals: logObj.PoolQuoteAmount,
			FeeInfo:                     feeInfo,
		}, nil

	}

	return nil, nil
}
