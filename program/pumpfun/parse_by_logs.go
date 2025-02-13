package pumpfun

import (
	"encoding/base64"
	"fmt"
	"strings"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	pumpfun_type "github.com/pefish/go-coin-sol/program/pumpfun/type"
	raydium_amm_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	util "github.com/pefish/go-coin-sol/util"
)

func ParseSwapByLogs(logs []string) []*type_.SwapDataType {
	if strings.Contains(strings.Join(logs, ""), "failed") {
		return nil
	}

	swaps := make([]*type_.SwapDataType, 0)

	isSwap := false
	stack := util.NewStack()
	for _, log := range logs {
		pushPrefix := fmt.Sprintf("Program %s invoke", pumpfun_constant.Pumpfun_Program)
		popLog := fmt.Sprintf("Program %s success", pumpfun_constant.Pumpfun_Program)
		if strings.HasPrefix(log, pushPrefix) {
			stack.Push(log)
			continue
		}
		if log == popLog {
			stack.Pop()
			continue
		}
		if stack.Size() == 0 {
			continue
		}

		if log == "Program log: Instruction: Buy" ||
			log == "Program log: Instruction: Sell" {
			isSwap = true
			continue
		}
		if !isSwap {
			continue
		}

		if !strings.HasPrefix(log, "Program data:") {
			continue
		}
		data := log[14:]
		if len(data) < 150 {
			continue
		}
		b, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			continue
		}
		var logObj struct {
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
		err = bin.NewBorshDecoder(b).Decode(&logObj)
		if err != nil {
			// 说明记录的不是 swap 信息
			continue
		}
		if logObj.VirtualSolReserves == 0 ||
			logObj.VirtualTokenReserves == 0 ||
			logObj.Timestamp == 0 {
			continue
		}
		var swapData type_.SwapDataType
		if logObj.IsBuy {
			swapData = type_.SwapDataType{
				InputAddress:             solana.SolMint,
				OutputAddress:            logObj.Mint,
				InputAmountWithDecimals:  logObj.SOLAmount,
				OutputAmountWithDecimals: logObj.TokenAmount,
				InputDecimals:            constant.SOL_Decimals,
				OutputDecimals:           pumpfun_constant.Pumpfun_Token_Decimals,
				UserAddress:              logObj.User,

				PairAddress: solana.PublicKey{},
				InputVault:  solana.PublicKey{},
				OutputVault: solana.PublicKey{},

				ParsedKeys: nil,
				ExtraDatas: &pumpfun_type.ExtraDatasType{
					ReserveSOLAmountWithDecimals:   logObj.VirtualSolReserves,
					ReserveTokenAmountWithDecimals: logObj.VirtualTokenReserves,
					Timestamp:                      uint64(logObj.Timestamp * 1000),
				},

				Program:  pumpfun_constant.Pumpfun_Program,
				Keys:     nil,
				MethodId: "",
			}
		} else {
			swapData = type_.SwapDataType{
				InputAddress:             logObj.Mint,
				OutputAddress:            solana.SolMint,
				InputAmountWithDecimals:  logObj.TokenAmount,
				OutputAmountWithDecimals: logObj.SOLAmount,
				InputDecimals:            pumpfun_constant.Pumpfun_Token_Decimals,
				OutputDecimals:           constant.SOL_Decimals,
				UserAddress:              logObj.User,

				PairAddress: solana.PublicKey{},
				InputVault:  solana.PublicKey{},
				OutputVault: solana.PublicKey{},

				ParsedKeys: nil,
				ExtraDatas: &pumpfun_type.ExtraDatasType{
					ReserveSOLAmountWithDecimals:   logObj.VirtualSolReserves,
					ReserveTokenAmountWithDecimals: logObj.VirtualTokenReserves,
					Timestamp:                      uint64(logObj.Timestamp * 1000),
				},

				Program:  pumpfun_constant.Pumpfun_Program,
				Keys:     nil,
				MethodId: "",
			}
		}

		swaps = append(swaps, &swapData)
	}

	return swaps
}

func ParseCreateByLogs(logs []string) *pumpfun_type.CreateDataType {
	if strings.Contains(strings.Join(logs, ""), "failed") {
		return nil
	}

	isCreate := false
	stack := util.NewStack()
	for _, log := range logs {
		pushPrefix := fmt.Sprintf("Program %s invoke", pumpfun_constant.Pumpfun_Program)
		popLog := fmt.Sprintf("Program %s success", pumpfun_constant.Pumpfun_Program)
		if strings.HasPrefix(log, pushPrefix) {
			stack.Push(log)
			continue
		}
		if log == popLog {
			stack.Pop()
			continue
		}
		if stack.Size() == 0 {
			continue
		}

		if log == "Program log: Instruction: Create" {
			isCreate = true
			continue
		}
		if !isCreate {
			continue
		}

		if !strings.HasPrefix(log, "Program data:") {
			continue
		}
		data := log[14:]
		if len(data) < 200 {
			continue
		}
		b, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			continue
		}
		var logObj struct {
			Id           uint64           `json:"id"`
			Name         string           `json:"name"`
			Symbol       string           `json:"symbol"`
			URI          string           `json:"uri"`
			Mint         solana.PublicKey `json:"mint"`
			BondingCurve solana.PublicKey `json:"bondingCurve"`
			User         solana.PublicKey `json:"user"`
		}
		err = bin.NewBorshDecoder(b).Decode(&logObj)
		if err != nil {
			continue
		}
		return &pumpfun_type.CreateDataType{
			Name:                logObj.Name,
			Symbol:              logObj.Symbol,
			URI:                 logObj.URI,
			UserAddress:         logObj.User,
			BondingCurveAddress: logObj.BondingCurve,
			TokenAddress:        logObj.Mint,
		}
	}

	return nil
}

func IsRemoveLiqByLogs(logs []string) bool {
	if strings.Contains(strings.Join(logs, ""), "failed") {
		return false
	}

	stack := util.NewStack()
	for _, log := range logs {
		pushPrefix := fmt.Sprintf("Program %s invoke", pumpfun_constant.Pumpfun_Program)
		popLog := fmt.Sprintf("Program %s success", pumpfun_constant.Pumpfun_Program)
		if strings.HasPrefix(log, pushPrefix) {
			stack.Push(log)
			continue
		}
		if log == popLog {
			stack.Pop()
			continue
		}
		if stack.Size() == 0 {
			continue
		}

		if log == "Program log: Instruction: Withdraw" {
			return true
		}
	}

	return false
}

func IsAddLiqByLogs(network rpc.Cluster, logs []string) bool {
	if strings.Contains(strings.Join(logs, ""), "failed") {
		return false
	}

	stack := util.NewStack()
	for _, log := range logs {

		pushPrefix := fmt.Sprintf("Program %s invoke", raydium_amm_constant.Raydium_Liquidity_Pool_V4[network])
		popLog := fmt.Sprintf("Program %s success", raydium_amm_constant.Raydium_Liquidity_Pool_V4[network])
		if strings.HasPrefix(log, pushPrefix) {
			stack.Push(log)
			continue
		}
		if log == popLog {
			stack.Pop()
			continue
		}
		if stack.Size() == 0 {
			continue
		}

		if log == "Program log: Instruction: MintTo" {
			return true
		}
	}

	return false
}
