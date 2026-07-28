package main

import (
	"context"

	"github.com/sebasmannem/mailtenance/internal/logging"
	"github.com/sebasmannem/mailtenance/internal/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagDebug = "debug"
	flagPort  = "port"
	flagHost  = "host"
	flagUser  = "user"
	flagToken = "token"
)

type mainOptions struct {
	Debug bool `mapstructure:"debug"`
}

var mainOpts mainOptions

// mainCmd defines a cobra command to execute when running mailtenance
var mainCmd = &cobra.Command{
	Use:     "mailtenance",
	Short:   "mail maintenance tool",
	Version: version.GetAppVersion(),
	PersistentPreRun: func(c *cobra.Command, _ []string) {
		if mainOpts.Debug {
			logging.SetStaticLevel("debug")
		}
		_, logger := logging.GetLogComponent(c.Context(), logging.CmdComponent)
		logger.Debug().Bool("debug", mainOpts.Debug).Msg("debugging enabled")

	},
	Run: func(c *cobra.Command, _ []string) { _ = c.Help() },
}

func init() {
	_, logger := logging.GetLogComponent(context.Background(), logging.CmdComponent)
	mainCmd.PersistentFlags().BoolVarP(&mainOpts.Debug, flagDebug, "d", false,
		"username for for connecting to the imap server")
	if err := viper.BindPFlag(flagDebug, mainCmd.PersistentFlags().Lookup(flagDebug)); err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}
	viper.SetEnvPrefix("mailt")
	viper.AutomaticEnv()
}

var cmdVersion = &cobra.Command{
	Use: "version",
	Run: func(_ *cobra.Command, _ []string) {
		_, logger := logging.GetLogComponent(context.Background(), logging.CmdComponent)
		logger.Info().Str("version", version.GetAppVersion()).Msg("")
	},
	Short: "Display the version",
}

func init() {
	mainCmd.AddCommand(cmdVersion)
}

func main() {
	_, logger := logging.GetLogComponent(context.Background(), logging.CmdComponent)
	if err := mainCmd.Execute(); err != nil {
		logger.Fatal().AnErr("err", err).Msg("")
	}
}
