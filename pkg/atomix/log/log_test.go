// Copyright 2020-present Open Networking Foundation.
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

package log

import (
	"context"
	"testing"
	"time"

	client "github.com/atomix/go-client/pkg/client/log"

	"github.com/atomix/go-client/pkg/client/primitive"
	"github.com/atomix/go-client/pkg/client/session"
	"github.com/atomix/go-client/pkg/client/util/net"
	"github.com/atomix/go-framework/pkg/atomix/test"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	address, node := test.StartTestNode()
	defer node.Stop()

	name := primitive.NewName("default", "test", "default", "test")
	_, err := client.New(context.TODO(), name, []net.Address{address}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)
}
