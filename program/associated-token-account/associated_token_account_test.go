package associated_token_account

import (
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	go_test_ "github.com/pefish/go-test"
)

func TestGetAssociatedTokenAccountData(t *testing.T) {
	return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	r, err := GetAssociatedTokenAccountData(client, solana.MustPublicKeyFromBase58("DpUcSNu7gh4P2fXMx2s8ub3dVQC5dwjX6CAo2KRekaBt"))
	go_test_.Equal(t, nil, err)
	fmt.Println(r)
}

func TestGetAssociatedTokenAccountDatas(t *testing.T) {
	return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	r, err := GetAssociatedTokenAccountDatas(client, []solana.PublicKey{
		solana.MustPublicKeyFromBase58("DpUcSNu7gh4P2fXMx2s8ub3dVQC5dwjX6CAo2KRekaBt"),
		solana.MustPublicKeyFromBase58("6shUsSy5zF4X5UjttcEDrhuKF6bdtkiEt1LemsFjoN6m"),
	})
	go_test_.Equal(t, nil, err)
	for _, a := range r {
		fmt.Println(a)
	}
}
