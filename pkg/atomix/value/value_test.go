// Copyright 2019-present Open Networking Foundation.
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

package value

import (
	"context"
	"github.com/atomix/go-client/pkg/client/primitive"
	"github.com/atomix/go-client/pkg/client/session"
	"github.com/atomix/go-client/pkg/client/util/net"
	client "github.com/atomix/go-client/pkg/client/value"
	"github.com/atomix/go-framework/pkg/atomix/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestValue(t *testing.T) {
	address, node := test.StartTestNode()
	defer node.Stop()

	name := primitive.NewName("default", "test", "default", "test")
	value, err := client.New(context.TODO(), name, []net.Address{address}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)
	assert.NotNil(t, value)

	val, version, err := value.Get(context.TODO())
	assert.NoError(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint64(0), version)

	ch := make(chan *client.Event)

	err = value.Watch(context.TODO(), ch)
	assert.NoError(t, err)

	_, err = value.Set(context.TODO(), []byte("foo"), client.IfVersion(1))
	assert.EqualError(t, err, "version mismatch")

	_, err = value.Set(context.TODO(), []byte("foo"), client.IfValue([]byte("bar")))
	assert.EqualError(t, err, "value mismatch")

	version, err = value.Set(context.TODO(), []byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), version)

	val, version, err = value.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), version)
	assert.Equal(t, "foo", string(val))

	_, err = value.Set(context.TODO(), []byte("foo"), client.IfVersion(2))
	assert.EqualError(t, err, "version mismatch")

	version, err = value.Set(context.TODO(), []byte("bar"), client.IfVersion(1))
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), version)

	val, version, err = value.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), version)
	assert.Equal(t, "bar", string(val))

	version, err = value.Set(context.TODO(), []byte("baz"))
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), version)

	val, version, err = value.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), version)
	assert.Equal(t, "baz", string(val))

	event := <-ch
	assert.Equal(t, client.EventUpdated, event.Type)
	assert.Equal(t, uint64(1), event.Version)
	assert.Equal(t, "foo", string(event.Value))

	event = <-ch
	assert.Equal(t, client.EventUpdated, event.Type)
	assert.Equal(t, uint64(2), event.Version)
	assert.Equal(t, "bar", string(event.Value))

	event = <-ch
	assert.Equal(t, client.EventUpdated, event.Type)
	assert.Equal(t, uint64(3), event.Version)
	assert.Equal(t, "baz", string(event.Value))

	err = value.Close()
	assert.NoError(t, err)

	value1, err := client.New(context.TODO(), name, []net.Address{address}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)

	value2, err := client.New(context.TODO(), name, []net.Address{address}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)

	val, _, err = value1.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, "baz", string(val))

	err = value1.Close()
	assert.NoError(t, err)

	err = value1.Delete()
	assert.NoError(t, err)

	err = value2.Delete()
	assert.NoError(t, err)

	value, err = client.New(context.TODO(), name, []net.Address{address}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)

	val, _, err = value.Get(context.TODO())
	assert.NoError(t, err)
	assert.Nil(t, val)
}
