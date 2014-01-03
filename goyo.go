// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style // license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const _CHECK_INTERVAL = 120 * time.Second

//WARNING: All of these flags are unstable and currently subject to change
var ROOT_DIRECTORY = flag.String("rootdir", "", "root directory for mail")
var EMAIL_ADDRESS = flag.String("email", "", "email address")
var EMAIL_PASSWORD = flag.String("password", "", "email password")
var CONFIGURED_EMAIL = flag.String("configuredemail", "", "configured email")

var TIME_REGEX = regexp.MustCompile(`\+([0-9A-Za-z\.]+)@`)

var UNIQ_FILENAME_REGEX = regexp.MustCompile(`(.+):`)

func init() {
	flag.Parse()
}

type Response struct {
	Recipient string
	Subject   string
	Body      string
}

//processMessage processes each new message that appears in /new
func processMessage(filename string) error {
	//Parse message and determine when the message should be yo-yoed

	bts, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	message, err := mail.ReadMessage(bytes.NewBuffer(bts))

	if err != nil {
		return err
	}

	message_id := message.Header.Get("Message-ID")
	subject := message.Header.Get("Subject")
	from_address := message.Header.Get("Return-Path")

	//Assume that valid messages have only one recipient - the one we care about
	addresses, err := message.Header.AddressList("To")
	if err != nil {
		return err
	}

	to_address := addresses[0].Address
	log.Printf("Found address %s for message %s", to_address, message_id)

	//Only allow emails sent from the configured email
	//TODO do something less hacky to allow angled brackets
	if from_address != *CONFIGURED_EMAIL && from_address != "<"+*CONFIGURED_EMAIL+">" {
		log.Printf("Skipping email sent from %s", from_address)
		return nil
	}

	//Determine what time reminder message should be sent
	t, err := extractTimeFromAddress(to_address)
	if err != nil {
		log.Printf("Error extracting time from %v: %v", to_address, err)
		return err
	}

	//Schedule future message for that yo-yoed time

	log.Printf("Scheduling message for %v", t)
	if err := scheduleFutureMessage(*CONFIGURED_EMAIL, message_id, subject, t); err != nil {
		log.Printf("Error scheduling future message %v", err)
		return err
	}

	//Move message from /new to /cur, setting Maildir info flag to S (seen)
	destination := filepath.Join(*ROOT_DIRECTORY, "cur", strings.TrimPrefix(uniqueFromFilename(filename)+":2,S", filepath.Join(*ROOT_DIRECTORY, "new")))
	log.Printf("Moving message from %s to %s", filename, destination)
	err = os.Rename(filename, destination)

	return err
}

//Parse an email address and return the future time at which to bounce the email
func extractTimeFromAddress(to_address string) (future_time time.Time, err error) {

	matches := TIME_REGEX.FindStringSubmatch(to_address)

	if len(matches) != 2 {
		err = fmt.Errorf("Could not extract time from email address %s", to_address)
		return
	}

	delay, err := time.ParseDuration(matches[1])
	if err != nil {
		return
	}

	//TODO use the time the message was sent instead of time.Now
	future_time = time.Now().Add(delay)
	return

}

//scheduleFutureMessage schedules a future email delivery
func scheduleFutureMessage(recipient_email, message_id, original_subject string, t time.Time) (err error) {

	//TODO store a journal of jobs in a persistent database for logging/auditing/etc.
	time_to_sleep := t.Sub(time.Now())
	go func(recipient_email, message_id, original_subject string, d time.Duration) {
		log.Printf("Sending email %s to %s in %v from now", original_subject, recipient_email, d)
		time.Sleep(d)
		sendMail(recipient_email, message_id, original_subject)
	}(recipient_email, message_id, original_subject, time_to_sleep)
	return nil
}

//uniqueFromFilename extracts the unique part of a Maildir filename
func uniqueFromFilename(filename string) (uniq string) {
	//The real input set may actually be larger/more complicated than this
	//But this works for now
	matches := UNIQ_FILENAME_REGEX.FindStringSubmatch(filename)
	uniq = matches[1]
	return
}

//sendMail sends a reply email, given the original Message-ID header and original Subject header
//This will allow clients which support threading to thread conversations properly
func sendMail(recipient_email, message_id, original_subject string) error {
	tmpl, err := template.New("email_response").Parse(`To: {{.Recipient}}
Subject: {{.Subject}}

{{.Body}}`)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}
	recipient_address, err := mail.ParseAddress(recipient_email)
	if err != nil {
		return err
	}
	r := Response{recipient_email, "Re: " + original_subject, "Test body"}
	err = tmpl.Execute(b, r)

	auth := smtp.PlainAuth(
		"",
		*EMAIL_ADDRESS,
		*EMAIL_PASSWORD,
		"smtp.gmail.com", //TODO abstract this beyond Google/Gmail
	)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	log.Printf("Sending email %v", r)
	err = smtp.SendMail(
		"smtp.gmail.com:25",
		auth,
		*EMAIL_ADDRESS,
		[]string{recipient_address.Address},

		//TODO use proper Go templating for this
		b.Bytes(),
	)

	if err != nil {
		return err
	}
	return nil
}

//monitorBox will check periodically (every 2 minutes?) for new messages that need to be scheduled, and schedule them if present
func monitorBox() {

	for {
		//TODO abstract this to use any root directory for the box
		files, err := ioutil.ReadDir(filepath.Join(*ROOT_DIRECTORY, "new"))
		if err != nil {
			log.Printf("error reading directory: %v", err)
		}

		for _, file := range files {
			err := processMessage(filepath.Join(*ROOT_DIRECTORY, "new", file.Name()))
			if err != nil {
				log.Printf("ERROR processing message %v", err)
			}
		}
		log.Printf("Sleeping for %s seconds", _CHECK_INTERVAL)
		time.Sleep(_CHECK_INTERVAL)
	}
}

func main() {
	monitorBox()
}
