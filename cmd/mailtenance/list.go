// Package cmd is a package which provides utilities that underly the specific command
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"slices"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/sebasmannem/mailtenance/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List all folders",
}

var listFoldersCommand = &cobra.Command{
	Use:   "folders",
	Run:   listFolders,
	Short: "List folders and the number of message older than a certain age",
}

var listMessagesCommand = &cobra.Command{
	Use:   "messages",
	Run:   listMessages,
	Short: "List number of messages for a specific folder",
}

func init() {
	const (
		day  = time.Hour * 24
		year = day * 365
	)
	_ = listFoldersCommand.PersistentFlags().StringP(flagUser, "U", "",
		"username for for connecting to the imap server")
	_ = listFoldersCommand.PersistentFlags().StringP(flagToken, "T", "",
		"token for for connecting to the imap server")
	_ = listFoldersCommand.PersistentFlags().Uint16P(flagPort, "p", defaultIMAPPort,
		"protocol for connecting to the imap server")
	_ = listFoldersCommand.PersistentFlags().StringP(flagHost, "H", defaultIMAPHost,
		"hostname for ip for connecting to the imap server")
	_ = listFoldersCommand.PersistentFlags().DurationP(flagAge, "A", year*10,
		"age of mails (default is 3650 days)")
	_ = listFoldersCommand.PersistentFlags().StringP(flagFolder, "F", "",
		"folder to list")

	_ = listMessagesCommand.PersistentFlags().StringP(flagUser, "U", "",
		"username for for connecting to the imap server")
	_ = listMessagesCommand.PersistentFlags().StringP(flagToken, "T", "",
		"token for for connecting to the imap server")
	_ = listMessagesCommand.PersistentFlags().Uint16P(flagPort, "p", defaultIMAPPort,
		"protocol for connecting to the imap server")
	_ = listMessagesCommand.PersistentFlags().StringP(flagHost, "H", defaultIMAPHost,
		"hostname for ip for connecting to the imap server")
	_ = listMessagesCommand.PersistentFlags().DurationP(flagAge, "A", year*10,
		"age of mails (default is 3650 days)")
	_ = listMessagesCommand.PersistentFlags().StringP(flagFolder, "F", "",
		"folder to list")

	listCommand.AddCommand(listFoldersCommand)
	listCommand.AddCommand(listMessagesCommand)
	mainCmd.AddCommand(listCommand)
}

func listFolders(_ *cobra.Command, _ []string) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	_, logger := logging.GetLogComponent(ctx, logging.CmdComponent)
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

	mailboxes := make(chan *imap.MailboxInfo, 100)
	err = c.List("", "*", mailboxes)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}

	type box struct {
		name  string
		count int
	}
	var boxes []box
	folder := viper.GetString(flagFolder)
	for mb := range mailboxes {
		if folder != "" && mb.Name != folder {
			continue
		}
		_, err := c.Select(mb.Name, false)
		if err != nil {
			logger.Fatal().AnErr("error", err).Msg("")
		}

		criteria := imap.NewSearchCriteria()
		criteria.Before = time.Now().Add(-viper.GetDuration(flagAge))

		uids, err := c.Search(criteria)
		if err != nil {
			logger.Fatal().AnErr("error", err).Msg("")
		}
		boxes = append(boxes, box{name: mb.Name, count: len(uids)})
	}
	slices.SortFunc(boxes, func(a, b box) int {
		return b.count - a.count
	})
	for _, b := range boxes {
		fmt.Printf("mailbox %s (%d messages)\n", b.name, b.count)
	}
}

func listMessages(_ *cobra.Command, _ []string) {
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

	mailboxes := make(chan *imap.MailboxInfo, 100)
	err = c.List("", "*", mailboxes)
	if err != nil {
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
	fmt.Printf("mailbox %s (%d messages)\n", folder, len(uids))
}
