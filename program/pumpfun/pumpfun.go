package pumpfun

import (
	"context"
	"encoding/hex"
	"time"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	pumpfun_instruction "github.com/pefish/go-coin-sol/program/pumpfun/instruction"
	util "github.com/pefish/go-coin-sol/util"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
)

type SwapType string

const (
	SwapType_Buy  SwapType = "buy"
	SwapType_Sell SwapType = "sell"
)

type SwapDataType struct {
	TokenAddress         solana.PublicKey
	SOLAmount            string
	TokenAmount          string
	Type                 SwapType
	UserAddress          solana.PublicKey
	Timestamp            uint64
	VirtualSolReserves   string
	VirtualTokenReserves string
	UserTokenBalance     string // 交易之后用户的余额
}

type SwapTxDataType struct {
	Swaps   []*SwapDataType
	FeeInfo *util.FeeInfo
}

type ParseTxResult struct {
	SwapTxData      *SwapTxDataType
	CreateTxData    *CreateTxDataType
	RemoveLiqTxData *RemoveLiqTxDataType
}

func ParseTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*ParseTxResult, error) {
	swapData, err := ParseSwapTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	createData, err := ParseCreateTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	removeLiqData, err := ParseRemoveLiqTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	return &ParseTxResult{
		SwapTxData:      swapData,
		CreateTxData:    createData,
		RemoveLiqTxData: removeLiqData,
	}, nil
}

func ParseSwapTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*SwapTxDataType, error) {
	swaps := make([]*SwapDataType, 0)
	accountKeys := transaction.Message.AccountKeys
	if meta.LoadedAddresses.Writable != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
	}
	if meta.LoadedAddresses.ReadOnly != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.ReadOnly...)
	}
	for index, instruction := range transaction.Message.Instructions {
		programPKey := accountKeys[instruction.ProgramIDIndex]
		if !programPKey.Equals(pumpfun_constant.Pumpfun_Program) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:16]
		if methodId != "66063d1201daebea" && methodId != "33e685a4017f83ad" {
			continue
		}

		innerInstructions := util.FindInnerInstructions(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		// 查找 log
		for _, innerInstruction := range innerInstructions {
			programPKey := accountKeys[innerInstruction.ProgramIDIndex]
			if !programPKey.Equals(pumpfun_constant.Pumpfun_Program) {
				continue
			}
			if len(innerInstruction.Accounts) != 1 || !accountKeys[innerInstruction.Accounts[0]].Equals(pumpfun_constant.Pumpfun_Event_Authority) {
				continue
			}
			// 记录事件的指令
			if hex.EncodeToString(innerInstruction.Data)[:16] != "e445a52e51cb9a1d" {
				continue
			}
			var log struct {
				Id                   bin.Uint128      `json:"id"`
				Mint                 solana.PublicKey `json:"mint"`
				SOLAmount            uint64           `json:"solAmount"`
				TokenAmount          uint64           `json:"tokenAmount"`
				IsBuy                bool             `json:"isBuy"`
				User                 solana.PublicKey `json:"user"`
				Timestamp            int64            `json:"timestamp"`
				VirtualSolReserves   uint64           `json:"virtualSolReserves"`
				VirtualTokenReserves uint64           `json:"virtualTokenReserves"`
			}
			err := bin.NewBorshDecoder(innerInstruction.Data).Decode(&log)
			if err != nil {
				// 说明记录的不是 swap 信息
				continue
			}
			// 不报错的，也有可能将其他内容误解读为 swap，所以做校验
			if !log.Mint.Equals(accountKeys[instruction.Accounts[2]]) || !log.User.Equals(accountKeys[instruction.Accounts[6]]) {
				continue
			}
			userTokenBalance := "0"
			for _, postTokenBalanceInfo := range meta.PostTokenBalances {
				if postTokenBalanceInfo.Owner.Equals(log.User) {
					userTokenBalance = postTokenBalanceInfo.UiTokenAmount.UiAmountString
					break
				}
			}
			swaps = append(swaps, &SwapDataType{
				TokenAddress: log.Mint,
				SOLAmount:    go_decimal.Decimal.MustStart(log.SOLAmount).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
				TokenAmount:  go_decimal.Decimal.MustStart(log.TokenAmount).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
				Type: func() SwapType {
					if log.IsBuy {
						return SwapType_Buy
					} else {
						return SwapType_Sell
					}
				}(),
				UserAddress:          log.User,
				UserTokenBalance:     userTokenBalance,
				Timestamp:            uint64(log.Timestamp),
				VirtualSolReserves:   go_decimal.Decimal.MustStart(log.VirtualSolReserves).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
				VirtualTokenReserves: go_decimal.Decimal.MustStart(log.VirtualTokenReserves).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
			})
		}
	}

	feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	return &SwapTxDataType{
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}

type CreateTxDataType struct {
	Name                string
	Symbol              string
	URI                 string
	UserAddress         solana.PublicKey
	BondingCurveAddress solana.PublicKey
	TokenAddress        solana.PublicKey
	FeeInfo             *util.FeeInfo
}

func ParseCreateTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*CreateTxDataType, error) {
	accountKeys := transaction.Message.AccountKeys
	if meta.LoadedAddresses.Writable != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
	}
	if meta.LoadedAddresses.ReadOnly != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.ReadOnly...)
	}
	for _, instruction := range transaction.Message.Instructions {
		programPKey := accountKeys[instruction.ProgramIDIndex]
		if !programPKey.Equals(pumpfun_constant.Pumpfun_Program) {
			continue
		}
		if hex.EncodeToString(instruction.Data)[:16] != "181ec828051c0777" {
			continue
		}
		var params struct {
			Id     uint64 `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
			URI    string `json:"uri"`
		}
		err := bin.NewBorshDecoder(instruction.Data).Decode(&params)
		if err != nil {
			return nil, err
		}
		feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
		if err != nil {
			return nil, err
		}
		return &CreateTxDataType{
			Name:                params.Name,
			Symbol:              params.Symbol,
			URI:                 params.URI,
			UserAddress:         accountKeys[instruction.Accounts[7]],
			BondingCurveAddress: accountKeys[instruction.Accounts[2]],
			TokenAddress:        accountKeys[instruction.Accounts[0]],
			FeeInfo:             feeInfo,
		}, nil

	}

	return nil, nil
}

type RemoveLiqTxDataType struct {
	BondingCurveAddress solana.PublicKey
	TokenAddress        solana.PublicKey
	FeeInfo             *util.FeeInfo
}

// 上岸
func ParseRemoveLiqTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*RemoveLiqTxDataType, error) {
	accountKeys := transaction.Message.AccountKeys
	if meta.LoadedAddresses.Writable != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
	}
	if meta.LoadedAddresses.ReadOnly != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.ReadOnly...)
	}
	if !accountKeys[0].Equals(pumpfun_constant.Pumpfun_Raydium_Migration) {
		return nil, nil
	}
	for _, instruction := range transaction.Message.Instructions {
		programPKey := accountKeys[instruction.ProgramIDIndex]
		if !programPKey.Equals(pumpfun_constant.Pumpfun_Program) {
			continue
		}
		if hex.EncodeToString(instruction.Data)[:16] != "b712469c946da122" {
			continue
		}
		feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
		if err != nil {
			return nil, err
		}
		return &RemoveLiqTxDataType{
			BondingCurveAddress: accountKeys[instruction.Accounts[3]],
			TokenAddress:        accountKeys[instruction.Accounts[2]],
			FeeInfo:             feeInfo,
		}, nil

	}

	return nil, nil
}

type TokenMetadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Image       string `json:"image"`
	ShowName    bool   `json:"showName"`
	CreatedOn   string `json:"createdOn"`
	Twitter     string `json:"twitter"`
	Telegram    string `json:"telegram"`
	Website     string `json:"website"`
}

func URIInfo(logger i_logger.ILogger, uri string) (*TokenMetadata, error) {
	var httpResult TokenMetadata
	_, _, err := go_http.NewHttpRequester(
		go_http.WithTimeout(5*time.Second),
		go_http.WithLogger(logger),
	).GetForStruct(&go_http.RequestParams{
		Url: uri,
	}, &httpResult)
	if err != nil {
		return nil, err
	}
	return &httpResult, nil
}

type BondingCurveDataType struct {
	VirtualTokenReserves string
	VirtualSolReserves   string
	RealTokenReserves    string
	RealSolReserves      string
	TokenTotalSupply     string
	Complete             bool
}

func GetBondingCurveData(
	rpcClient *rpc.Client,
	tokenAddress solana.PublicKey,
) (*BondingCurveDataType, error) {
	bondingCurveAddress, _, err := solana.FindProgramAddress([][]byte{
		[]byte("bonding-curve"),
		tokenAddress.Bytes(),
	}, pumpfun_constant.Pumpfun_Program)
	if err != nil {
		return nil, err
	}
	var data struct {
		Id                   uint64
		VirtualTokenReserves uint64
		VirtualSolReserves   uint64
		RealTokenReserves    uint64
		RealSolReserves      uint64
		TokenTotalSupply     uint64
		Complete             bool
	}
	err = rpcClient.GetAccountDataBorshInto(context.Background(), bondingCurveAddress, &data)
	if err != nil {
		return nil, err
	}
	return &BondingCurveDataType{
		VirtualTokenReserves: go_decimal.Decimal.MustStart(data.VirtualTokenReserves).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
		VirtualSolReserves:   go_decimal.Decimal.MustStart(data.VirtualSolReserves).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
		RealTokenReserves:    go_decimal.Decimal.MustStart(data.RealTokenReserves).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
		RealSolReserves:      go_decimal.Decimal.MustStart(data.RealSolReserves).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
		TokenTotalSupply:     go_decimal.Decimal.MustStart(data.TokenTotalSupply).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
		Complete:             data.Complete,
	}, nil
}

func GetSwapInstructions(
	userAddress solana.PublicKey,
	swapType SwapType,
	tokenAddress solana.PublicKey,
	tokenAmount string,
	isCloseUserAssociatedTokenAddress bool,
	virtualSolReserve string,
	virtualTokenReserve string,
	slippage uint64,
) ([]solana.Instruction, error) {
	if slippage == 0 {
		slippage = 50 // 0.5%
	}
	instructions := make([]solana.Instruction, 0)

	userAssociatedTokenAddress, _, err := solana.FindAssociatedTokenAddress(userAddress, tokenAddress)
	if err != nil {
		return nil, err
	}
	if swapType == SwapType_Buy {
		instructions = append(instructions, associated_token_account_instruction.NewCreateIdempotentInstruction(
			userAddress,
			userAssociatedTokenAddress,
			userAddress,
			tokenAddress,
		))
	}

	bondingCurveAddress, _, err := solana.FindProgramAddress([][]byte{
		[]byte("bonding-curve"),
		tokenAddress.Bytes(),
	}, pumpfun_constant.Pumpfun_Program)
	if err != nil {
		return nil, err
	}
	var swapInstruction solana.Instruction
	if swapType == SwapType_Buy {
		// 应该花费的 sol 数量
		shouldCostSolAmount := go_decimal.Decimal.MustStart(virtualSolReserve).MustMulti(tokenAmount).MustDiv(virtualTokenReserve).MustMultiForString(1.01) // pumpfun 收取 1% 手续费
		// 最大多花 sol 的数量
		maxMoreSolAmount := go_decimal.Decimal.MustStart(shouldCostSolAmount).MustMulti(slippage).MustDivForString(10000)
		maxCostSolAmount := go_decimal.Decimal.MustStart(shouldCostSolAmount).MustAdd(maxMoreSolAmount).RoundDownForString(constant.SOL_Decimals)
		instruction, err := pumpfun_instruction.NewBuyInstruction(
			userAddress,
			tokenAddress,
			bondingCurveAddress,
			userAssociatedTokenAddress,
			tokenAmount,
			maxCostSolAmount,
		)
		if err != nil {
			return nil, err
		}
		swapInstruction = instruction
	} else {
		// 应该收到的 sol 数量
		shouldReceiveSolAmount := go_decimal.Decimal.MustStart(virtualSolReserve).MustMulti(tokenAmount).MustDiv(virtualTokenReserve).MustMultiForString(0.99)
		// 最大少收到 sol 的数量
		maxLessSolAmount := go_decimal.Decimal.MustStart(shouldReceiveSolAmount).MustMulti(slippage).MustDivForString(10000)
		minReceiveSolAmount := go_decimal.Decimal.MustStart(shouldReceiveSolAmount).MustSub(maxLessSolAmount).RoundDownForString(constant.SOL_Decimals)
		instruction, err := pumpfun_instruction.NewSellInstruction(
			userAddress,
			tokenAddress,
			bondingCurveAddress,
			userAssociatedTokenAddress,
			tokenAmount,
			minReceiveSolAmount,
		)
		if err != nil {
			return nil, err
		}
		swapInstruction = instruction
	}
	instructions = append(instructions, swapInstruction)

	if swapType == SwapType_Sell && isCloseUserAssociatedTokenAddress {
		instructions = append(
			instructions,
			token.NewCloseAccountInstruction(
				userAssociatedTokenAddress,
				userAddress,
				userAddress,
				nil,
			).Build(),
		)
	}

	return instructions, nil
}
