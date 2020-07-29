// Copyright 2020 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package conf

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/G-Node/gogs/internal/process"
)

// openSSHVersion returns string representation of OpenSSH version via command "ssh -V".
func openSSHVersion() (string, error) {
	// NOTE: Somehow the version is printed to stderr.
	_, stderr, err := process.Exec("conf.openSSHVersion", "ssh", "-V")
	if err != nil {
		return "", errors.Wrap(err, stderr)
	}

	// Trim unused information, see https://github.com/gogs/gogs/issues/4507#issuecomment-305150441.
	v := strings.TrimRight(strings.Fields(stderr)[0], ",1234567890")
	v = strings.TrimSuffix(strings.TrimPrefix(v, "OpenSSH_"), "p")
	return v, nil
}

// ensureAbs prepends the WorkDir to the given path if it is not an absolute path.
func ensureAbs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(WorkDir(), path)
}
