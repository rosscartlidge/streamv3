package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// execCommand implements the exec command
type execCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newExecCommand())
}

func newExecCommand() *execCommand {
	// For exec, we don't use normal flag parsing because everything after "--" is the command
	// We'll handle parsing manually in Execute()
	cmd := cf.NewCommand("exec").
		Description("Execute command and parse output as records").
		Handler(func(ctx *cf.Context) error {
			// This handler won't be called - we handle everything in Execute()
			return fmt.Errorf("exec command requires manual argument parsing")
		}).
		Build()

	return &execCommand{cmd: cmd}
}

func (c *execCommand) Name() string {
	return "exec"
}

func (c *execCommand) Description() string {
	return "Execute command and parse output as records"
}

func (c *execCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *execCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("exec - Execute command and parse output as records")
		fmt.Println()
		fmt.Println("Usage: streamv3 exec [-generate] -- [command] [args...]")
		fmt.Println()
		fmt.Println("Executes a command and parses its column-aligned output into records.")
		fmt.Println("The first line is treated as the header with column names.")
		fmt.Println("Use '--' to separate streamv3 flags from the command to execute.")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -generate  Generate Go code instead of executing")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  streamv3 exec -- ps -efl")
		fmt.Println("  streamv3 exec -- ps -efl | streamv3 where -match CMD eq bash")
		fmt.Println("  streamv3 exec -- ls -la")
		fmt.Println("  streamv3 exec -generate -- ps -efl | streamv3 generate-go")
		return nil
	}

	// Parse -generate flag and command separately to avoid parsing command flags
	generate := false
	var cmdAndArgs []string

	// Find -- separator
	separatorIdx := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIdx = i
			break
		}
	}

	if separatorIdx == -1 {
		return fmt.Errorf("missing '--' separator before command (usage: streamv3 exec [-generate] -- command args...)")
	}

	// Parse flags before --
	for i := 0; i < separatorIdx; i++ {
		if args[i] == "-generate" {
			generate = true
		} else {
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	// Everything after -- is the command
	if separatorIdx+1 >= len(args) {
		return fmt.Errorf("no command specified after '--'")
	}
	cmdAndArgs = args[separatorIdx+1:]

	command := cmdAndArgs[0]
	cmdArgs := cmdAndArgs[1:]

	// If -generate flag is set, generate Go code instead of executing
	if shouldGenerate(generate) {
		return c.generateCodeDirect(command, cmdArgs)
	}

	// Normal execution: Execute command and parse output
	records, err := streamv3.ExecCommand(command, cmdArgs)
	if err != nil {
		return fmt.Errorf("executing command: %w", err)
	}

	// Write as JSONL to stdout
	if err := lib.WriteJSONL(os.Stdout, records); err != nil {
		return fmt.Errorf("writing JSONL: %w", err)
	}

	return nil
}

// generateCodeDirect generates Go code for the exec command
func (c *execCommand) generateCodeDirect(command string, args []string) error {
	// Generate ExecCommand call with error handling
	var argsStr string
	if len(args) == 0 {
		argsStr = "nil"
	} else {
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			quotedArgs[i] = fmt.Sprintf("%q", arg)
		}
		argsStr = "[]string{" + strings.Join(quotedArgs, ", ") + "}"
	}

	code := fmt.Sprintf(`records, err := streamv3.ExecCommand(%q, %s)
	if err != nil {
		return fmt.Errorf("executing command: %%w", err)
	}`, command, argsStr)

	// Create init fragment (first in pipeline)
	frag := lib.NewInitFragment("records", code, []string{"fmt"})

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
