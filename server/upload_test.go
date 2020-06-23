// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteFile(t *testing.T) {
	// Write the input file
	input_data := []byte("input data\n")
	err := ioutil.WriteFile("input_file.txt", input_data, 0644)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test function
	input_file, err := os.Open("input_file.txt")
	err = WriteFile("output_file.txt", input_file)
	input_file.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Read back data and check if it is the same
	output_data, err := ioutil.ReadFile("output_file.txt")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output_data) != string(input_data) {
		t.Fatalf("Wrote %v but expected %v", string(output_data), string(input_data))
	}
}
