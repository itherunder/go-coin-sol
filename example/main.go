package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
)

func main() {
	err := do()
	if err != nil {
		log.Fatal(err)
	}
}

func do() error {
	envMap, _ := godotenv.Read("./.env")
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	userSOLTokenAccount, _, err := solana.FindAssociatedTokenAddress(
		solana.MustPublicKeyFromBase58("4AZRPNEfCJ7iw28rJu5aUyeQhYcvdcNm8cswyL51AY9i"),
		solana.SolMint,
	)
	if err != nil {
		return err
	}
	fmt.Println(userSOLTokenAccount.String())
	return nil
}
