package util

import (
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	go_decimal "github.com/pefish/go-decimal"
)

func FindInnerInstructions(meta *rpc.TransactionMeta, index uint64) []solana.CompiledInstruction {
	for _, innerInstruction := range meta.InnerInstructions {
		if innerInstruction.Index == uint16(index) {
			return innerInstruction.Instructions
		}
	}
	return nil
}

type FeeInfo struct {
	BaseFee          string
	PriorityFee      string
	TotalFee         string
	ComputeUnitPrice uint64
}

func GetFeeInfoFromTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*FeeInfo, error) {
	accountKeys := transaction.Message.AccountKeys
	if meta.LoadedAddresses.Writable != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
	}
	if meta.LoadedAddresses.ReadOnly != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.ReadOnly...)
	}

	baseFee := go_decimal.Decimal.MustStart(meta.Fee).MustUnShiftedBy(constant.SOL_Decimals).EndForString()
	priorityFee := "0"
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

		priorityFee = go_decimal.Decimal.MustStart(computeUnitPrice).MustMulti(computeUnitLimit).MustUnShiftedBy(constant.SOL_Decimals + 6).EndForString()
	}

	return &FeeInfo{
		BaseFee:          baseFee,
		PriorityFee:      priorityFee,
		TotalFee:         go_decimal.Decimal.MustStart(baseFee).MustAddForString(priorityFee),
		ComputeUnitPrice: uint64(computeUnitPrice),
	}, nil
}
