// Package cmd is a package which provides utilities that underly the specific command
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/sebasmannem/mailtenance/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultIMAPPort = 1143
	defaultIMAPHost = "localhost"
)

var cmdLister = &cobra.Command{
	Use:   "list",
	Short: "List all labels",
}

type folderListOptions struct {
	Host  string `mapstructure:"host"`
	Port  uint16 `mapstructure:"port"`
	User  string `mapstructure:"user"`
	Token string `mapstructure:"token"`
}

var folderListOpts folderListOptions

var listCommand = &cobra.Command{
	Use:   "folders",
	Run:   listFolders,
	Short: "Retrieve info from imap connection",
}

func init() {
	_, logger := logging.GetLogComponent(context.Background(), logging.CmdComponent)
	listCommand.PersistentFlags().StringVarP(&folderListOpts.User, flagUser, "U", "",
		"username for for connecting to the imap server")
	listCommand.PersistentFlags().StringVarP(&folderListOpts.Token, flagToken, "T", "",
		"token for for connecting to the imap server")
	listCommand.PersistentFlags().Uint16VarP(&folderListOpts.Port, flagPort, "p", defaultIMAPPort,
		"protocol for connecting to the imap server")
	listCommand.PersistentFlags().StringVarP(&folderListOpts.Host, flagHost, "H", defaultIMAPHost,
		"hostname for ip for connecting to the imap server")
	for _, f := range []string{flagUser, flagHost, flagPort, flagToken} {
		if err := viper.BindPFlag(f, listCommand.PersistentFlags().Lookup(f)); err != nil {
			logger.Fatal().AnErr("error", err).Msg("")
		}
	}
	cmdLister.AddCommand(listCommand)
	mainCmd.AddCommand(cmdLister)
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

	for mb := range mailboxes {
		fmt.Println("mailbox:", mb.Name)
	}

	_, err = c.Select("Labels/hcs", false)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("")
	}

	criteria := imap.NewSearchCriteria()
	criteria.Before = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	uids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Aantal te archiveren mails: %d", len(uids))

	/*
		if len(uids) > 0 {
			seqSet := new(imap.SeqSet)
			seqSet.AddNum(uids...)

			// 5. Verplaatsen naar Archief
			if err := c.Move(seqSet, "Archive"); err != nil {
				log.Fatalf("Fout bij verplaatsen: %v", err)
			}
			log.Println("Alle e-mails succesvol gearchiveerd!")
		}
	*/
}
