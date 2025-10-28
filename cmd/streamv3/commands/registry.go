package commands

import (
	"context"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/gogstools/gs"
)

// Command represents a subcommand
type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args []string) error
	GetGSCommand() *gs.GSCommand           // Expose gs command for completion (deprecated)
	GetCFCommand() *cf.Command             // Expose completionflags command for completion
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
