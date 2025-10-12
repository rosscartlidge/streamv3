package commands

import "context"

// Command represents a subcommand
type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args []string) error
}

var commands []Command

// RegisterCommand registers a command for use
func RegisterCommand(cmd Command) {
	commands = append(commands, cmd)
}

// GetCommands returns all registered commands
func GetCommands() []Command {
	return commands
}
