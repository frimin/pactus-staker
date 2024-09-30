# Pactus-staker

Easy-to-use pactus automatic stake tool

# Install & Build 

Download & install golang 1.22.2 or newer -> [link](https://go.dev/dl/)

Download this project and enter root folder, run simple build command:

    go build

Copy example config to `config.yml` (or other name)

Run without arguments or with arguments (if you need other config filename):

    ./pactus-staker 
    ./pactus-staker -config config.yml

# Configuration Instructions

`options.grpc_server`: Connect pactus blockchain node grpc address, recommend connect to local node.
`options.retry_delay`: When the action fails, the retry wait time and number of times are listed. If all fail, the action is skipped.

`pipeline[*].name`: Pipine name, A pipeline supports multiple actions
`pipelins[*].reward.wallets`:  Broadcast the bond command from the reward address in the specified wallet file
`pipelins[*].reward.wallets[*].path` : Wallet file path
`pipelins[*].reward.wallets[*].password`: Wallet flle password