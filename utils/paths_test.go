// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"strings"
	"testing"
)

func TestExpandUser(t *testing.T) {
	if result := ExpandUser("~"); result == "~" {
		t.Error("Unable to expand user for ~")
	}
	if result := ExpandUser("~/path/to/file"); strings.HasPrefix(result, "~") {
		t.Error("Unable to expand user for path starting with ~")
	}
	if result := ExpandUser("/path/to/file"); result != "/path/to/file" {
		t.Error("Path without ~ is altered")
	}
}
