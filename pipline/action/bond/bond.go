package bond

import (
	"fmt"
	"log"
	"time"

	"github.com/frimin/pactus-staker/config"
	"github.com/frimin/pactus-staker/pipline/provider"
	"github.com/pactus-project/pactus/types/amount"
	"github.com/pactus-project/pactus/types/tx/payload"
	"github.com/pactus-project/pactus/wallet"
	"github.com/pactus-project/pactus/wallet/vault"
	pactus "github.com/pactus-project/pactus/www/grpc/gen/go"
)

type BondAction struct {
	validatorAddresses   []string
	validatorWallet      map[string]*wallet.Wallet
	validatorAddressInfo map[string]vault.AddressInfo
	pipline              provider.PiplineProvider
	time                 []string
	reserveFees          amount.Amount
}

func CreateBondAction(pipline provider.PiplineProvider, index int, optionsConfig *config.Options, actionConfig *config.Action) (*BondAction, error) {
	reserveFees, err := amount.NewAmount(optionsConfig.ReserveFees)

	if err != nil {
		return nil, fmt.Errorf("failed to create reserve fees: %w", err)
	}

	action := &BondAction{
		validatorAddresses:   make([]string, 0),
		validatorWallet:      make(map[string]*wallet.Wallet),
		validatorAddressInfo: make(map[string]vault.AddressInfo),
		pipline:              pipline,
		time:                 actionConfig.Time,
		reserveFees:          reserveFees,
	}

	processedAddresses := map[string]bool{}

	for _, target := range actionConfig.Targets {
		wlt, err := wallet.Open(target, false)

		if err != nil {
			log.Printf("Failed to open wallet: %v", err)
		}

		for _, address := range wlt.AllValidatorAddresses() {
			if _, ok := processedAddresses[address.Address]; ok {
				log.Printf("ignore duplicate target address: %s", address.Address)
				continue
			}

			action.validatorAddresses = append(action.validatorAddresses, address.Address)
			processedAddresses[address.Address] = true

			action.validatorWallet[address.Address] = wlt

			for _, address := range wlt.AllValidatorAddresses() {
				action.validatorAddressInfo[address.Address] = address
			}
		}
	}

	log.Printf("Pipline %s action %d has %d bond targets", action.pipline.GetName(), index, len(action.validatorAddresses))

	totalStake := amount.Amount(0)

	for i, address := range action.validatorAddresses {
		amount, validatorInfo, err := action.pipline.GetValidatorStake(address)
		if err != nil {
			return nil, err
		}

		availabilityScore := 1.0

		if validatorInfo != nil {
			availabilityScore = validatorInfo.AvailabilityScore
		}

		log.Printf("%d - %s - stake: %s (score: %v)", i+1, address, amount.String(), availabilityScore)
		totalStake += amount
	}

	log.Printf("Total stake: %s", totalStake.String())

	return action, nil
}

func (p *BondAction) GetTime() []string {
	return p.time
}

func (p *BondAction) GetName() string {
	return "bond"
}

const (
	MIN_STAKE = amount.Amount(1000000000)
	//KEEP_FOR_FEE = MIN_STAKE
	MAX_STAKE = amount.Amount(1000000000000)

	NEAR_MAX_STAKE = MAX_STAKE - MIN_STAKE
)

func (p *BondAction) Run() error {
	addresses, amounts, err := p.pipline.GetAllBalance()

	if err != nil {
		return fmt.Errorf("failed to get all balances: %w", err)
	}

	type stakeCacheInfo struct {
		amount        amount.Amount
		validatorInfo *pactus.ValidatorInfo
	}

	stakeCacheMap := make(map[string]*stakeCacheInfo)

	broadcastCount := 0

	for accountIndex, accountAddress := range addresses {
		/*switch accountAddress {
		case "tpc1z2xcq07fpf2vehnc9u8cgc28s2q4c7kfl8y63k7",
			"tpc1zkxqqxvgkz4fm8gvefefex936k2t2909f6n0d83",
			"tpc1zs6txk95fww68eaqlx3rsnjhs84y3lmyetml2wt",
			"tpc1zytlvs44sxrhwd9l6jzqqcq85mf4ssdfj2a7dq0":
			continue
		}*/

		balance := amounts[accountIndex]

		log.Printf("[account facts] - %s - balance: %s", accountAddress, balance)

		if balance <= (p.reserveFees + MIN_STAKE) {
			// keep 1 PAC for fee and keep 1 PAC for minimum stake
			continue
		}

		availableTx := balance - p.reserveFees

		wlt, password := p.pipline.GetAccountWallet(accountAddress)

		if wlt == nil {
			return fmt.Errorf("failed to get wallet for address: %s", accountAddress)
		}

		validatorAddresses := make([]string, len(p.validatorAddresses))

		copy(validatorAddresses[:], p.validatorAddresses[:])

		for {
			if len(validatorAddresses) == 0 {
				break
			}

			/*switch validatorAddresses[0] {
			case "tpc1pphac0a0qta6h85y2t45vlj6r6lndyh5szk0et3",
				"tpc1p9jj07hxw74r6eu33ep9uja522s5ml83hvh0tja",
				"tpc1pcl7yd9vtqvnefzwl3q2g7r6f4eky66uuqltw5k",
				"tpc1pjlgtzlk36my0qnzm9krf0jgf6jvs2uafnpqvkg":
				validatorAddresses = validatorAddresses[1:]
				continue
			}*/

			stake := amount.Amount(0)
			var validatorInfo *pactus.ValidatorInfo

			if stakeCache, ok := stakeCacheMap[validatorAddresses[0]]; ok {
				stake = stakeCache.amount
				validatorInfo = stakeCache.validatorInfo
			} else {
				stakeGet, validatorInfoGet, err := p.pipline.GetValidatorStake(validatorAddresses[0])

				if err != nil {
					return err
				}

				stake = stakeGet
				validatorInfo = validatorInfoGet

				stakeCacheMap[validatorAddresses[0]] = &stakeCacheInfo{
					amount:        stake,
					validatorInfo: validatorInfo,
				}
			}

			wants := MAX_STAKE - stake

			log.Printf("[validator facts] validator=%v stake=%v wants=%v", validatorAddresses[0], stake, wants)

			if wants < MIN_STAKE {
				validatorAddresses = validatorAddresses[1:]
				continue
			}

			// stakeAvailable := min(wants, amount)
			stakeAvailable := wants

			if stakeAvailable > availableTx {
				stakeAvailable = availableTx
			}

			after := stakeAvailable + stake

			if after != MAX_STAKE && after > NEAR_MAX_STAKE {
				stakeAvailable = NEAR_MAX_STAKE - stake
				after = stake + stakeAvailable
			}

			if stakeAvailable < MIN_STAKE {
				continue
			}

			fee, err := wlt.CalculateFee(stakeAvailable, payload.TypeBond)

			if err != nil {
				log.Printf("Failed to calculate fee: %v", err)
			}

			log.Printf("[validator bond] validator=%v bond=%v fee=%v after=%v", validatorAddresses[0], stakeAvailable, fee, after)

			opts := []wallet.TxOption{
				wallet.OptionFee(fee),
				//wallet.OptionLockTime(uint32(*lockTime)),
				//wallet.OptionMemo(*memoOpt),
			}

			//addressInfo := wlt.AddressFromPath(validatorAddresses[0])

			//wltAddr, _ := crypto.AddressFromString(wltAddrInfo.Address)

			pub := ""

			if info, ok := p.validatorAddressInfo[validatorAddresses[0]]; ok {
				pub = info.PublicKey
			} else {
				return fmt.Errorf("failed to get public key for address: %s", validatorAddresses[0])
			}

			trx, err := wlt.MakeBondTx(accountAddress, validatorAddresses[0], pub, stakeAvailable, opts...)

			if err != nil {
				return fmt.Errorf("failed to make bond transaction: %w", err)
			}

			err = wlt.SignTransaction(password, trx)

			if err != nil {
				return fmt.Errorf("failed to sign transaction: %w", err)
			}

			bs, _ := trx.Bytes()

			log.Printf("Signed transaction data: %x", bs)

			res, err := wlt.BroadcastTransaction(trx)

			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %w", err)
			}

			broadcastCount++

			log.Printf("Transaction hash: %s", res)

			if validatorInfo == nil && after != MAX_STAKE {
				// last account and last validator, use end of run action wait
				if accountIndex == len(addresses)-1 && len(validatorAddresses) == 1 {
					break
				}

				// create new validator without full stake, wait for next block
				log.Printf("[validator bond] validator=%v wait block confirm", validatorAddresses[0])
				// ensure the transaction is confirmed
				time.Sleep(11 * time.Second)

				stakeGet, validatorInfoGet, err := p.pipline.GetValidatorStake(validatorAddresses[0])

				if err != nil {
					return err
				}

				stakeCacheMap[validatorAddresses[0]].amount = stakeGet
				stakeCacheMap[validatorAddresses[0]].validatorInfo = validatorInfoGet
			} else {
				stakeCacheMap[validatorAddresses[0]].amount = after
			}

			validatorAddresses = validatorAddresses[1:]

			balance -= stakeAvailable + fee
			availableTx = balance - p.reserveFees

			log.Printf("[account update] - %s - balance: %s", accountAddress, balance)

			if balance <= (p.reserveFees + MIN_STAKE) {
				// keep 1 PAC for fee and keep 1 PAC for minimum stake
				break // to next account
			}
		}
	}

	if broadcastCount > 0 {
		log.Printf("wait block confirm")
		// ensure the transaction is confirmed
		time.Sleep(11 * time.Second)
	}

	return nil
}
