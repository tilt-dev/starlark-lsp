package cli

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RootCmd struct {
	*cobra.Command
	debug   bool
	verbose bool
}

func NewRootCmd() *RootCmd {
	cmd := RootCmd{
		Command: &cobra.Command{
			Use:   "starlark-lsp",
			Short: "Language server for Starlark",
		},
	}

	cmd.PersistentFlags().BoolVar(&cmd.debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().BoolVar(&cmd.verbose, "verbose", false, "Enable verbose logging")

	cmd.AddCommand(newStartCmd().Command)

	return &cmd
}

func (c *RootCmd) Logger() (logger *zap.Logger, cleanup func()) {
	level := zap.NewAtomicLevelAt(zapcore.WarnLevel)
	logger = mustInitializeLogger(level)
	if c.debug {
		level.SetLevel(zapcore.DebugLevel)
	} else if c.verbose {
		level.SetLevel(zapcore.InfoLevel)
	}

	cleanup = func() {
		// use defer rather than PersistentPostRun to ensure execution on panic
		_ = logger.Sync()
	}

	return logger, cleanup
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func (c *RootCmd) Execute(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	setupSignalHandler(cancel)

	return c.ExecuteContext(ctx)
}

func Execute() {
	c := NewRootCmd()
	logger, cleanup := c.Logger()
	defer cleanup()

	ctx := protocol.WithLogger(context.Background(), logger)

	err := c.Execute(ctx)
	if err != nil {
		if !isCobraError(err) {
			logger := protocol.LoggerFromContext(ctx)
			logger.Error("fatal error", zap.Error(err))
		}
		os.Exit(1)
	}
}

func setupSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				// TODO(milas): give open conns a grace period to close gracefully
				cancel()
				os.Exit(0)
			}
		}
	}()
}

func isCobraError(err error) bool {
	// Cobra doesn't give us a good way to distinguish between Cobra errors
	// (e.g. invalid command/args) and app errors, so ignore them manually
	// to avoid logging out scary stack traces for benign invocation issues
	return strings.Contains(err.Error(), "unknown flag")
}
