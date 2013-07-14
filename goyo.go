// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goyo

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/mail"
	"time"
)

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

	//Assume that there is only one recipient - the one we care about
	addresses, err := message.Header.AddressList("To")
	if err != nil {
		return err
	}

	to_address := addresses[0].Address

	t, err := extractTimeFromAddress(to_address)
	if err != nil {
		return err
	}

	//Schedule future message for that yo-yoed time

	if err := scheduleFutureMessage(filename, t); err != nil {
		return err
	}

	//Move message from /new to /cur, setting Maildir info flag to S (seen)

	//TODO actually implement this

	return nil
}

//Parse an email and return the future time at which to bounce the email
func extractTimeFromAddress(to_address string) (time.Time, error) {
	//TODO actually implement this!
	return time.Now(), nil
}

//scheduleFutureMessage schedules a future email delivery
func scheduleFutureMessage(filename string, t time.Time) (err error) {
	//TODO actually implement this
	uniq := uniqueFromFilename(filename)
	log.Print(uniq)

	return nil
}

//uniqueFromFilename extracts the unique part of a Maildir filename
func uniqueFromFilename(filename string) string {
	//TODO actually implement this
	return filename
}
