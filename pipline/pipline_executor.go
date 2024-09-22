package pipline

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/frimin/pactus-staker/config"
	"github.com/frimin/pactus-staker/pipline/action"
)

type PiplineExecutor interface {
	Run() error
}

type piplineExecutor struct {
	piplines []*pipline
	retry    []int
}

type pendingAction struct {
	piplineIndex int
	actionIndex  int
	pipline      *pipline
	action       action.Action
	triggerTime  time.Time
}

func CreateExecutor(config *config.Config) (PiplineExecutor, error) {
	pipExecutor := &piplineExecutor{
		piplines: []*pipline{},
		retry:    make([]int, len(config.Options.RetryDelay)),
	}

	if len(config.Options.RetryDelay) == 0 {
		log.Fatalf("no retry delay found")
	}

	copy(pipExecutor.retry[:], config.Options.RetryDelay[:])

	for _, p := range config.Pipeline {
		p, err := createPipline(config.Options, p)

		if err != nil {
			return nil, fmt.Errorf("error creating pipline: %w", err)
		}

		pipExecutor.piplines = append(pipExecutor.piplines, p)
	}

	if len(pipExecutor.piplines) == 0 {
		return nil, fmt.Errorf("no piplines found")
	}

	return pipExecutor, nil
}

func getTimeFromHHMM(now time.Time, hhmm string) (time.Time, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}

	return today.Add(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute), nil
}

func (p *piplineExecutor) makePendingActions(now time.Time) []*pendingAction {
	actions := []*pendingAction{}

	for piplineIndex, pipline := range p.piplines {
		for actionIndex, action := range pipline.actions {
			for _, triggerTime := range action.GetTime() {
				t, err := getTimeFromHHMM(now, triggerTime)

				if err != nil {
					log.Fatalf("error parsing time: %v", err)
				}

				if t.After(now) {
					actions = append(actions, &pendingAction{
						piplineIndex: piplineIndex,
						actionIndex:  actionIndex,
						pipline:      pipline,
						action:       action,
						triggerTime:  t,
					})
				}
			}
		}
	}

	sort.Slice(actions, func(i, j int) bool {
		if !actions[i].triggerTime.Equal(actions[j].triggerTime) {
			return actions[i].triggerTime.Before(actions[j].triggerTime)
		}

		if actions[i].piplineIndex != actions[j].piplineIndex {
			return actions[i].piplineIndex < actions[j].piplineIndex
		}

		return actions[i].actionIndex < actions[j].actionIndex
	})

	return actions
}

func (p *piplineExecutor) GetNextActions(lastTime time.Time) []*pendingAction {
	actions := p.makePendingActions(lastTime)

	if len(actions) == 0 {
		// is no action pending, try get next day
		nextDay := lastTime.AddDate(0, 0, 1)
		nextDayStart := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())
		actions = p.makePendingActions(nextDayStart)

		if len(actions) == 0 {
			log.Fatalf("no actions found")
		}
	}

	for _, action := range actions {
		log.Printf("[pipline %d %s action %d %s] Waiting at %s", action.piplineIndex, action.pipline.name, action.actionIndex, action.action.GetName(), action.triggerTime)
	}

	return actions
}

func (p *piplineExecutor) Run() error {
	pendingActions := p.GetNextActions(time.Now())

	for {
		if len(pendingActions) == 0 {
			pendingActions = p.GetNextActions(time.Now())
		}

		if time.Now().After(pendingActions[0].triggerTime) {
			action := pendingActions[0]

			retry := make([]int, len(p.retry))

			copy(retry[:], p.retry[:])

			log.Printf("[pipline %d %s action %d %s] Running at %s", action.piplineIndex, action.pipline.name, action.actionIndex, action.action.GetName(), action.triggerTime)

			err := action.action.Run()

			for {
				if err != nil {
					if len(retry) > 0 {
						log.Printf("Error running action: %v, retry later ...", err)
					} else {
						log.Printf("Error running action: %v, no retry left", err)
					}
				} else {
					break
				}

				if len(retry) > 0 {
					time.Sleep(time.Duration(retry[0]) * time.Second)
					retry = retry[1:]
					err = action.action.Run()
				} else {
					break
				}
			}

			if err == nil {
				log.Printf("[pipline %d %s action %d %s] done", action.piplineIndex, action.pipline.name, action.actionIndex, action.action.GetName())
			} else {
				log.Printf("[pipline %d %s action %d %s] failed: %v", action.piplineIndex, action.pipline.name, action.actionIndex, action.action.GetName(), err)
			}

			pendingActions = pendingActions[1:]

			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}
