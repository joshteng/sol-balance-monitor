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

type Accounts struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	MinLamports int64  `json:"minLamports"`
}

var lastAlerts = make(map[string]time.Time)

func main() {
	accounts := getAccounts()

	rpcUrl := os.Getenv("RPC")
	if rpcUrl == "" {
		rpcUrl = rpc.MainNetBeta_RPC
	}

	intervalStr := os.Getenv("INTERVAL")
	interval, err := strconv.ParseUint(intervalStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid INTERVAL: %s", err)
	}

	alertIntervalStr := os.Getenv("ALERT_INTERVAL")
	alertInterval, err := strconv.ParseUint(alertIntervalStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid ALERT_INTERVAL: %s", err)
	}

	checkBalances(rpcUrl, accounts, alertInterval)

	betterStackHeartbeatUrl := os.Getenv("BETTERSTACK_HEARTBEAT_URL")
	if betterStackHeartbeatUrl != "" {
		go betterStackHeartbeat(betterStackHeartbeatUrl)
	}

	monitorAccounts(rpcUrl, accounts, interval, alertInterval)
}

func monitorAccounts(rpcUrl string, accounts []Accounts, interval uint64, alertInterval uint64) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		checkBalances(rpcUrl, accounts, alertInterval)
	}
}

func checkBalances(rpcUrl string, accounts []Accounts, alertInterval uint64) {
	for _, account := range accounts {
		checkBalance(rpcUrl, account, alertInterval)
	}
}

func betterStackHeartbeat(url string) {
	if _, err := http.Get(url); err != nil {
		log.Print(err)
	} else {
		log.Println("Sent Betterstack heartbeat")
	}

	ticker := time.NewTicker(time.Duration(1) * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if _, err := http.Get(url); err != nil {
			log.Print(err)
		} else {
			log.Println("Sent Betterstack heartbeat")
		}
	}
}

func getAccounts() []Accounts {
	accountsStr := os.Getenv("ACCOUNTS")
	if accountsStr == "" {
		log.Fatalf("ACCOUNTS environment variable is not set")
	}

	var accounts []Accounts
	err := json.Unmarshal([]byte(accountsStr), &accounts)
	if err != nil {
		log.Fatalf("error parsing ADDRESSES: %v", err)
	}

	for _, account := range accounts {
		_, err := solana.PublicKeyFromBase58(account.Address)
		if err != nil {
			log.Fatalf("Invalid wallet address: %s", err)
		}
	}

	return accounts
}

func checkBalance(rpcUrl string, account Accounts, alertInterval uint64) {
	client := rpc.New(rpcUrl)

	publicKey, _ := solana.PublicKeyFromBase58(account.Address)

	balance, err := client.GetBalance(context.Background(), publicKey, rpc.CommitmentConfirmed)
	if err != nil {
		log.Printf("Error retrieving balance: %s", err)
	} else {
		solBalance := lamportsToSol(balance.Value)
		fmt.Printf("%s SOL Balance: %f (%s)\n", account.Name, solBalance, publicKey.String())

		if balance.Value < uint64(account.MinLamports) {
			lastAlert, hasAlerted := lastAlerts[publicKey.String()]
			if !hasAlerted || time.Since(lastAlert) > time.Duration(alertInterval)*time.Second {
				summary := fmt.Sprintf("%s LOW BALANCE (%s) Left %f SOL\n", account.Name, publicKey.String(), solBalance)
				message := fmt.Sprintf("%s Left %f SOL (%s)\n", account.Name, solBalance, publicKey.String())
				discordWebhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")
				if discordWebhookUrl != "" {
					sendDiscordWebhook(discordWebhookUrl, message)
				}

				betterStackBearer := os.Getenv("BETTERSTACK_TOKEN")
				if betterStackBearer != "" {
					createBetterStackIncident(betterStackBearer, summary, message)
				}

				lastAlerts[publicKey.String()] = time.Now()
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
