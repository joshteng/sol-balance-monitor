# SOL Balance Monitor

Monitor SOL balance of a given address and sends a Discord Webhook if balance is below set threshold

Github Repo: https://github.com/joshteng/sol-balance-monitor

## Running Based on Docker Hub image:
1. Create `docker-compose.yml` with the following content
    ```yml
    version: '3.8'

    services:
      sol-balance-monitor:
        image: joshteng/sol-balance-monitor:latest
        environment:
          - INTERVAL
          - MINIMUM_LAMPORTS
          - RPC
          - ADDRESS
          - DISCORD_WEBHOOK_URL
        restart: always
    ```
2. Set environment variables inside `.env` in the same directory as `docker-compose.yml` or simply hardcode it in the Docker compose file
    ```yml
    version: '3.8'

    services:
      sol-balance-monitor:
        image: joshteng/sol-balance-monitor:latest
        environment:
          - INTERVAL=5
          - MINIMUM_LAMPORTS=50000000000
          - RPC=https://api.devnet.solana.com
          - ADDRESS=GUhB2ohrfqWspztgCrQpAmeVFBWmnWYhPcZuwY52WWRe
          - DISCORD_WEBHOOK_URL=https://
        restart: always
    ```
3. then run `docker compose up`
