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
	"github.com/hidevopsio/hiboot/pkg/app/cli"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRootCommands(t *testing.T) {
	testApp := cli.NewTestApplication(t, NewRootCommand)

	t.Run("should run first command", func(t *testing.T) {
		_, err := testApp.Run("-t", "10")
		assert.Equal(t, nil, err)
	})

	t.Run("should report unknown command", func(t *testing.T) {
		_, err := testApp.Run("not-exist-command")
		assert.NotEqual(t, nil, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}
