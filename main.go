package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	walletAddress := os.Getenv("ADDRESS")
	publicKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		log.Fatalf("Invalid wallet address: %s", err)
	}

	rpcUrl := os.Getenv("RPC")
	if rpcUrl == "" {
		rpcUrl = rpc.MainNetBeta_RPC
	}

	minBalanceStr := os.Getenv("MINIMUM_LAMPORTS")
	minBalance, err := strconv.ParseUint(minBalanceStr, 10, 64)
	if err != nil {
		log.Fatalf("Error converting MINIMUM_LAMPORTS to integer: %s", err)
	}

	webhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")
	if rpcUrl == "" {
		log.Fatalf("Missing Webhook URL")
	}

	intervalStr := os.Getenv("INTERVAL")
	interval, err := strconv.ParseUint(intervalStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid INTERVAL: %s", err)
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		checkBalance(rpcUrl, publicKey, minBalance, webhookUrl)
	}

}

func checkBalance(rpcUrl string, publicKey solana.PublicKey, minBalance uint64, webhookUrl string) {
	client := rpc.New(rpcUrl)

	balance, err := client.GetBalance(context.Background(), publicKey, rpc.CommitmentConfirmed)
	if err != nil {
		log.Printf("Error retrieving balance: %s", err)
	} else {
		fmt.Printf("Balance of %s is %d lamports\n", publicKey.String(), balance.Value)

		if balance.Value < minBalance {
			message := fmt.Sprintf("LOW BALANCE Balance of %s is %d lamports\n", publicKey.String(), balance.Value)
			sendDiscordWebhook(webhookUrl, message)
		}
	}
}

func sendDiscordWebhook(webhookURL string, message string) {
	body := map[string]string{"content": message}

	bytesRepresentation, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	_, err = http.Post(webhookURL, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		panic(err)
	}
}
