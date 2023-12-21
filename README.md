# SOL Balance Monitor

Monitor SOL balance of a given address and sends a Discord Webhook and / or create Betterstack Incident if balance is below set threshold

Github Repo: https://github.com/joshteng/sol-balance-monitor
Docker Hub: https://hub.docker.com/repository/docker/joshteng/sol-balance-monitor/general

## Running Based on Docker Hub image:
1. Create `docker-compose.yml` with the following content
    ```yml
    version: '3.8'

    services:
      sol-balance-monitor:
        image: joshteng/sol-balance-monitor:latest
        environment:
          - ACCOUNTS
          - BETTERSTACK_HEARTBEAT_URL
          - BETTERSTACK_TOKEN
          - DISCORD_WEBHOOK_URL
          - INTERVAL
          - REQUESTER_EMAIL
          - RPC
        restart: always
    ```
2. Create env variables
    ```
    INTERVAL=300
    BETTERSTACK_HEARTBEAT_URL= # setting this will send a heartbeat to Betterstack every min
    DISCORD_WEBHOOK_URL=
    REQUESTER_EMAIL='bot@example.com'
    RPC=https://api.devnet.solana.com
    ACCOUNTS='[{"name":"Wallet 1","address":"GUhB2ohrfqWspztgCrQpAmeVFBWmnWYhPcZuwY52WWRe","minLamports":50000000000},{"name":"Wallet 2","address":"GUhB2ohrfqWspztgCrQpAmeVFBWmnWYhPcZuwY52WWRe","minLamports":50000000000}]'
    BETTERSTACK_TOKEN=""
    ```
3. then run `docker compose up`


## Deploying to Docker Hub
```sh
docker build . -t sol-balance-monitor
docker tag sol-balance-monitor joshteng/sol-balance-monitor:latest # or get the image id from docker dashboard or docker images and run docker tag <image-id> joshteng/sol-balance-monitor:latest
docker push joshteng/sol-balance-monitor:latest
```
