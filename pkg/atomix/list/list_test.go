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

package list

import (
	"context"
	client "github.com/atomix/atomix-go-client/pkg/client/list"
	"github.com/atomix/atomix-go-client/pkg/client/primitive"
	"github.com/atomix/atomix-go-client/pkg/client/session"
	"github.com/atomix/atomix-go-node/pkg/atomix/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestList(t *testing.T) {
	node, conn := test.StartTestNode()
	defer node.Stop()

	name := primitive.NewName("default", "test", "default", "test")
	list, err := client.New(context.TODO(), name, []*grpc.ClientConn{conn}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)
	assert.NotNil(t, list)

	size, err := list.Len(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 0, size)

	_, err = list.Get(context.TODO(), 0)
	assert.EqualError(t, err, "index out of bounds")

	err = list.Append(context.TODO(), "foo")
	assert.NoError(t, err)

	size, err = list.Len(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 1, size)

	value, err := list.Get(context.TODO(), 0)
	assert.NoError(t, err)
	assert.Equal(t, "foo", value)

	err = list.Append(context.TODO(), "bar")
	assert.NoError(t, err)

	size, err = list.Len(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 2, size)

	err = list.Insert(context.TODO(), 1, "baz")
	assert.NoError(t, err)

	size, err = list.Len(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 3, size)

	ch := make(chan string)
	err = list.Items(context.TODO(), ch)
	assert.NoError(t, err)

	value, ok := <-ch
	assert.True(t, ok)
	assert.Equal(t, "foo", value)
	value, ok = <-ch
	assert.True(t, ok)
	assert.Equal(t, "baz", value)
	value, ok = <-ch
	assert.True(t, ok)
	assert.Equal(t, "bar", value)

	_, ok = <-ch
	assert.False(t, ok)

	events := make(chan *client.Event)
	err = list.Watch(context.TODO(), events)
	assert.NoError(t, err)

	done := make(chan struct{})
	go func() {
		event := <-events
		assert.Equal(t, client.EventInserted, event.Type)
		assert.Equal(t, 3, event.Index)
		assert.Equal(t, "Hello world!", event.Value)

		event = <-events
		assert.Equal(t, client.EventInserted, event.Type)
		assert.Equal(t, 2, event.Index)
		assert.Equal(t, "Hello world again!", event.Value)

		event = <-events
		assert.Equal(t, client.EventRemoved, event.Type)
		assert.Equal(t, 1, event.Index)
		assert.Equal(t, "baz", event.Value)

		event = <-events
		assert.Equal(t, client.EventRemoved, event.Type)
		assert.Equal(t, 1, event.Index)
		assert.Equal(t, "Hello world again!", event.Value)

		event = <-events
		assert.Equal(t, client.EventInserted, event.Type)
		assert.Equal(t, 1, event.Index)
		assert.Equal(t, "Not hello world!", event.Value)

		close(done)
	}()

	err = list.Append(context.TODO(), "Hello world!")
	assert.NoError(t, err)

	err = list.Insert(context.TODO(), 2, "Hello world again!")
	assert.NoError(t, err)

	value, err = list.Remove(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "baz", value)

	err = list.Set(context.TODO(), 1, "Not hello world!")
	assert.NoError(t, err)

	<-done

	err = list.Close()
	assert.NoError(t, err)

	list1, err := client.New(context.TODO(), name, []*grpc.ClientConn{conn}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)

	list2, err := client.New(context.TODO(), name, []*grpc.ClientConn{conn}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)

	size, err = list1.Len(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 4, size)

	err = list1.Close()
	assert.NoError(t, err)

	err = list1.Delete()
	assert.NoError(t, err)

	err = list2.Delete()
	assert.NoError(t, err)

	list, err = client.New(context.TODO(), name, []*grpc.ClientConn{conn}, session.WithTimeout(5*time.Second))
	assert.NoError(t, err)

	size, err = list.Len(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 0, size)

}
