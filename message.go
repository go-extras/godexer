package executor

func init() {
	RegisterCommand("message", NewMessageCommand)
}

type MessageCommand struct {
	BaseCommand
}

func NewMessageCommand(ectx *ExecutorContext) Command {
	return &MessageCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

func (r *MessageCommand) Execute(variables map[string]any) error {
	return nil
}
