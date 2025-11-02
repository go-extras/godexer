package executor

import (
	"time"
)

func init() {
	RegisterCommand("sleep", NewSleepCommand)
}

func NewSleepCommand(ectx *ExecutorContext) Command {
	return &SleepCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

type SleepCommand struct {
	BaseCommand
	Seconds int
}

func (s *SleepCommand) Execute(variables map[string]any) error {
	s.Ectx.Logger.Infof("Sleeping for %d seconds", s.Seconds)
	TimeSleep(time.Duration(s.Seconds) * time.Second)
	return nil
}
