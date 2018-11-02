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
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/cli"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/mio-cli/pkg/types"
	"github.com/hidevopsio/mio-cli/pkg/utils"
	"os"
	"path/filepath"
)

type loginCommand struct {
	cli.SubCommand

}

func newLoginCommand() *loginCommand {
	c := &loginCommand{
	}
	c.Use = "login"
	c.Short = "login command"
	c.Long = "Run login command"

	return c
}

func init() {
	app.Register(newLoginCommand)
}

func (c *loginCommand) Run(args []string) error {
	log.Info("handle login command")

	username := utils.GetInput("username")
	password := utils.GetInput("password")

	if token, err := utils.Login(types.TOKEN_API, username, password); err != nil {
		log.Info("Error", err)
		return err
	} else {

		homeDir, err := utils.GetHomeDir()
		if err != nil {
			log.Info("Error", err)
			return err
		}

		filePath := filepath.Join(homeDir, types.TOKEN_DIR, types.TOKEN_FILE)
		if err := os.MkdirAll(filepath.Join(homeDir, types.TOKEN_DIR), 0777); err != nil {
			log.Info("Error", err)
			return err
		}

		if err := utils.WriteText(filePath, token); err != nil {
			log.Info("Error", err)
			return err
		}
	}
	log.Info("Login successful")
	return nil
}
