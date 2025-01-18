package pumpfun

import (
	"context"
	"encoding/hex"
	"time"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	pumpfun_instruction "github.com/pefish/go-coin-sol/program/pumpfun/instruction"
	pumpfun_type "github.com/pefish/go-coin-sol/program/pumpfun/type"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	util "github.com/pefish/go-coin-sol/util"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

func ParseTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*pumpfun_type.ParseTxResult, error) {
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

	return &pumpfun_type.ParseTxResult{
		SwapTxData:      swapData,
		CreateTxData:    createData,
		RemoveLiqTxData: removeLiqData,
		AddLiqTxData:    addLiqData,
	}, nil
}

func ParseSwapTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*pumpfun_type.SwapTxDataType, error) {
	swaps := make([]*pumpfun_type.SwapDataType, 0)
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
		dataHexString := hex.EncodeToString(instruction.Data)
		if dataHexString[:16] != "e445a52e51cb9a1d" {
			continue
		}
		if len(dataHexString) > 350 {
			continue
		}
		var log struct {
			Id                   uint64           `json:"id"`
			Mint                 solana.PublicKey `json:"mint"`
			SOLAmount            uint64           `json:"solAmount"`
			TokenAmount          uint64           `json:"tokenAmount"`
			IsBuy                bool             `json:"isBuy"`
			User                 solana.PublicKey `json:"user"`
			Timestamp            int64            `json:"timestamp"`
			VirtualSolReserves   uint64           `json:"virtualSolReserves"`
			VirtualTokenReserves uint64           `json:"virtualTokenReserves"`
		}
		err := bin.NewBorshDecoder(instruction.Data[8:]).Decode(&log)
		if err != nil {
			// 说明记录的不是 swap 信息
			continue
		}
		if log.VirtualSolReserves == 0 ||
			log.VirtualTokenReserves == 0 ||
			log.Timestamp == 0 {
			continue
		}
		swaps = append(swaps, &pumpfun_type.SwapDataType{
			TokenAddress:            log.Mint,
			SOLAmountWithDecimals:   log.SOLAmount,
			TokenAmountWithDecimals: log.TokenAmount,
			Type: func() type_.SwapType {
				if log.IsBuy {
					return type_.SwapType_Buy
				} else {
					return type_.SwapType_Sell
				}
			}(),
			UserAddress:                    log.User,
			Timestamp:                      uint64(log.Timestamp * 1000),
			ReserveSOLAmountWithDecimals:   log.VirtualSolReserves,
			ReserveTokenAmountWithDecimals: log.VirtualTokenReserves,
		})
	}

	feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
	if err != nil {
		return nil, err
	}
	return &pumpfun_type.SwapTxDataType{
		TxId:                          transaction.Signatures[0].String(),
		Swaps:                         swaps,
		FeeInfo:                       feeInfo,
		UserBalanceWithDecimals:       meta.PostBalances[0],
		BeforeUserBalanceWithDecimals: meta.PreBalances[0],
	}, nil
}

func ParseCreateTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*pumpfun_type.CreateTxDataType, error) {
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
			return nil, errors.Wrap(err, "")
		}
		feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
		if err != nil {
			return nil, err
		}
		return &pumpfun_type.CreateTxDataType{
			TxId: transaction.Signatures[0].String(),
			CreateDataType: pumpfun_type.CreateDataType{
				Name:                params.Name,
				Symbol:              params.Symbol,
				URI:                 params.URI,
				UserAddress:         accountKeys[instruction.Accounts[7]],
				BondingCurveAddress: accountKeys[instruction.Accounts[2]],
				TokenAddress:        accountKeys[instruction.Accounts[0]],
			},
			FeeInfo: feeInfo,
		}, nil

	}

	return nil, nil
}

// 上岸
func ParseRemoveLiqTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*pumpfun_type.RemoveLiqTxDataType, error) {
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
		return &pumpfun_type.RemoveLiqTxDataType{
			TxId:                transaction.Signatures[0].String(),
			BondingCurveAddress: accountKeys[instruction.Accounts[3]],
			TokenAddress:        accountKeys[instruction.Accounts[2]],
			FeeInfo:             feeInfo,
		}, nil

	}

	return nil, nil
}

func ParseAddLiqTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*pumpfun_type.AddLiqTxDataType, error) {
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
			return nil, errors.Wrap(err, "")
		}

		feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
		if err != nil {
			return nil, err
		}
		return &pumpfun_type.AddLiqTxDataType{
			TxId:                        transaction.Signatures[0].String(),
			TokenAddress:                accountKeys[instruction.Accounts[9]],
			AMMAddress:                  accountKeys[instruction.Accounts[4]],
			PoolCoinTokenAccount:        accountKeys[instruction.Accounts[10]],
			PoolPcTokenAccount:          accountKeys[instruction.Accounts[11]],
			InitSOLAmountWithDecimals:   params.InitCoinAmount,
			InitTokenAmountWithDecimals: params.InitPcAmount,
			FeeInfo:                     feeInfo,
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
		return nil, errors.Wrap(err, "")
	}
	return &httpResult, nil
}

type BondingCurveDataType struct {
	BondingCurveAddress             string
	VirtualTokenReserveWithDecimals uint64
	VirtualSolReserveWithDecimals   uint64
	RealTokenReserveWithDecimals    uint64
	RealSolReserveWithDecimals      uint64
	TokenTotalSupplyWithDecimals    uint64
	Complete                        bool
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
			return nil, errors.Wrapf(err, "<tokenAddress: %s>", tokenAddress.String())
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
		return nil, errors.Wrapf(err, "<bondingCurveAddress: %s>", bondingCurveAddress.String())
	}
	return &BondingCurveDataType{
		BondingCurveAddress:             bondingCurveAddress.String(),
		VirtualTokenReserveWithDecimals: data.VirtualTokenReserves,
		VirtualSolReserveWithDecimals:   data.VirtualSolReserves,
		RealTokenReserveWithDecimals:    data.RealTokenReserves,
		RealSolReserveWithDecimals:      data.RealSolReserves,
		TokenTotalSupplyWithDecimals:    data.TokenTotalSupply,
		Complete:                        data.Complete,
	}, nil
}

func GetSwapInstructions(
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	isCloseUserAssociatedTokenAddress bool,
	virtualSolReserveWithDecimals uint64,
	virtualTokenReserveWithDecimals uint64,
	slippage int64,
) ([]solana.Instruction, error) {
	if slippage == 0 {
		slippage = 50 // 0.5%
	}
	instructions := make([]solana.Instruction, 0)

	userAssociatedTokenAddress, _, err := solana.FindAssociatedTokenAddress(userAddress, tokenAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, tokenAddress)
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
		return nil, errors.Wrapf(err, "<tokenAddress: %s>", tokenAddress)
	}
	var swapInstruction solana.Instruction
	if swapType == type_.SwapType_Buy {
		if slippage == -1 {
			return nil, errors.New("购买必须设置滑点")
		}
		maxCostSolAmountWithDecimals := uint64(
			float64(slippage+10000) * 1.01 * float64(virtualSolReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(virtualTokenReserveWithDecimals) / 10000,
		) // pumpfun 收取 1% 手续费
		instruction, err := pumpfun_instruction.NewBuyBaseOutInstruction(
			userAddress,
			tokenAddress,
			bondingCurveAddress,
			userAssociatedTokenAddress,
			tokenAmountWithDecimals,
			maxCostSolAmountWithDecimals,
		)
		if err != nil {
			return nil, err
		}
		swapInstruction = instruction
	} else {
		minReceiveSolAmountWithDecimals := uint64(0)
		if slippage != -1 {
			// 应该收到的 sol 数量
			minReceiveSolAmountWithDecimals = uint64(
				0.99 * float64(10000-slippage) * float64(virtualSolReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(virtualTokenReserveWithDecimals) / 10000,
			)
		}
		instruction, err := pumpfun_instruction.NewSellBaseInInstruction(
			userAddress,
			tokenAddress,
			bondingCurveAddress,
			userAssociatedTokenAddress,
			tokenAmountWithDecimals,
			minReceiveSolAmountWithDecimals,
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

func DeriveBondingCurveAddress(tokenAddress solana.PublicKey) (solana.PublicKey, error) {
	bondingCurveAddress, _, err := solana.FindProgramAddress([][]byte{
		[]byte("bonding-curve"),
		tokenAddress.Bytes(),
	}, pumpfun_constant.Pumpfun_Program)
	if err != nil {
		return solana.PublicKey{}, errors.Wrapf(err, "<tokenAddress: %s>", tokenAddress)
	}

	return bondingCurveAddress, nil
}
