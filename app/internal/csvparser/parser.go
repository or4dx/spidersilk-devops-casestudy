package csvparser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
)

// CSVResult holds the parsed contents of a CSV file.
type CSVResult struct {
	Headers []string
	Rows    [][]string
}

// ParseCSV reads all records from r, treating the first row as headers.
// Returns an error for empty input or malformed CSV without panicking.
func ParseCSV(r io.Reader) (*CSVResult, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading csv: %w", err)
	}
	// Go's csv.Reader treats \n and \r\n as record terminators but not bare \r.
	// Files exported from Excel on macOS often use \r only, which causes the
	// entire file to be read as a single record. Normalise before parsing.
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	raw = bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))

	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1 // allow rows with varying column counts
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing csv: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	return &CSVResult{
		Headers: records[0],
		Rows:    records[1:],
	}, nil
}
