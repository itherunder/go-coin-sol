package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	Proxy = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("Bvc2iuuybMJRsRfJUbjrmZTpks16UCLoYHn2KdKNeg9m"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("Bvc2iuuybMJRsRfJUbjrmZTpks16UCLoYHn2KdKNeg9m"),
	}
)
