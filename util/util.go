package util

import (
	"encoding/hex"
	"fmt"
	"math"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
)

func FindInnerInstructions(meta *rpc.TransactionMeta, index uint64) []solana.CompiledInstruction {
	for _, innerInstruction := range meta.InnerInstructions {
		if innerInstruction.Index == uint16(index) {
			return innerInstruction.Instructions
		}
	}
	return nil
}

func GetComputeUnitPriceFromHelius(
	logger i_logger.ILogger,
	key string,
	accountKeys []string,
) (uint64, error) {
	var httpResult struct {
		Result struct {
			PriorityFeeEstimate float64 `json:"priorityFeeEstimate"`
		} `json:"result"`
	}
	_, _, err := go_http.NewHttpRequester(
		go_http.WithLogger(logger),
		go_http.WithTimeout(10*time.Second),
	).PostForStruct(
		&go_http.RequestParams{
			Url: fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", key),
			Params: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "helius-example",
				"method":  "getPriorityFeeEstimate",
				"params": []map[string]interface{}{
					{
						"accountKeys": accountKeys,
						"options": map[string]interface{}{
							"recommended": true,
						},
					},
				},
			},
		},
		&httpResult,
	)
	if err != nil {
		return 0, err
	}
	return uint64(httpResult.Result.PriorityFeeEstimate), nil
}

func GetFeeInfoFromTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*type_.FeeInfo, error) {
	accountKeys := transaction.Message.AccountKeys
	if meta.LoadedAddresses.Writable != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
	}
	if meta.LoadedAddresses.ReadOnly != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.ReadOnly...)
	}

	priorityFeeWithDecimals := uint64(0)
	computeUnitPrice := 0

	var setComputeUnitLimitInstru solana.CompiledInstruction
	var setComputeUnitPriceInstru solana.CompiledInstruction
	for _, instruction := range transaction.Message.Instructions {
		programPKey := accountKeys[instruction.ProgramIDIndex]
		if !programPKey.Equals(constant.Compute_Budget) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:2]
		if methodId == "02" {
			setComputeUnitLimitInstru = instruction
		}
		if methodId == "03" {
			setComputeUnitPriceInstru = instruction
		}
	}
	computeUnitLimit := 200000
	if setComputeUnitLimitInstru.ProgramIDIndex != 0 {
		var params struct {
			Id    uint8  `json:"id"`
			Units uint32 `json:"units"`
		}
		err := bin.NewBorshDecoder(setComputeUnitLimitInstru.Data).Decode(&params)
		if err != nil {
			return nil, err
		}
		computeUnitLimit = int(params.Units)
	}

	if setComputeUnitPriceInstru.ProgramIDIndex != 0 {
		var params struct {
			Id            uint8  `json:"id"`
			MicroLamports uint64 `json:"microLamports"`
		}
		err := bin.NewBorshDecoder(setComputeUnitPriceInstru.Data).Decode(&params)
		if err != nil {
			return nil, err
		}
		computeUnitPrice = int(params.MicroLamports)

		priorityFeeWithDecimals = uint64((computeUnitPrice * computeUnitLimit) / int(math.Pow(10, 6)))
	}

	return &type_.FeeInfo{
		BaseFeeWithDecimals:     meta.Fee - priorityFeeWithDecimals,
		PriorityFeeWithDecimals: priorityFeeWithDecimals,
		TotalFeeWithDecimals:    meta.Fee,
		ComputeUnitPrice:        uint64(computeUnitPrice),
	}, nil
}
