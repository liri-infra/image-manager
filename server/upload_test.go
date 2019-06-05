/****************************************************************************
 * This file is part of Liri.
 *
 * Copyright (C) 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
 *
 * $BEGIN_LICENSE:AGPL3+$
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * $END_LICENSE$
 ***************************************************************************/

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
