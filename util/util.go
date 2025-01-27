package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/pefish/go-coin-sol/type"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

func FindInnerInstructions(meta *rpc.TransactionMeta, index uint64) []solana.CompiledInstruction {
	for _, innerInstruction := range meta.InnerInstructions {
		if innerInstruction.Index == uint16(index) {
			return innerInstruction.Instructions
		}
	}
	return nil
}

func FindInnerInstructionsFromParsedMeta(meta *rpc.ParsedTransactionMeta, index uint64) []*rpc.ParsedInstruction {
	for _, innerInstruction := range meta.InnerInstructions {
		if innerInstruction.Index == index {
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
		return 0, errors.Wrap(err, "")
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
		if !programPKey.Equals(solana.ComputeBudget) {
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
			return nil, errors.Wrap(err, "")
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
			return nil, errors.Wrap(err, "")
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

func GetFeeInfoFromParsedTx(meta *rpc.ParsedTransactionMeta, parsedTransaction *rpc.ParsedTransaction) (*type_.FeeInfo, error) {
	priorityFeeWithDecimals := uint64(0)
	computeUnitPrice := 0

	var setComputeUnitLimitInstru *rpc.ParsedInstruction
	var setComputeUnitPriceInstru *rpc.ParsedInstruction
	for _, parsedInstruction := range parsedTransaction.Message.Instructions {
		if !parsedInstruction.ProgramId.Equals(solana.ComputeBudget) {
			continue
		}
		methodId := hex.EncodeToString(parsedInstruction.Data)[:2]
		if methodId == "02" {
			setComputeUnitLimitInstru = parsedInstruction
		}
		if methodId == "03" {
			setComputeUnitPriceInstru = parsedInstruction
		}
	}
	computeUnitLimit := 200000
	if setComputeUnitLimitInstru != nil {
		var params struct {
			Id    uint8  `json:"id"`
			Units uint32 `json:"units"`
		}
		err := bin.NewBorshDecoder(setComputeUnitLimitInstru.Data).Decode(&params)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		computeUnitLimit = int(params.Units)
	}

	if setComputeUnitPriceInstru != nil {
		var params struct {
			Id            uint8  `json:"id"`
			MicroLamports uint64 `json:"microLamports"`
		}
		err := bin.NewBorshDecoder(setComputeUnitPriceInstru.Data).Decode(&params)
		if err != nil {
			return nil, errors.Wrap(err, "")
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

type TransferInstructionDataType struct {
	Source             solana.PublicKey
	Destination        solana.PublicKey
	AmountWithDecimals uint64
	Authority          solana.PublicKey
}

func DecodeTransferParsedInstruction(transferInstruction *rpc.ParsedInstruction) (*TransferInstructionDataType, error) {
	if transferInstruction.Parsed == nil {
		return nil, errors.New("Parsed 内容不存在，可能不是 transfer 指令")
	}
	d, _ := transferInstruction.Parsed.MarshalJSON()
	var transferData struct {
		Info struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Amount      string `json:"amount"`
			Authority   string `json:"authority"`
		} `json:"info"`
		Type string `json:"type"`
	}
	err := json.Unmarshal(d, &transferData)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if transferData.Type != "transfer" {
		return nil, errors.Errorf("不是 transfer 指令, %#v", transferData)
	}
	amountWithDecimals, _ := strconv.ParseUint(transferData.Info.Amount, 10, 64)

	return &TransferInstructionDataType{
		Source:             solana.MustPublicKeyFromBase58(transferData.Info.Source),
		Destination:        solana.MustPublicKeyFromBase58(transferData.Info.Destination),
		AmountWithDecimals: amountWithDecimals,
		Authority:          solana.MustPublicKeyFromBase58(transferData.Info.Authority),
	}, nil
}
