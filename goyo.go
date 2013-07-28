// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goyo

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//WARNING: All of these flags are unstable and currently subject to change
var ROOT_DIRECTORY = flag.String("rootdir", "", "root directory for mail")
var EMAIL_ADDRESS = flag.String("email", "", "email address")
var EMAIL_PASSWORD = flag.String("password", "", "email password")
var CONFIGURED_EMAIL = flag.String("configuredemail", "", "configured email")

var TIME_REGEX = regexp.MustCompile(`\+([0-9]+)\.([A-Za-z]+)@`)

var UNIQ_FILENAME_REGEX = regexp.MustCompile(`(.+):`)

func init() {
	flag.Parse()
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

	//Assume that there is only one recipient - the one we care about
	addresses, err := message.Header.AddressList("To")
	if err != nil {
		return err
	}

	to_address := addresses[0].Address
	log.Printf("Found address %s for message %s", to_address, message_id)

	t, err := extractTimeFromAddress(to_address)
	if err != nil {
		return err
	}

	//Schedule future message for that yo-yoed time

	log.Printf("Scheduling message for %v", t)
	if err := scheduleFutureMessage(filename, t); err != nil {
		return err
	}

	//Move message from /new to /cur, setting Maildir info flag to S (seen)
	destination := filepath.Join(*ROOT_DIRECTORY, "cur", strings.TrimPrefix(uniqueFromFilename(filename)+":2,S", filepath.Join(*ROOT_DIRECTORY, "new")))
	log.Printf("Moving message from %s to %s", filename, destination)
	err = os.Rename(filename, destination)

	return err
}

//Parse an email address and return the future time at which to bounce the email
func extractTimeFromAddress(to_address string) (time.Time, error) {

	matches := TIME_REGEX.FindStringSubmatch(to_address)

	number_s := matches[1]
	time_unit_s := matches[2]

	number, err := strconv.Atoi(number_s)
	if err != nil {
		panic(err)
	}

	//For now, we'll support minutes, hours, days, weeks, and months

	var time_unit time.Duration

	switch strings.ToLower(time_unit_s) {
	case "minute", "minutes":
		{
			time_unit = time.Minute
		}

	case "hour", "hours":
		{
			time_unit = time.Hour
		}

	case "day", "days":
		{
			time_unit = 24 * time.Hour
		}

	case "week", "weeks":
		{
			time_unit = 7 * 24 * time.Hour
		}

	case "month", "months":
		{
			time_unit = 30 * 7 * 24 * time.Hour
		}
	}

	delay := time.Duration(number) * time_unit
	//TODO use the time the message was sent instead of time.Now
	future_time := time.Now().Add(delay)
	return future_time, nil

}

//scheduleFutureMessage schedules a future email delivery
func scheduleFutureMessage(filename string, t time.Time) (err error) {
	//TODO actually implement this
	uniq := uniqueFromFilename(filename)
	log.Print(uniq)

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

func sendMail(recipient_email string) {
	auth := smtp.PlainAuth(
		"",
		*EMAIL_ADDRESS,
		*EMAIL_PASSWORD,
		"smtp.gmail.com", //TODO abstract this beyond Google/Gmail
	)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		"smtp.gmail.com:25",
		auth,
		*EMAIL_ADDRESS,
		[]string{recipient_email},

		//TODO use proper Go templating for this
		[]byte(`To: kev23819@gmail.com
Subject: Test email

This is the body of the reminder email.`),
	)

	if err != nil {
		log.Fatal(err)
	}

}
