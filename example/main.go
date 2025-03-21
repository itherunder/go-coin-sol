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
		solana.MustPublicKeyFromBase58("4iucvyLyWumRqkL1WQXvcu1RyzPboczkKFjmEeR9WAN1"),
		solana.MustPublicKeyFromBase58("DP4MXhEhe9USfRr1pdDazEdqVftSVH95X7fAXG2epump"),
	)
	if err != nil {
		return err
	}
	fmt.Println(userSOLTokenAccount.String())
	return nil
}
