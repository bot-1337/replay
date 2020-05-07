package replay

import (
	"testing"
)

func TestStringToTime(t *testing.T) {
	tdata := []struct {
		testCase        string
		input           string
		expectedAnError bool
	}{
		{
			testCase:        "empty",
			expectedAnError: true,
		},
		{
			testCase:        "not_a_time",
			input:           "THIS IS NOT A TIME",
			expectedAnError: true,
		},
		// Returns an error as the layout is not a valid time value
		{
			testCase:        "example_from__golang.org/pkg/time/#pkg-constants",
			input:           "2006-01-02T15:04:05.999999999Z07:00",
			expectedAnError: true,
		},
		{
			testCase:        "example_from__audit-prompt.md__changeTime",
			input:           "2016-01-01T03:24:30.001180",
			expectedAnError: false,
		},
		{
			testCase:        "example_from__audit-prompt.md__./replay__input",
			input:           "2016-01-01T03:00",
			expectedAnError: false,
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// logic under test
			// TODO: test the output also
			_, err := stringToTime(test.input)

			// assertions
			if test.expectedAnError && err == nil {
				t.Error("expected an error, but there was none!")
			}
			if !test.expectedAnError && err != nil {
				t.Error(err)
			}
		})
	}
}
