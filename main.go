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
	walletAddressesStr := os.Getenv("ADDRESSES")
	if walletAddressesStr == "" {
		fmt.Println("ADDRESSES environment variable is not set")
		return
	}

	var walletAddresses []string
	err := json.Unmarshal([]byte(walletAddressesStr), &walletAddresses)
	if err != nil {
		fmt.Printf("Error parsing ADDRESSES: %v\n", err)
		return
	}

	var publicKeys []solana.PublicKey
	for _, address := range walletAddresses {
		publicKey, err := solana.PublicKeyFromBase58(address)
		if err != nil {
			log.Fatalf("Invalid wallet address: %s", err)
		}

		publicKeys = append(publicKeys, publicKey)
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

	intervalStr := os.Getenv("INTERVAL")
	interval, err := strconv.ParseUint(intervalStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid INTERVAL: %s", err)
	}

	for _, publicKey := range publicKeys {
		checkBalance(rpcUrl, publicKey, minBalance)
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, publicKey := range publicKeys {
			checkBalance(rpcUrl, publicKey, minBalance)
		}
	}
}

func checkBalance(rpcUrl string, publicKey solana.PublicKey, minBalance uint64) {
	client := rpc.New(rpcUrl)

	balance, err := client.GetBalance(context.Background(), publicKey, rpc.CommitmentConfirmed)
	if err != nil {
		log.Printf("Error retrieving balance: %s", err)
	} else {
		solBalance := lamportsToSol(balance.Value)
		fmt.Printf("Balance of %s is %f SOL\n", publicKey.String(), solBalance)

		if balance.Value < minBalance {
			summary := fmt.Sprintf("KEEPER LOW BALANCE %s\n", publicKey.String())
			message := fmt.Sprintf("%s Left %f SOL\n", publicKey.String(), solBalance)
			discordWebhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")
			if discordWebhookUrl != "" {
				sendDiscordWebhook(discordWebhookUrl, message)
			}

			betterStackBearer := os.Getenv("BETTERSTACK_TOKEN")
			if betterStackBearer != "" {
				createBetterStackIncident(betterStackBearer, summary, message)
			}
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

func createBetterStackIncident(bearer string, summary string, description string) error {
	url := "https://uptime.betterstack.com/api/v2/incidents"
	requesterEmail := os.Getenv("REQUESTER_EMAIL")

	requestData := struct {
		Summary        string `json:"summary"`
		RequesterEmail string `json:"requester_email"`
		Description    string `json:"description"`
	}{
		Summary:        summary,
		RequesterEmail: requesterEmail,
		Description:    description,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("error marshalling request data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}

func lamportsToSol(lamports uint64) float64 {
	return float64(lamports) / 1000000000
}
