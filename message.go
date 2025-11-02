package executor

//nolint:gochecknoinits // init is used for automatic command registration
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

func (*MessageCommand) Execute(_ map[string]any) error {
	return nil
}
