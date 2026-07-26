package main

import (
	"crypto/tls"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	c, err := client.Dial("127.0.0.1:1143")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	if err := c.StartTLS(tlsConfig); err != nil {
		log.Fatal(err)
	}

	user := os.Getenv("PROTONUSER")
	token := os.Getenv("PROTONTOKEN")
	if err := c.Login(user+"@proton.me", token); err != nil {
		log.Fatal(err)
	}

	_, err = c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
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
