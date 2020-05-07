package replay

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type getStateInput struct {
	fields     []string
	dataSource string
	dateTime   string
	readerFunc readerFunc
}

type getStateOutput struct {
	State map[string]interface{} `json:"state"`
	Ts    string                 `json:"ts"`
}

type fieldData struct {
	value interface{}
	time  time.Time
}

type fileLineJSON struct {
	After      map[string]interface{} `json:"after"`
	Before     map[string]interface{} `json:"before"`
	ChangeTime string                 `json:"changeTime"`
}

func getState(input getStateInput) (output getStateOutput, err error) {
	// parse dateTime input
	inputDateTime, err := stringToTime(input.dateTime)
	if err != nil {
		err = fmt.Errorf("error parsing dateTime: %w", err)
		return getStateOutput{}, err
	}
	inputYear, inputMonth, inputDay := inputDateTime.Date()

	// construct path for reader
	// paths look like so => /tmp/ehub_data/2016/01/01.jsonl.gz
	path := fmt.Sprintf(`%s/%d/%02d/%02d.jsonl.gz`, input.dataSource, inputYear, inputMonth, inputDay)

	// get reader data
	fileData, found, err := input.readerFunc(path)
	if err != nil {
		err = fmt.Errorf("error reading state data: %w", err)
		return getStateOutput{}, err
	}
	if found == false {
		err = fmt.Errorf("the file %s was not found", path)
		return getStateOutput{}, err
	}

	// the key for this map is "field"
	nearestBefore := make(map[string]fieldData)
	nearestAfter := make(map[string]fieldData)

	// unpack the json lines and find our output values
	//
	// time complexity => O(n), we only iterate through the input data once
	// space complexity => O(n) but nearly O(k), we only need to store nearestBefore and nearestAfter
	// 		in memory, but the way this is written causes all of the fileData to be read into memory
	// 		at once. A potential future optimization would be to only load one line of the file into
	// 		memory at a time, which would make this function O(k).
	for lineNumber, lineString := range strings.Split(fileData, "\n") {
		// skip empty lines
		lineString = strings.TrimSpace(lineString)
		if lineString == "" {
			continue
		}

		// get json data
		var lineData fileLineJSON
		err := json.Unmarshal([]byte(lineString), &lineData)
		if err != nil {
			err = fmt.Errorf("error reading json line number (%d) for file (%s): %w", lineNumber, path, err)
		}

		// get changeTime from json data
		changeTime, err := stringToTime(lineData.ChangeTime)
		if err != nil {
			err = fmt.Errorf("error parsing changeTime for json line number (%d) for file (%s): %w", lineNumber, path, err)
			return getStateOutput{}, err
		}

		nearestBefore = setNearest(setNearestInput{
			// shared fields
			inputFields:   input.fields,
			changeTime:    changeTime,
			inputDateTime: inputDateTime,
			// changing fields
			debugString:   "nearestBefore",
			fieldData:     lineData.After,    // the nearest before uses the *after* attribute
			firstCompare:  changeTime.Before, // the nearest before is *before* our input time
			secondCompare: changeTime.After,  // if this is the new nearest before, it should be *after* the existing one
			nearest:       nearestBefore,
		})

		nearestAfter = setNearest(setNearestInput{
			// shared fields
			inputFields:   input.fields,
			changeTime:    changeTime,
			inputDateTime: inputDateTime,
			// changing fields
			debugString:   "nearestAfter",
			fieldData:     lineData.Before,   // the nearest after uses the *before* attribute
			firstCompare:  changeTime.After,  // the nearest after is *after* our input time
			secondCompare: changeTime.Before, // if this is the new nearest after, it should be *before* the existing one
			nearest:       nearestAfter,
		})
	}

	output.State = make(map[string]interface{})
	output.State, err = updateOutputState(output.State, nearestBefore, nearestAfter)
	if err != nil {
		return getStateOutput{}, err
	}
	output.State, err = updateOutputState(output.State, nearestAfter, nearestBefore)
	if err != nil {
		return getStateOutput{}, err
	}

	if len(output.State) == 0 {
		err = fmt.Errorf("no data found for fields %s", input.fields)
		return getStateOutput{}, err
	}

	// as far as I can tell, the output time is literally just the input time but
	// displayed at a higher percision???
	output.Ts = inputDateTime.Format("2006-01-02T15:04:05")

	return output, nil
}

type setNearestInput struct {
	debugString   string
	fieldData     map[string]interface{}
	inputFields   []string
	firstCompare  func(time.Time) bool
	secondCompare func(time.Time) bool
	inputDateTime time.Time
	changeTime    time.Time
	nearest       map[string]fieldData
}

func setNearest(input setNearestInput) map[string]fieldData {
	for checkingField, value := range input.fieldData {
		for _, inputField := range input.inputFields {
			if checkingField == inputField {
				// set values if the existing values are empty
				if input.firstCompare(input.inputDateTime) && input.nearest[checkingField].time.IsZero() {
					input.nearest[checkingField] = fieldData{
						value: value,
						time:  input.changeTime,
					}
					logrus.Debugf("%s %s (was empty) => %+v\n", input.debugString, checkingField, value)
				}
				// set values if the time comparison succeed
				if input.firstCompare(input.inputDateTime) && input.secondCompare(input.nearest[checkingField].time) {
					input.nearest[checkingField] = fieldData{
						value: value,
						time:  input.changeTime,
					}
					logrus.Debugf("%s %s (comparison succeed) => %+v\n", input.debugString, checkingField, value)
				}
			}
		}
	}
	return input.nearest
}

func updateOutputState(outputState map[string]interface{}, sourceNearest map[string]fieldData, otherNearest map[string]fieldData) (map[string]interface{}, error) {
	for field, fieldData := range sourceNearest {
		if !fieldData.time.IsZero() {
			otherFieldData := otherNearest[field]
			if !otherFieldData.time.IsZero() && fieldData.value != otherFieldData.value {
				err := fmt.Errorf("data error, mismatched values on \"before\" and \"after\" (%v, %v) data for the field %s", fieldData.value, otherFieldData.value, field)
				return outputState, err
			}
			outputState[field] = fieldData.value
		}
	}
	return outputState, nil
}
