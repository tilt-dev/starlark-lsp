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

type rootCmd struct {
	*cobra.Command
	debug   bool
	verbose bool
}

func newRootCmd() *rootCmd {
	cmd := rootCmd{
		Command: &cobra.Command{
			Use:   "starlark-lsp",
			Short: "Language server for Starlark",
		},
	}

	cmd.PersistentFlags().BoolVar(&cmd.debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().BoolVar(&cmd.verbose, "verbose", false, "Enable verbose logging")

	return &cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(
		newStartCmd().Command,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSignalHandler(cancel)

	level := zap.NewAtomicLevelAt(zapcore.WarnLevel)
	logger := mustInitializeLogger(level)
	defer func() {
		// use defer rather than PersistentPostRun to ensure execution on panic
		_ = logger.Sync()
	}()
	ctx = protocol.WithLogger(ctx, logger)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if rootCmd.debug {
			level.SetLevel(zapcore.DebugLevel)
		} else if rootCmd.verbose {
			level.SetLevel(zapcore.InfoLevel)
		}
	}

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		if !isCobraError(err) {
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
