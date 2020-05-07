package replay

import (
	"errors"
	"time"
)

// stringToTime tries to parse a given string representation of a dateTime.
//
// This function returns early if there *was no error*, which contrasts the
// standard golang pattern of returning early if there *is an error*.
// That is done because we essentially only care if *all* of the time parsing
// attempts fail, rather than if only one of them fails.
//
// This function works by trying to parse the most specific time value that
// we know of (RFC3339Nano, which has nanoseconds and timezones) and
// then gradually testing less specific times. This produces a `stringToTime`
// function that can be used anywhere in our package without data loss,
// so long as you use any of the 6 (!!!) time formats we try and parse.
func stringToTime(dateTime string) (output time.Time, err error) {

	if dateTime == "" {
		err = errors.New("the dateTime argument was empty")
		return time.Time{}, err
	}

	// this defines the "layout" of our dateTime argument
	// see the following docs for details
	// - https://golang.org/pkg/time/#Parse
	// - https://golang.org/pkg/time/#pkg-constants
	var layout string

	// I'm on the fence about DRYing these next few blocks.
	// They're low complexity, so the copy pasting is fine IMO.

	// try with the 1st format, what we wish we had, RFC3339Nano
	//
	// format attributes
	// 	timezone: true
	// 	nanoseconds: true
	// 	seconds: true
	layout = time.RFC3339Nano
	output, err = time.Parse(layout, dateTime)
	if err == nil {
		return output, nil
	}

	// try again! this custom format is RFC3339Nano without the timezome
	//
	// format attributes
	// 	timezone: false
	// 	nanoseconds: true
	// 	seconds: true
	layout = "2006-01-02T15:04:05.999999999"
	output, err = time.Parse(layout, dateTime)
	if err == nil {
		return output, nil
	}

	// try again! this is RFC3339Nano without the nanoseconds
	//
	// format attributes
	// 	timezone: true
	// 	nanoseconds: false
	// 	seconds: true
	layout = time.RFC3339
	output, err = time.Parse(layout, dateTime)
	if err == nil {
		return output, nil
	}

	// try again! this custom format is RFC3339 without the timezone
	//
	// format attributes
	// 	timezone: false
	// 	nanoseconds: false
	// 	seconds: true
	layout = "2006-01-02T15:04:05"
	output, err = time.Parse(layout, dateTime)
	if err == nil {
		return output, nil
	}

	// try again! this custom format is RFC3339 without the seconds
	//
	// format attributes
	// 	timezone: true
	// 	nanoseconds: false
	// 	seconds: false
	layout = "2006-01-02T15:04:05Z07:00"
	output, err = time.Parse(layout, dateTime)
	if err == nil {
		return output, nil
	}

	// last try! this custom format is RFC3339 without the seconds or timezone
	//
	// format attributes
	// 	timezone: false
	// 	nanoseconds: false
	// 	seconds: false
	layout = "2006-01-02T15:04"
	output, err = time.Parse(layout, dateTime)
	if err == nil {
		return output, nil
	}

	return output, err
}
