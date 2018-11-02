// Copyright 2018 John Deng (hi.devops.io@gmail.com).
//
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

package cmd

import (
	"fmt"
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/cli"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/mio-cli/pkg/types"
	"github.com/hidevopsio/mio-cli/pkg/utils"
	"path/filepath"
	"time"
)

type runCommand struct {
	cli.SubCommand

	profile    string
	project    string
	sourcecode string
	app        string
	verbose    string
}

func newRunCommand() *runCommand {
	c := &runCommand{
	}
	c.Use = "run"
	c.Short = "run command"
	c.Long = "Run run command"

	pf := c.PersistentFlags()
	pf.StringVarP(&c.profile, "profile", "P", "dev", "--profile=test")
	pf.StringVarP(&c.project, "project", "p", "", "--project=project-name")
	pf.StringVarP(&c.app, "app", "a", "", "--app=my-app")
	pf.StringVarP(&c.sourcecode, "sourcecode", "s", "", "--sourcecode=java")
	pf.StringVarP(&c.verbose, "verbose", "v", "", "--verbose=true")
	return c
}

func init() {
	app.Register(newRunCommand)
}

//mio-cli run --profile=test --project=demo --app=hello-world --sourcecode=java --verbose=true
func (c *runCommand) Run(args []string) error {
	log.Info("handle run command")

	homeDir, err := utils.GetHomeDir()
	if err != nil {
		log.Info("Error", err)
		return err
	}

	filePath := filepath.Join(homeDir, types.TOKEN_DIR, types.TOKEN_FILE)

	token, err := utils.ReadText(filePath)
	if err != nil {
		log.Info("Error", err)
		return err
	}

	if err := utils.Start(types.PipelineStart{Name: c.app, Namespace: c.project, SourceCode: c.sourcecode}, types.PIPELINE_START_API, token); err != nil {
		fmt.Println("Error ", err)
		return err
	}

	time.Sleep(time.Second * 3)
	var url = fmt.Sprintf("%s?namespace=%s&name=%s&sourcecode=%s&verbose=%s",types.BUILDLOG_API,c.project,c.app,c.sourcecode,c.verbose)
	//utils.WsLogsOut("localhost:8080", "/logs", fmt.Sprintf("namespace=%s&name=%s", c.project, c.app))
	if err := utils.ClientLoop(url);err != nil {
		fmt.Println("Error ", err)
		return err
	}
	return nil
}
