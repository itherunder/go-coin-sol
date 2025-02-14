package associated_token_account

import (
	"context"
	"encoding/json"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

type AssociatedTokenAccountDataType struct {
	Parsed struct {
		Info struct {
			IsNative          bool   `json:"isNative"`
			Mint              string `json:"mint"`
			Owner             string `json:"owner"`
			RentExemptReserve struct {
				Amount         string  `json:"amount"`
				Decimals       uint64  `json:"decimals"`
				UIAmount       float64 `json:"uiAmount"`
				UIAmountString string  `json:"uiAmountString"`
			} `json:"rentExemptReserve"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount         string  `json:"amount"`
				Decimals       uint64  `json:"decimals"`
				UIAmount       float64 `json:"uiAmount"`
				UIAmountString string  `json:"uiAmountString"`
			} `json:"tokenAmount"`
		} `json:"info"`
		Type string `json:"type"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   uint64 `json:"space"`
}

func GetAssociatedTokenAccountData(
	rpcClient *rpc.Client,
	associatedTokenAccount solana.PublicKey,
) (*AssociatedTokenAccountDataType, error) {
	info, err := rpcClient.GetAccountInfoWithOpts(context.Background(), associatedTokenAccount, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingJSONParsed,
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var data AssociatedTokenAccountDataType
	err = json.Unmarshal(info.Value.Data.GetRawJSON(), &data)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &data, nil
}

func GetAssociatedTokenAccountDatas(
	rpcClient *rpc.Client,
	accounts []solana.PublicKey,
) ([]*AssociatedTokenAccountDataType, error) {
	results := make([]*AssociatedTokenAccountDataType, 0)
	result, err := rpcClient.GetMultipleAccountsWithOpts(
		context.Background(),
		accounts,
		&rpc.GetMultipleAccountsOpts{
			Encoding:   solana.EncodingJSONParsed,
			Commitment: rpc.CommitmentConfirmed,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	for _, account := range result.Value {
		if account == nil || account.Data == nil {
			results = append(results, nil)
			continue
		}
		var data AssociatedTokenAccountDataType
		err = json.Unmarshal(account.Data.GetRawJSON(), &data)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		results = append(results, &data)
	}

	return results, nil
}
