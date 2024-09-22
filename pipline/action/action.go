package action

import (
	"fmt"

	"github.com/frimin/pactus-staker/config"
	"github.com/frimin/pactus-staker/pipline/action/bond"
	"github.com/frimin/pactus-staker/pipline/provider"
)

type Action interface {
	Run() error
	GetTime() []string
	GetName() string
}

func CreateAction(pipline provider.PiplineProvider, index int, optionsConfig *config.Options, actionConfig *config.Action) (Action, error) {
	switch actionConfig.Type {
	case "bond":
		rewardAction, err := bond.CreateBondAction(pipline, index, optionsConfig, actionConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating bond action: %w", err)
		}
		return rewardAction, nil
	default:
		return nil, fmt.Errorf("unknown action type: %s", actionConfig.Type)
	}
}
