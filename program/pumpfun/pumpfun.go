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
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	util "github.com/pefish/go-coin-sol/util"
	go_decimal "github.com/pefish/go-decimal"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type SwapDataType struct {
	TokenAddress       solana.PublicKey `json:"token_address"`
	SOLAmount          string           `json:"sol_amount"`
	TokenAmount        string           `json:"token_amount"`
	Type               type_.SwapType   `json:"type"`
	UserAddress        solana.PublicKey `json:"user_address"`
	ReserveSOLAmount   string           `json:"reserve_sol_amount"`
	ReserveTokenAmount string           `json:"reserve_token_amount"`
	UserTokenBalance   string           `json:"user_token_balance"` // 交易之后用户的余额
	UserBalance        string           `json:"user_balance"`
	BeforeUserBalance  string           `json:"before_user_balance"`
}

type SwapTxDataType struct {
	Swaps   []*SwapDataType
	FeeInfo *type_.FeeInfo
	TxId    string
}

type ParseTxResult struct {
	SwapTxData      *SwapTxDataType
	CreateTxData    *CreateTxDataType
	RemoveLiqTxData *RemoveLiqTxDataType
	AddLiqTxData    *AddLiqTxDataType
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

	addLiqData, err := ParseAddLiqTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	return &ParseTxResult{
		SwapTxData:      swapData,
		CreateTxData:    createData,
		RemoveLiqTxData: removeLiqData,
		AddLiqTxData:    addLiqData,
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

	allInstructions := make([]solana.CompiledInstruction, 0)
	for index, instruction := range transaction.Message.Instructions {
		allInstructions = append(allInstructions, instruction)
		innerInstructions := util.FindInnerInstructions(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		allInstructions = append(allInstructions, innerInstructions...)
	}

	for _, instruction := range allInstructions {
		programPKey := accountKeys[instruction.ProgramIDIndex]
		if !programPKey.Equals(pumpfun_constant.Pumpfun_Program) {
			continue
		}
		if len(instruction.Accounts) != 1 || !accountKeys[instruction.Accounts[0]].Equals(pumpfun_constant.Pumpfun_Event_Authority) {
			continue
		}
		// 记录事件的指令
		if hex.EncodeToString(instruction.Data)[:16] != "e445a52e51cb9a1d" {
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
		err := bin.NewBorshDecoder(instruction.Data).Decode(&log)
		if err != nil {
			// 说明记录的不是 swap 信息
			continue
		}
		if log.VirtualSolReserves == 0 ||
			log.VirtualTokenReserves == 0 ||
			log.Timestamp == 0 {
			continue
		}
		userTokenBalance := "0"
		for _, postTokenBalanceInfo := range meta.PostTokenBalances {
			if postTokenBalanceInfo.Owner.Equals(log.User) &&
				postTokenBalanceInfo.Mint.Equals(log.Mint) {
				userTokenBalance = postTokenBalanceInfo.UiTokenAmount.UiAmountString
				break
			}
		}
		tokenAmount := go_decimal.Decimal.MustStart(log.TokenAmount).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString()
		swaps = append(swaps, &SwapDataType{
			TokenAddress: log.Mint,
			SOLAmount:    go_decimal.Decimal.MustStart(log.SOLAmount).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
			TokenAmount:  tokenAmount,
			Type: func() type_.SwapType {
				if log.IsBuy {
					return type_.SwapType_Buy
				} else {
					return type_.SwapType_Sell
				}
			}(),
			UserAddress:        log.User,
			UserTokenBalance:   userTokenBalance,
			UserBalance:        go_decimal.Decimal.MustStart(meta.PostBalances[0]).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
			BeforeUserBalance:  go_decimal.Decimal.MustStart(meta.PreBalances[0]).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
			ReserveSOLAmount:   go_decimal.Decimal.MustStart(log.VirtualSolReserves).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
			ReserveTokenAmount: go_decimal.Decimal.MustStart(log.VirtualTokenReserves).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
		})
	}

	feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	return &SwapTxDataType{
		TxId:    transaction.Signatures[0].String(),
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}

type CreateTxDataType struct {
	TxId                string           `json:"txid"`
	Name                string           `json:"name"`
	Symbol              string           `json:"symbol"`
	URI                 string           `json:"uri"`
	UserAddress         solana.PublicKey `json:"user_address"`
	BondingCurveAddress solana.PublicKey `json:"bonding_curve_address"`
	TokenAddress        solana.PublicKey `json:"token_address"`
	FeeInfo             *type_.FeeInfo   `json:"fee_info"`
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
			TxId:                transaction.Signatures[0].String(),
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
	TxId                string           `json:"txid"`
	BondingCurveAddress solana.PublicKey `json:"bonding_curve_address"`
	TokenAddress        solana.PublicKey `json:"token_address"`
	FeeInfo             *type_.FeeInfo   `json:"fee_info"`
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
			TxId:                transaction.Signatures[0].String(),
			BondingCurveAddress: accountKeys[instruction.Accounts[3]],
			TokenAddress:        accountKeys[instruction.Accounts[2]],
			FeeInfo:             feeInfo,
		}, nil

	}

	return nil, nil
}

type AddLiqTxDataType struct {
	TxId                 string           `json:"txid"`
	TokenAddress         solana.PublicKey `json:"token_address"`
	InitSOLAmount        string           `json:"init_sol_amount"`
	InitTokenAmount      string           `json:"init_token_amount"`
	AMMAddress           solana.PublicKey `json:"amm_address"`
	PoolCoinTokenAccount solana.PublicKey `json:"pool_coin_token_account"`
	PoolPcTokenAccount   solana.PublicKey `json:"pool_pc_token_account"`

	FeeInfo *type_.FeeInfo `json:"fee_info"`
}

func ParseAddLiqTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*AddLiqTxDataType, error) {
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
		if !programPKey.Equals(raydium_constant.Raydium_Liquidity_Pool_V4) {
			continue
		}
		if hex.EncodeToString(instruction.Data)[:2] != "01" {
			continue
		}
		var params struct {
			Discriminator  uint8  `json:"discriminator"`
			Nonce          uint8  `json:"nonce"`
			OpenTime       uint64 `json:"openTime"`
			InitPcAmount   uint64 `json:"initPcAmount"`
			InitCoinAmount uint64 `json:"initCoinAmount"`
		}
		err := bin.NewBorshDecoder(instruction.Data).Decode(&params)
		if err != nil {
			return nil, err
		}

		feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
		if err != nil {
			return nil, err
		}
		return &AddLiqTxDataType{
			TxId:                 transaction.Signatures[0].String(),
			TokenAddress:         accountKeys[instruction.Accounts[9]],
			AMMAddress:           accountKeys[instruction.Accounts[4]],
			PoolCoinTokenAccount: accountKeys[instruction.Accounts[10]],
			PoolPcTokenAccount:   accountKeys[instruction.Accounts[11]],
			InitSOLAmount:        go_decimal.Decimal.MustStart(params.InitCoinAmount).MustUnShiftedBy(constant.SOL_Decimals).EndForString(),
			InitTokenAmount:      go_decimal.Decimal.MustStart(params.InitPcAmount).MustUnShiftedBy(pumpfun_constant.Pumpfun_Token_Decimals).EndForString(),
			FeeInfo:              feeInfo,
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
	BondingCurveAddress  string
	VirtualTokenReserves string
	VirtualSolReserves   string
	RealTokenReserves    string
	RealSolReserves      string
	TokenTotalSupply     string
	Complete             bool
}

func GetBondingCurveData(
	rpcClient *rpc.Client,
	tokenAddress *solana.PublicKey,
	bondingCurveAddress *solana.PublicKey,
) (*BondingCurveDataType, error) {
	if tokenAddress == nil && bondingCurveAddress == nil {
		return nil, errors.New("Token address or bondingCurve address can not both be nil.")
	}
	if bondingCurveAddress == nil {
		bondingCurveAddress_, _, err := solana.FindProgramAddress([][]byte{
			[]byte("bonding-curve"),
			tokenAddress.Bytes(),
		}, pumpfun_constant.Pumpfun_Program)
		if err != nil {
			return nil, err
		}
		bondingCurveAddress = &bondingCurveAddress_
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
	err := rpcClient.GetAccountDataBorshInto(context.Background(), *bondingCurveAddress, &data)
	if err != nil {
		return nil, err
	}
	return &BondingCurveDataType{
		BondingCurveAddress:  bondingCurveAddress.String(),
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
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmount string,
	isCloseUserAssociatedTokenAddress bool,
	virtualSolReserve string,
	virtualTokenReserve string,
	slippage int64,
) ([]solana.Instruction, error) {
	if slippage == 0 {
		slippage = 50 // 0.5%
	}
	instructions := make([]solana.Instruction, 0)

	userAssociatedTokenAddress, _, err := solana.FindAssociatedTokenAddress(userAddress, tokenAddress)
	if err != nil {
		return nil, err
	}
	if swapType == type_.SwapType_Buy {
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
	if swapType == type_.SwapType_Buy {
		if slippage == -1 {
			return nil, errors.New("购买必须设置滑点")
		}
		// 应该花费的 sol 数量
		shouldCostSolAmount := go_decimal.Decimal.MustStart(virtualSolReserve).MustMulti(tokenAmount).MustDiv(virtualTokenReserve).MustMultiForString(1.01) // pumpfun 收取 1% 手续费
		// 最大多花 sol 的数量
		maxMoreSolAmount := go_decimal.Decimal.MustStart(shouldCostSolAmount).MustMulti(slippage).MustDivForString(10000)
		maxCostSolAmount := go_decimal.Decimal.MustStart(shouldCostSolAmount).MustAdd(maxMoreSolAmount).RoundDownForString(constant.SOL_Decimals)
		instruction, err := pumpfun_instruction.NewBuyBaseOutInstruction(
			userAddress,
			tokenAddress,
			bondingCurveAddress,
			userAssociatedTokenAddress,
			type_.TokenAmountInfo{
				Amount:   tokenAmount,
				Decimals: pumpfun_constant.Pumpfun_Token_Decimals,
			},
			maxCostSolAmount,
		)
		if err != nil {
			return nil, err
		}
		swapInstruction = instruction
	} else {
		minReceiveSolAmount := "0"
		if slippage != -1 {
			// 应该收到的 sol 数量
			shouldReceiveSolAmount := go_decimal.Decimal.MustStart(virtualSolReserve).MustMulti(tokenAmount).MustDiv(virtualTokenReserve).MustMultiForString(0.99)
			// 最大少收到 sol 的数量
			maxLessSolAmount := go_decimal.Decimal.MustStart(shouldReceiveSolAmount).MustMulti(slippage).MustDivForString(10000)
			minReceiveSolAmount = go_decimal.Decimal.MustStart(shouldReceiveSolAmount).MustSub(maxLessSolAmount).RoundDownForString(constant.SOL_Decimals)
		}
		instruction, err := pumpfun_instruction.NewSellBaseInInstruction(
			userAddress,
			tokenAddress,
			bondingCurveAddress,
			userAssociatedTokenAddress,
			type_.TokenAmountInfo{
				Amount:   tokenAmount,
				Decimals: pumpfun_constant.Pumpfun_Token_Decimals,
			},
			minReceiveSolAmount,
		)
		if err != nil {
			return nil, err
		}
		swapInstruction = instruction
	}
	instructions = append(instructions, swapInstruction)

	if swapType == type_.SwapType_Sell && isCloseUserAssociatedTokenAddress {
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
