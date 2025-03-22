package pumpfun

import (
	"fmt"
	"strconv"
	"strings"

	solana "github.com/gagliardetto/solana-go"
	okx_aggregation_constant "github.com/itherunder/go-coin-sol/program/okx-aggregation/constant"
	okx_aggregation_type "github.com/itherunder/go-coin-sol/program/okx-aggregation/type"
	type_ "github.com/itherunder/go-coin-sol/type"
	util "github.com/itherunder/go-coin-sol/util"
)

func ParseSwapByLogs(logs []string) *okx_aggregation_type.SwapDataType {
	if strings.Contains(strings.Join(logs, ""), "failed") {
		return nil
	}

	var result *okx_aggregation_type.SwapDataType
	stack := util.NewStack()
	okxProgramStartIndex := 0
	for i, log := range logs {
		pushPrefix := fmt.Sprintf("Program %s invoke", okx_aggregation_constant.Aggregation_Router_V2)
		popLog := fmt.Sprintf("Program %s success", okx_aggregation_constant.Aggregation_Router_V2)
		if strings.HasPrefix(log, pushPrefix) {
			stack.Push(log)
			okxProgramStartIndex = i
			result = &okx_aggregation_type.SwapDataType{}
			continue
		}
		if log == popLog {
			stack.Pop()
			continue
		}
		if stack.Size() == 0 {
			continue
		}

		if okxProgramStartIndex == i-1 {
			if log != "Program log: Instruction: Swap2" {
				return nil
			}
			fromTokenAddress := strings.TrimPrefix(logs[i+1], "Program log: ")
			toTokenAddress := strings.TrimPrefix(logs[i+2], "Program log: ")
			result.UserAddress = solana.MustPublicKeyFromBase58(strings.TrimPrefix(logs[i+3], "Program log: "))
			if fromTokenAddress == solana.SolMint.String() {
				result.Type = type_.SwapType_Buy
				result.TokenAddress = solana.MustPublicKeyFromBase58(toTokenAddress)
			} else {
				result.Type = type_.SwapType_Sell
				result.TokenAddress = solana.MustPublicKeyFromBase58(fromTokenAddress)
			}
		}

		if strings.HasPrefix(log, "Program log: after_source_balance: ") {
			arr := strings.Split(strings.TrimPrefix(log, "Program log: "), ", ")

			afterSourceBalanceWithDecimals, _ := strconv.ParseUint(strings.Split(arr[0], ": ")[1], 10, 64)
			afterDestinationBalanceWithDecimals, _ := strconv.ParseUint(strings.Split(arr[1], ": ")[1], 10, 64)
			sourceTokenChangeWithDecimals, _ := strconv.ParseUint(strings.Split(arr[2], ": ")[1], 10, 64)
			destinationTokenChangeWithDecimals, _ := strconv.ParseUint(strings.Split(arr[3], ": ")[1], 10, 64)

			if result.Type == type_.SwapType_Buy {
				result.SOLAmountWithDecimals = sourceTokenChangeWithDecimals
				result.TokenAmountWithDecimals = destinationTokenChangeWithDecimals
				result.UserTokenBalanceWithDecimals = afterDestinationBalanceWithDecimals
			} else {
				result.SOLAmountWithDecimals = destinationTokenChangeWithDecimals
				result.TokenAmountWithDecimals = sourceTokenChangeWithDecimals
				result.UserTokenBalanceWithDecimals = afterSourceBalanceWithDecimals
			}
			return result
		}
	}

	return result
}
