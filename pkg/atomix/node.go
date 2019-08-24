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

package atomix

import (
	"fmt"
	"github.com/atomix/atomix-api/proto/atomix/controller"
	"github.com/atomix/atomix-go-node/pkg/atomix/counter"
	"github.com/atomix/atomix-go-node/pkg/atomix/election"
	"github.com/atomix/atomix-go-node/pkg/atomix/list"
	"github.com/atomix/atomix-go-node/pkg/atomix/lock"
	map_ "github.com/atomix/atomix-go-node/pkg/atomix/map"
	"github.com/atomix/atomix-go-node/pkg/atomix/primitive"
	"github.com/atomix/atomix-go-node/pkg/atomix/service"
	"github.com/atomix/atomix-go-node/pkg/atomix/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

// Protocol
type Protocol interface {
	// Start starts the protocol
	Start(cluster Cluster, registry *service.ServiceRegistry) error

	// Client returns the protocol client
	Client() service.Client

	// Stop stops the protocol
	Stop() error
}

// NewNode creates a new node running the given protocol
func NewNode(nodeID string, config *controller.PartitionConfig, protocol Protocol, opts ...NodeOption) *Node {
	node := &Node{
		Id:       nodeID,
		config:   config,
		protocol: protocol,
	}
	(&defaultOption{}).apply(node)
	for _, opt := range opts {
		opt.apply(node)
	}
	return node
}

// NodeOption is an option for constructing a Node
type NodeOption interface {
	apply(*Node)
}

// defaultOption is a node option which applies initial defaults
type defaultOption struct{}

func (o *defaultOption) apply(node *Node) {
	node.port = 5678
	node.listener = tcpListener{}
}

// WithLocal sets the node to local mode for testing
func WithLocal(lis net.Listener) NodeOption {
	return &localOption{lis}
}

type localOption struct {
	listener net.Listener
}

func (o *localOption) apply(node *Node) {
	node.listener = localListener{o.listener}
}

// WithPort sets the port on the node
func WithPort(port int) NodeOption {
	return &portOption{port: port}
}

type portOption struct {
	port int
}

func (o *portOption) apply(node *Node) {
	node.port = o.port
}

// Node is an Atomix node
type Node struct {
	Id       string
	config   *controller.PartitionConfig
	protocol Protocol
	port     int
	listener listener
	server   *grpc.Server
}

// Start starts the node
func (n *Node) Start() error {
	members := make(map[string]Member)
	for _, member := range n.config.Members {
		members[member.ID] = Member{
			ID:   member.ID,
			Host: member.Host,
			Port: int(member.Port),
		}
	}

	cluster := Cluster{
		MemberID: n.Id,
		Members:  members,
	}

	log.Info("Starting protocol")
	err := n.protocol.Start(cluster, getServiceRegistry())
	if err != nil {
		return err
	}

	lis, err := n.listener.listen(n)
	if err != nil {
		return err
	}

	// Set the ready file to indicate startup of the protocol is complete.
	ready := util.NewFileReady()
	_ = ready.Set()

	log.Info("Starting gRPC server")
	n.server = grpc.NewServer()
	registerServers(n.server, n.protocol)
	return n.server.Serve(lis)
}

// Stop stops the node
func (n *Node) Stop() error {
	n.server.Stop()
	if err := n.protocol.Stop(); err != nil {
		return err
	}
	return nil
}

type listener interface {
	listen(*Node) (net.Listener, error)
}

type tcpListener struct{}

func (l tcpListener) listen(node *Node) (net.Listener, error) {
	log.Infof("Listening on port %d", node.port)
	return net.Listen("tcp", fmt.Sprintf(":%d", node.port))
}

type localListener struct {
	listener net.Listener
}

func (l localListener) listen(node *Node) (net.Listener, error) {
	return l.listener, nil
}

// registerServers registers all primitive servers on the given gRPC server
func registerServers(server *grpc.Server, protocol Protocol) {
	primitive.RegisterPrimitiveServer(server, protocol.Client())
	counter.RegisterCounterServer(server, protocol.Client())
	election.RegisterElectionServer(server, protocol.Client())
	list.RegisterListServer(server, protocol.Client())
	lock.RegisterLockServer(server, protocol.Client())
	map_.RegisterMapServer(server, protocol.Client())
}

// getServiceRegistry returns a service registry for the node
func getServiceRegistry() *service.ServiceRegistry {
	registry := service.NewServiceRegistry()
	counter.RegisterCounterService(registry)
	election.RegisterElectionService(registry)
	list.RegisterListService(registry)
	lock.RegisterLockService(registry)
	map_.RegisterMapService(registry)
	return registry
}
