# Pactus-staker

Pactus automatic stake tool

# Install & Build 

Download & install golang 1.23.2 -> [link](https://go.dev/dl/)

Download this project and enter root folder, run simple build command:

    go build

Copy example config `config.example.yml` to `config.yml` (or other name)

Run without arguments or with arguments (if you need other config filename):

    ./pactus-staker 
    ./pactus-staker -config config.yml

## Windows support

Download & install golang with setup: [go1.23.2.windows-amd64.msi](https://go.dev/dl/go1.23.2.windows-amd64.msi)

Click `build_run.bat` file build and run

# Configuration reference 

`options.grpc_server`: Connect to a Pactus blockchain node via gRPC. It is recommended to connect to a local node. Use `localhost:50051` for a local mainnet node or `localhost:50052` for a local testnet node. 

`options.retry_delay`: When the action fails, the retry wait time and number of attempts are specified. If all attempts fail, the action is skipped.

`options.reserve_fees`: This balance is reserved in each account for transfer fees.

`pipeline[*].name`: Pipine name, a pipeline supports multiple actions

`pipelins[*].reward.wallets`: Broadcast the bond command from the reward address specified in the wallet file. 

`pipelins[*].reward.wallets[*].path` : Wallet file path

`pipelins[*].reward.wallets[*].password`: Wallet flle password

`pipelins[*].actions`: Actions for pipline 

## Bond action

`pipeline[*].actions[*].type` = `"bond"`

`pipeline[*].actions.time`: List of times that trigger this action

`pipeline[*].actions.targets`: Staking address target wallet list (must use a wallet file and not a validator address, as the public key is required when first create a validator.) **The target wallet does not require a password**

# Configuration Example

In the case of a single wallet file, accounts bond to it self validators, Do this once a day:

    options:
        grpc_server: "localhost:50052"
        retry_delay: [10,15,30,60]
        reserve_fees: 1
    pipeline:
      - name: myname1
        reward: 
            wallets:
                - path: ./default_wallet
                password: 123456
            actions:
            - type: "bond"
                time: [ "00:00" ]
                targets:
                - ./default_wallet

The bond action will attempt to bond each account to each validator in sequence, provided that the account's balance is sufficient.


Like this:

    reward_wallet1.account1 -> target_wallet1.validator1
    reward_wallet1.account1 -> target_wallet1.validator2
    reward_wallet1.account1 -> target_wallet1.validator3
    ...
    reward_wallet1.account1 -> target_wallet2.validator1
    reward_wallet1.account1 -> target_wallet2.validator2
    reward_wallet1.account1 -> target_wallet2.validator3
    ...
    reward_wallet2.account1 -> target_wallet1.validator1
    reward_wallet2.account1 -> target_wallet1.validator2
    reward_wallet2.account1 -> target_wallet1.validator3
    ...
    reward_wallet2.account1 -> target_wallet2.validator1
    reward_wallet2.account1 -> target_wallet2.validator2
    reward_wallet2.account1 -> target_wallet2.validator3
    ...



You can define multiple time points to trigger action execution. While this is possible, it's recommended to execute once a day, as the validator will not enter the committee for one hour after a staking operation, during which you won't receive any block rewards.

    pipeline:
      - name: myname1
        reward: 
            ...
            actions:
            - type: "bond"
                time: [ "00:00", "6:00", "12:00", "18:00" ]
                ...


By default, 1 PAC is reserved as a fee; however, you can lower this amount.

    options:
        ...
        reserve_fees: 0.01
        ...

# Donation

If this tool is helpful to you, you can send some PACs to the donation address

    pc1rrrrrrrjwxqj8v3pvhayensqjhhwvm2s2yfu8g8