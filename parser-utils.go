package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// ProcessTimestamp converts the Timestamp column from US/Pacific to RFC3339 US/Eastern.
func ProcessTimestamp(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	timestampIndex := -1
	for i, h := range header {
		if h == "Timestamp" {
			timestampIndex = i
			break
		}
	}
	if timestampIndex == -1 {
		return nil, fmt.Errorf("timestamp column not found")
	}

	const inputLayout = "1/2/06 3:04:05 PM"

	locPacific, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, fmt.Errorf("failed to load Pacific timezone: %v", err)
	}
	locEastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, fmt.Errorf("failed to load Eastern timezone: %v", err)
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		if timestampIndex < len(newRow) {
			t, err := time.ParseInLocation(inputLayout, newRow[timestampIndex], locPacific)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not parse timestamp %q: %v\n", newRow[timestampIndex], err)
			} else {
				newRow[timestampIndex] = t.In(locEastern).Format(time.RFC3339)
			}
		}
		updated = append(updated, newRow)
	}

	return updated, nil
}

// ProcessZIPs pads all numeric ZIP codes to 5 digits.
func ProcessZIPs(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	zipIndex := -1
	for i, h := range header {
		if strings.EqualFold(h, "ZIP") {
			zipIndex = i
			break
		}
	}
	if zipIndex == -1 {
		return nil, fmt.Errorf("ZIP column not found")
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		if zipIndex < len(newRow) {
			zip := strings.TrimSpace(newRow[zipIndex])
			if zip == "" {
			} else {
				if _, err := strconv.Atoi(zip); err == nil {
					newRow[zipIndex] = fmt.Sprintf("%05s", zip)
				} else {
					fmt.Fprintf(os.Stderr, "Warning: invalid ZIP value %q\n", zip)
				}
			}
		}

		updated = append(updated, newRow)
	}

	return updated, nil
}

// ProcessFirstName capitalizes the first word of FullName.
func ProcessFirstName(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	fullNameIndex := -1
	for i, h := range header {
		if strings.EqualFold(h, "FullName") {
			fullNameIndex = i
			break
		}
	}
	if fullNameIndex == -1 {
		return nil, fmt.Errorf("FullName column not found")
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		if fullNameIndex < len(newRow) {
			original := strings.TrimSpace(newRow[fullNameIndex])
			if original != "" {
				parts := strings.Fields(original)
				if len(parts) > 0 {
					parts[0] = strings.ToUpper(parts[0])
					newRow[fullNameIndex] = strings.Join(parts, " ")
				}
			}
		}

		updated = append(updated, newRow)
	}

	return updated, nil
}

// ProcessAddress checks that all values in the Address column are valid UTF-8.
func ProcessAddress(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	addressIndex := -1
	for i, h := range header {
		if strings.EqualFold(h, "Address") {
			addressIndex = i
			break
		}
	}
	if addressIndex == -1 {
		return nil, fmt.Errorf("address column not found")
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		if addressIndex < len(newRow) {
			val := newRow[addressIndex]
			if !utf8.ValidString(val) {
				fmt.Fprintf(os.Stderr, "Warning: invalid UTF-8 in Address: %q\n", val)
			}
		}

		updated = append(updated, newRow)
	}

	return updated, nil
}

// ProcessDurations converts FooDuration and BarDuration fields to float seconds.
func ProcessDurations(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	fooIndex := -1
	barIndex := -1
	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "FooDuration":
			fooIndex = i
		case "BarDuration":
			barIndex = i
		}
	}
	if fooIndex == -1 || barIndex == -1 {
		return nil, fmt.Errorf("FooDuration or BarDuration column not found")
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		// Convert FooDuration
		if fooIndex < len(newRow) {
			sec, err := parseHHMMSS(newRow[fooIndex])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not parse FooDuration %q: %v\n", newRow[fooIndex], err)
			} else {
				newRow[fooIndex] = fmt.Sprintf("%.3f", sec)
			}
		}

		// Convert BarDuration
		if barIndex < len(newRow) {
			sec, err := parseHHMMSS(newRow[barIndex])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not parse BarDuration %q: %v\n", newRow[barIndex], err)
			} else {
				newRow[barIndex] = fmt.Sprintf("%.3f", sec)
			}
		}

		updated = append(updated, newRow)
	}

	return updated, nil
}

// parseHHMMSS parses a duration in HH:MM:SS.MS format and returns total seconds.
func parseHHMMSS(input string) (float64, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid format: %q", input)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %v", err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %v", err)
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %v", err)
	}

	total := float64(hours*3600 + minutes*60)
	return total + seconds, nil
}

// ProcessTotalDuration replaces the TotalDuration column with the sum of FooDuration and BarDuration.
func ProcessTotalDuration(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	fooIndex := -1
	barIndex := -1
	totalIndex := -1

	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "FooDuration":
			fooIndex = i
		case "BarDuration":
			barIndex = i
		case "TotalDuration":
			totalIndex = i
		}
	}

	if fooIndex == -1 || barIndex == -1 || totalIndex == -1 {
		return nil, fmt.Errorf("required duration columns not found")
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		foo, err1 := strconv.ParseFloat(strings.TrimSpace(newRow[fooIndex]), 64)
		bar, err2 := strconv.ParseFloat(strings.TrimSpace(newRow[barIndex]), 64)

		if err1 != nil || err2 != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse Foo/Bar durations for row: %v\n", row)
			newRow[totalIndex] = ""
		} else {
			total := foo + bar
			newRow[totalIndex] = fmt.Sprintf("%.3f", total)
		}

		updated = append(updated, newRow)
	}

	return updated, nil
}

// ProcessNotes cleans the notes column of any non utf-8 characters
func ProcessNotes(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return records, nil
	}

	header := records[0]
	notesIndex := -1
	for i, h := range header {
		if strings.EqualFold(h, "Notes") {
			notesIndex = i
			break
		}
	}
	if notesIndex == -1 {
		return nil, fmt.Errorf("notes column not found")
	}

	var updated [][]string
	updated = append(updated, header)

	for _, row := range records[1:] {
		newRow := make([]string, len(row))
		copy(newRow, row)

		if notesIndex < len(newRow) {
			val := newRow[notesIndex]
			if !utf8.ValidString(val) {
				cleaned, _, err := transform.String(unicode.UTF8.NewDecoder(), val)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to clean invalid UTF-8 in Notes: %v\n", err)
				} else {
					newRow[notesIndex] = cleaned
				}
			}
		}

		updated = append(updated, newRow)
	}

	return updated, nil
}
