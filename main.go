package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	// Provide usage if all parameters aren't provided
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: normalizer <input.csv> <output.csv>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Open input csv
	f, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input CSV: %v\n", err)
		os.Exit(1)
	}

	// Process Timestamp
	records, err = ProcessTimestamp(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting timestamps: %v\n", err)
		os.Exit(1)
	}

	// Process Zip Codes
	records, err = ProcessZIPs(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error normalizing ZIPs: %v\n", err)
		os.Exit(1)
	}

	// Process First Names
	records, err = ProcessFirstName(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing FullName: %v\n", err)
		os.Exit(1)
	}

	// Process Address
	records, err = ProcessAddress(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating Address column: %v\n", err)
		os.Exit(1)
	}

	// Process Durations
	records, err = ProcessDurations(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting durations: %v\n", err)
		os.Exit(1)
	}

	// Process Total Duration
	records, err = ProcessTotalDuration(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fixing TotalDuration: %v\n", err)
		os.Exit(1)
	}

	// Process Notes
	records, err = ProcessNotes(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating Notes column: %v\n", err)
		os.Exit(1)
	}

	// Write output csv
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	w := csv.NewWriter(outFile)
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output CSV: %v\n", err)
		os.Exit(1)
	}
}
