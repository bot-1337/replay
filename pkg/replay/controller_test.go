package replay

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestGetState(t *testing.T) {
	tdata := []struct {
		testCase        string
		input           getStateInput
		expectedOutput  getStateOutput
		expectedAnError bool // TODO: check for specific types of errors
	}{
		{
			testCase:        "empty",
			expectedAnError: true,
		},
		{
			testCase: "reader_returned_not_found",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return "", false, nil
				},
			},
			expectedAnError: true,
		},
		{
			testCase: "reader_raised_an_error",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return "", true, errors.New("some error here")
				},
			},
			expectedAnError: true,
		},
		{
			testCase: "simple_case__one_line__after",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `{"changeTime": "9999-01-01T00:43:00.001064", "before": {"ambientTemp": 80.0}}`, true, nil
				},
			},
			expectedOutput: getStateOutput{
				State: map[string]interface{}{
					"ambientTemp": 80.0,
				},
			},
		},
		{
			testCase: "floats_dont_truncate",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `{"changeTime": "9999-01-01T00:43:00.001064", "before": {"ambientTemp": 80.888}}`, true, nil
				},
			},
			expectedOutput: getStateOutput{
				State: map[string]interface{}{
					"ambientTemp": 80.888,
				},
			},
		},
		{
			testCase: "bad_input_field",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"BAD INPUT FIELD"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `{"changeTime": "9999-01-01T00:43:00.001064", "before": {"ambientTemp": 80.0}}`, true, nil
				},
			},
			expectedAnError: true,
		},
		{
			testCase: "simple_case__one_line__before",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `{"changeTime": "1111-01-01T00:43:00.001064", "after": {"ambientTemp": 80.0}}`, true, nil
				},
			},
			expectedOutput: getStateOutput{
				State: map[string]interface{}{
					"ambientTemp": 80.0,
				},
			},
		},
		{
			testCase: "simple_case__two_lines__both_after",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `
						{"changeTime": "1111-01-01T00:43:00.001064", "after": {"ambientTemp": 11.0}}
						{"changeTime": "9999-01-01T00:43:00.001064", "after": {"ambientTemp": 99.0}}
					`, true, nil
				},
			},
			expectedOutput: getStateOutput{
				State: map[string]interface{}{
					"ambientTemp": 11.0,
				},
			},
		},
		{
			testCase: "simple_case__two_lines__both_before",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `
						{"changeTime": "1111-01-01T00:43:00.001064", "before": {"ambientTemp": 11.0}}
						{"changeTime": "9999-01-01T00:43:00.001064", "before": {"ambientTemp": 99.0}}
					`, true, nil
				},
			},
			expectedOutput: getStateOutput{
				State: map[string]interface{}{
					"ambientTemp": 99.0,
				},
			},
		},
		{
			testCase: "data_error__before_and_after_mismatched",
			input: getStateInput{
				dateTime: "2016-01-01T00:43",
				fields:   []string{"ambientTemp"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `
						{"changeTime": "1111-01-01T00:43:00.001064", "before": {"ambientTemp": 11.0}, "after": {"ambientTemp": 11.0}}
						{"changeTime": "9999-01-01T00:43:00.001064", "before": {"ambientTemp": 99.0}, "after": {"ambientTemp": 99.0}}
					`, true, nil
				},
			},
			expectedAnError: true,
		},
		{
			testCase: "case_from_example_prompt",
			input: getStateInput{
				dateTime: "2016-01-01T03:00",
				fields:   []string{"ambientTemp", "schedule"},
				readerFunc: func(path string) (output string, found bool, err error) {
					return `
						{"changeTime": "2016-01-01T00:30:00.001059", "after": {"ambientTemp": 79.0}, "before": {"ambientTemp": 77.0}}
						{"changeTime": "2016-01-01T00:43:00.001064", "after": {"ambientTemp": 80.0}, "before": {"ambientTemp": 79.0}}
						{"changeTime": "2016-01-01T01:32:00.009816", "after": {"ambientTemp": 81.0}, "before": {"ambientTemp": 80.0}}
						{"changeTime": "2016-01-01T01:38:00.001038", "after": {"ambientTemp": 82.0}, "before": {"ambientTemp": 81.0}}
						{"changeTime": "2016-01-01T01:44:00.001145", "after": {"ambientTemp": 81.0}, "before": {"ambientTemp": 82.0}}
						{"changeTime": "2016-01-01T02:08:30.010956", "after": {"ambientTemp": 79.0}, "before": {"ambientTemp": 81.0}}
						{"changeTime": "2016-01-01T02:47:30.002413", "after": {"ambientTemp": 77.0}, "before": {"ambientTemp": 79.0}}
						{"changeTime": "2016-01-01T03:02:30.001424", "after": {"ambientTemp": 78.0}, "before": {"ambientTemp": 77.0}}
						{"changeTime": "2016-01-01T03:08:00.007712", "after": {"ambientTemp": 80.0}, "before": {"ambientTemp": 78.0}}
						{"changeTime": "2016-01-01T03:12:30.008936", "after": {"ambientTemp": 79.0}, "before": {"ambientTemp": 80.0}}
						{"changeTime": "2016-01-01T03:18:30.001950", "after": {"schedule": true}, "before": {"schedule": false}}
						{"changeTime": "2016-01-01T03:24:30.001180", "after": {"setpoint": {"heatTemp": 67.0}}, "before": {"setpoint": {"heatTemp": 69.0}}}
					`, true, nil
				},
			},
			expectedOutput: getStateOutput{
				State: map[string]interface{}{
					"ambientTemp": 77.0,
					"schedule":    false,
				},
			},
		},
	}
	for _, test := range tdata {
		t.Run(test.testCase, func(t *testing.T) {
			// logic under test
			output, err := getState(test.input)

			// assertions
			if !reflect.DeepEqual(test.expectedOutput, output) {
				t.Errorf("expected %+v to equal %+v", test.expectedOutput, output)
			}
			if test.expectedAnError && err == nil {
				t.Error("expected an error, but there was none!")
			}
			if !test.expectedAnError && err != nil {
				t.Error(err)
			}
		})
	}
}
