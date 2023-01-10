// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package hooks

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

func Register(app *kingpin.Application) {
	cmd := app.Command("hooks", "manage git server hooks")
	registerUpdate(cmd)
	registerPostReceive(cmd)
	registerPreReceive(cmd)
}