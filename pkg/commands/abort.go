// Copyright (c) 2017 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/docopt/docopt-go"

	"github.com/projectcalico/calico-upgrade/pkg/constants"
	"github.com/projectcalico/libcalico-go/lib/upgrade/migrator"
	"github.com/projectcalico/libcalico-go/lib/upgrade/migrator/clients"
)

func Abort(args []string) {
	doc := constants.DatastoreIntro + `Usage:
  calico-upgrade abort
      [--apiconfigv3=<V3_APICONFIG>]
      [--apiconfigv1=<V1_APICONFIG>]
      [--no-prompts]

Example:
  calico-upgrade abort --apiconfigv3=/path/to/v3/config --apiconfigv1=/path/to/v1/config

Options:
  -h --help                    Show this screen.
  --apiconfigv3=<V3_APICONFIG> Path to the file containing connection
                               configuration in YAML or JSON format for
                               the Calico v3 API.
                               [default: ` + constants.DefaultConfigPathV3 + `]
  --apiconfigv1=<V1_APICONFIG> Path to the file containing connection
                               configuration in YAML or JSON format for
                               the Calico v1 API.
                               [default: ` + constants.DefaultConfigPathV1 + `]
  --no-prompts                 Do not prompt the user. We do not recommend use
                               of this option unless necessary.

Description:
  Abort an upgrade that was started using 'calico-upgrade start'. In the event
  of a failure that requires an explicit abort, the start command indicates
  that the abort command should be executed.
`
	parsedArgs, err := docopt.Parse(doc, args, true, "", false, false)
	if err != nil {
		fmt.Printf("Invalid option:\n  calico-upgrade %s\nUse flag '--help' to read about a specific subcommand.\n", strings.Join(args, " "))
		os.Exit(1)
	}
	if len(parsedArgs) == 0 {
		return
	}
	cfv3 := parsedArgs["--apiconfigv3"].(string)
	cfv1 := parsedArgs["--apiconfigv1"].(string)
	noPrompts := parsedArgs["--no-prompts"].(bool)
	ch := &cliHelper{}

	// Obtain the v1 and v3 clients.
	clientv3, clientv1, err := clients.LoadClients(cfv3, cfv1)
	if err != nil {
		ch.Separator()
		ch.Msg("Failed to abort the upgrade.")
		ch.Bullet(fmt.Sprintf("Error accessing the Calico API: %v", err))
		ch.NewLine()
		os.Exit(1)
	}

	if clientv1.IsKDD() {
		ch.Separator()
		ch.Msg("It is not necessary to abort the upgrade when using Kubernetes API " +
			"as the datastore.  No action was taken.")
		ch.NewLine()
		os.Exit(0)
	}

	m := migrator.New(clientv3, clientv1, ch)

	// The abort command is interactive to prevent accidentally kicking off the abort.
	ch.NewLine()
	ch.Msg("You are about to abort the upgrade process to Calico v3. " +
		"At this point, all calico/node instances and orchestrator plugins " +
		"(e.g. CNI) should be running Calico v2.x.")
	ch.NewLine()
	if !noPrompts {
		ch.ConfirmProceed()
	}

	// Perform the data migration.
	err = m.Abort()
	if err == nil {
		ch.Separator()
		ch.Msg("Successfully aborted the upgrade process.")
		ch.NewLine()
		return
	}

	ch.Separator()
	ch.Msg("Failed to aborted the upgrade process.")
	ch.Bullet(err.Error())
	ch.NewLine()
	os.Exit(1)
}
