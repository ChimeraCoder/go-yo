// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goyo

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/mail"
	"regexp"
	"time"
)

//TODO this needs to be set in an init() function
var ROOT_DIRECTORY = ""

var TIME_REGEX = regexp.MustCompile(`\+([0-9]+)\.([A-Za-z]+)@`)

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
	err = os.Rename(filepath.Join(ROOT_DIRECTORY, "new", filename), filepath.Join(ROOT_DIRECTORY, "cur", filename))

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

	switch strings.ToLower(time_unit) {
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
			time_unit = time.Day
		}

	case "week", "weeks":
		{
			time_unit = 7 * time.Day
		}

	case "month", "months":
		{
			time_unit = 30 * time.Day
		}
	}

	delay := number * time_unit
	future_time := time.Now().Add(number * time_unit)
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
func uniqueFromFilename(filename string) string {
	//TODO actually implement this
	return filename
}
