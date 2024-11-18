package pipline

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/frimin/pactus-staker/config"
	"github.com/frimin/pactus-staker/pipline/action"
	"github.com/pactus-project/pactus/types/amount"
	"github.com/pactus-project/pactus/wallet"
	pactus "github.com/pactus-project/pactus/www/grpc/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Pipline interface {
	Run() error
}

type pipline struct {
	ctx              context.Context
	name             string
	actions          []action.Action
	walletList       []*wallet.Wallet
	walletPassword   []string
	accountAddresses map[string]int

	blockchainClient pactus.BlockchainClient
}

func (p *pipline) Run() error {
	return nil
}

func (p *pipline) GetName() string {
	return p.name
}

func (p *pipline) GetActions() []action.Action {
	return p.actions
}

func (p *pipline) GetAllBalance() ([]string, []amount.Amount, error) {
	addresses := make([]string, 0)
	amounts := make([]amount.Amount, 0)

	for _, wlt := range p.walletList {
		for _, address := range wlt.AllAccountAddresses() {
			amount, _ := wlt.Balance(address.Address)

			addresses = append(addresses, address.Address)
			amounts = append(amounts, amount)
		}
	}

	return addresses, amounts, nil
}

func (p *pipline) GetAccountWallet(address string) (*wallet.Wallet, string) {
	if i, ok := p.accountAddresses[address]; ok {
		return p.walletList[i], p.walletPassword[i]
	}

	return nil, ""
}

func (p *pipline) GetBlockchainClient() pactus.BlockchainClient {
	return p.blockchainClient
}

func (p *pipline) GetValidatorStake(address string) (amount.Amount, *pactus.ValidatorInfo, error) {
	resp, err := p.GetBlockchainClient().GetValidator(p.ctx, &pactus.GetValidatorRequest{Address: address})

	var validatorInfo *pactus.ValidatorInfo

	if resp != nil && resp.Validator != nil {
		validatorInfo = resp.Validator
	}

	if err != nil {
		if strings.Contains(err.Error(), "validator not found") {
			return amount.Amount(0), validatorInfo, nil
		}
		return amount.Amount(0), validatorInfo, err
	}

	return amount.Amount(resp.Validator.Stake), validatorInfo, nil
}

func (p *pipline) GetValidator(address string) (*pactus.GetValidatorResponse, error) {
	return p.GetBlockchainClient().GetValidator(p.ctx, &pactus.GetValidatorRequest{Address: address})
}

const TIMEOUT = 10 * time.Second

func (p *pipline) connect(optionsConfig *config.Options) error {
	conn, err := grpc.NewClient(optionsConfig.GrpcServer,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(_ context.Context, s string) (net.Conn, error) {
			return net.DialTimeout("tcp", s, TIMEOUT)
		}))

	if err != nil {
		return err
	}

	p.blockchainClient = pactus.NewBlockchainClient(conn)

	// Check if client is responding
	_, err = p.blockchainClient.GetBlockchainInfo(p.ctx,
		&pactus.GetBlockchainInfoRequest{})
	if err != nil {
		_ = conn.Close()

		return err
	}

	return nil
}

func createPipline(optionsConfig *config.Options, piplineConfig config.Pipline) (*pipline, error) {
	pip := &pipline{
		ctx:              context.Background(),
		name:             piplineConfig.Name,
		actions:          make([]action.Action, 0),
		walletList:       make([]*wallet.Wallet, 0),
		walletPassword:   make([]string, 0),
		accountAddresses: make(map[string]int),
	}

	err := pip.connect(optionsConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}

	for _, rewardWallet := range piplineConfig.Reward.Wallets {
		wlt, err := wallet.Open(rewardWallet.Path, false,
			wallet.WithTimeout(TIMEOUT),
			wallet.WithCustomServers([]string{optionsConfig.GrpcServer}))

		if err != nil {
			return nil, fmt.Errorf("failed to open wallet: %w", err)
		}

		pip.walletList = append(pip.walletList, wlt)
		pip.walletPassword = append(pip.walletPassword, rewardWallet.Password)
	}

	for i, wlt := range pip.walletList {
		for _, address := range wlt.AllAccountAddresses() {
			pip.accountAddresses[address.Address] = i
		}
	}

	log.Printf("Pipline %s has %d wallets", piplineConfig.Name, len(pip.walletList))

	addresses, amounts, err := pip.GetAllBalance()

	if err != nil {
		return nil, fmt.Errorf("failed to get all balances: %w", err)
	}

	totalAmount := amount.Amount(0)

	for i, address := range addresses {
		log.Printf("%d - %s - balance: %s", i+1, address, amounts[i].String())

		totalAmount += amounts[i]
	}

	log.Printf("Total balance: %s", totalAmount.String())

	for i, actionConfig := range piplineConfig.Actions {
		action, err := action.CreateAction(pip, i, optionsConfig, &actionConfig)

		if err != nil {
			return nil, err
		}

		pip.actions = append(pip.actions, action)
	}

	if len(pip.actions) != 1 {
		return nil, fmt.Errorf("limit to one action per pipline")
	}

	return pip, nil
}
