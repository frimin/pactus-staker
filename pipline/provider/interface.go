package provider

import (
	"github.com/pactus-project/pactus/types/amount"
	"github.com/pactus-project/pactus/wallet"
	pactus "github.com/pactus-project/pactus/www/grpc/gen/go"
)

type PiplineProvider interface {
	GetName() string
	GetAllBalance() ([]string, []amount.Amount, error)
	GetAccountWallet(address string) (*wallet.Wallet, string)
	GetBlockchainClient() pactus.BlockchainClient
	GetValidatorStake(address string) (amount.Amount, *pactus.ValidatorInfo, error)
}
