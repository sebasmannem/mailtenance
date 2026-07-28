// Package cmd is a package which provides utilities that underly the specific command
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/sebasmannem/mailtenance/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var moveCommand = &cobra.Command{
	Use:   "move",
	Short: "move messages",
}

var moveMessagesCommand = &cobra.Command{
	Use:   "messages",
	Run:   moveFolders,
	Short: "move messages to another folder",
}

func init() {
	const (
		day  = time.Hour * 24
		year = day * 365
	)
	_ = moveMessagesCommand.PersistentFlags().StringP(flagUser, "U", "",
		"username for for connecting to the imap server")
	_ = moveMessagesCommand.PersistentFlags().StringP(flagToken, "T", "",
		"token for for connecting to the imap server")
	_ = moveMessagesCommand.PersistentFlags().Uint16P(flagPort, "p", defaultIMAPPort,
		"protocol for connecting to the imap server")
	_ = moveMessagesCommand.PersistentFlags().StringP(flagHost, "H", defaultIMAPHost,
		"hostname for ip for connecting to the imap server")
	_ = moveMessagesCommand.PersistentFlags().StringP(flagFolder, "F", "",
		"folder to move from")
	_ = moveMessagesCommand.PersistentFlags().StringP(flagDestFolder, "d", defaultDestFolder,
		"folder to move to (defaults to"+defaultDestFolder+")")
	_ = moveMessagesCommand.PersistentFlags().DurationP(flagAge, "A", year*10,
		"age of messages (default is 3650 days)")
	moveCommand.AddCommand(moveMessagesCommand)
	mainCmd.AddCommand(moveCommand)
}

func moveFolders(_ *cobra.Command, _ []string) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	_, logger := logging.GetLogComponent(ctx, logging.CmdComponent)

	folder := viper.GetString(flagFolder)
	if folder == "" {
		logger.Panic().Msg("you must specify a folder to move from")
	}

	host := fmt.Sprintf("%s:%d", viper.GetString(flagHost), viper.GetInt(flagPort))
	logger.Debug().Str("host", host).Msg("")
	c, err := client.Dial(host)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}
	defer c.Logout()

	// TODO: implement tls
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	if err := c.StartTLS(tlsConfig); err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}

	logger.Debug().Str("user", viper.GetString(flagUser)).Msg("")
	if err := c.Login(viper.GetString(flagUser), viper.GetString(flagToken)); err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}

	_, err = c.Select(folder, false)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}

	criteria := imap.NewSearchCriteria()
	criteria.Before = time.Now().Add(-viper.GetDuration(flagAge))

	uids, err := c.Search(criteria)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}
	if len(uids) > 0 {
		var seqSet imap.SeqSet
		seqSet.AddNum(uids...)
		c.Move(&seqSet, viper.GetString(flagDestFolder))
	}
}
