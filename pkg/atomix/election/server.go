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

package election

import (
	"context"
	api "github.com/atomix/atomix-api/proto/atomix/election"
	"github.com/atomix/atomix-api/proto/atomix/headers"
	"github.com/atomix/atomix-go-node/pkg/atomix/node"
	"github.com/atomix/atomix-go-node/pkg/atomix/server"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func init() {
	node.RegisterServer(registerServer)
}

// registerServer registers an election server with the given gRPC server
func registerServer(server *grpc.Server, protocol node.Protocol) {
	api.RegisterLeaderElectionServiceServer(server, newServer(protocol.Client()))
}

func newServer(client node.Client) api.LeaderElectionServiceServer {
	return &Server{
		SessionizedServer: &server.SessionizedServer{
			Type:   "election",
			Client: client,
		},
	}
}

// Server is an implementation of LeaderElectionServiceServer for the election primitive
type Server struct {
	*server.SessionizedServer
}

// Enter enters a candidate in the election
func (s *Server) Enter(ctx context.Context, request *api.EnterRequest) (*api.EnterResponse, error) {
	log.Tracef("Received EnterRequest %+v", request)
	in, err := proto.Marshal(&EnterRequest{
		ID: request.CandidateID,
	})
	if err != nil {
		return nil, err
	}

	out, header, err := s.Command(ctx, "Enter", in, request.Header)
	if err != nil {
		return nil, err
	}

	enterResponse := &EnterResponse{}
	if err = proto.Unmarshal(out, enterResponse); err != nil {
		return nil, err
	}

	response := &api.EnterResponse{
		Header: header,
		Term: &api.Term{
			ID:         enterResponse.Term.ID,
			Timestamp:  enterResponse.Term.Timestamp,
			Leader:     enterResponse.Term.Leader,
			Candidates: enterResponse.Term.Candidates,
		},
	}
	log.Tracef("Sending EnterResponse %+v", response)
	return response, nil
}

// Withdraw withdraws a candidate from the election
func (s *Server) Withdraw(ctx context.Context, request *api.WithdrawRequest) (*api.WithdrawResponse, error) {
	log.Tracef("Received WithdrawRequest %+v", request)
	in, err := proto.Marshal(&WithdrawRequest{
		ID: request.CandidateID,
	})
	if err != nil {
		return nil, err
	}

	out, header, err := s.Command(ctx, "Withdraw", in, request.Header)
	if err != nil {
		return nil, err
	}

	withdrawResponse := &WithdrawResponse{}
	if err = proto.Unmarshal(out, withdrawResponse); err != nil {
		return nil, err
	}

	response := &api.WithdrawResponse{
		Header: header,
		Term: &api.Term{
			ID:         withdrawResponse.Term.ID,
			Timestamp:  withdrawResponse.Term.Timestamp,
			Leader:     withdrawResponse.Term.Leader,
			Candidates: withdrawResponse.Term.Candidates,
		},
	}
	log.Tracef("Sending WithdrawResponse %+v", response)
	return response, nil
}

// Anoint assigns leadership to a candidate
func (s *Server) Anoint(ctx context.Context, request *api.AnointRequest) (*api.AnointResponse, error) {
	log.Tracef("Received AnointRequest %+v", request)
	in, err := proto.Marshal(&AnointRequest{
		ID: request.CandidateID,
	})
	if err != nil {
		return nil, err
	}

	out, header, err := s.Command(ctx, "Anoint", in, request.Header)
	if err != nil {
		return nil, err
	}

	anointResponse := &AnointResponse{}
	if err = proto.Unmarshal(out, anointResponse); err != nil {
		return nil, err
	}

	response := &api.AnointResponse{
		Header: header,
		Term: &api.Term{
			ID:         anointResponse.Term.ID,
			Timestamp:  anointResponse.Term.Timestamp,
			Leader:     anointResponse.Term.Leader,
			Candidates: anointResponse.Term.Candidates,
		},
	}
	log.Tracef("Sending AnointResponse %+v", response)
	return response, nil
}

// Promote increases the priority of a candidate
func (s *Server) Promote(ctx context.Context, request *api.PromoteRequest) (*api.PromoteResponse, error) {
	log.Tracef("Received PromoteRequest %+v", request)
	in, err := proto.Marshal(&PromoteRequest{
		ID: request.CandidateID,
	})
	if err != nil {
		return nil, err
	}

	out, header, err := s.Command(ctx, "Promote", in, request.Header)
	if err != nil {
		return nil, err
	}

	promoteResponse := &PromoteResponse{}
	if err = proto.Unmarshal(out, promoteResponse); err != nil {
		return nil, err
	}

	response := &api.PromoteResponse{
		Header: header,
		Term: &api.Term{
			ID:         promoteResponse.Term.ID,
			Timestamp:  promoteResponse.Term.Timestamp,
			Leader:     promoteResponse.Term.Leader,
			Candidates: promoteResponse.Term.Candidates,
		},
	}
	log.Tracef("Sending PromoteResponse %+v", response)
	return response, nil
}

// Evict removes a candidate from the election
func (s *Server) Evict(ctx context.Context, request *api.EvictRequest) (*api.EvictResponse, error) {
	log.Tracef("Received EvictRequest %+v", request)
	in, err := proto.Marshal(&EvictRequest{
		ID: request.CandidateID,
	})
	if err != nil {
		return nil, err
	}

	out, header, err := s.Command(ctx, "Evict", in, request.Header)
	if err != nil {
		return nil, err
	}

	evictResponse := &EvictResponse{}
	if err = proto.Unmarshal(out, evictResponse); err != nil {
		return nil, err
	}

	response := &api.EvictResponse{
		Header: header,
		Term: &api.Term{
			ID:         evictResponse.Term.ID,
			Timestamp:  evictResponse.Term.Timestamp,
			Leader:     evictResponse.Term.Leader,
			Candidates: evictResponse.Term.Candidates,
		},
	}
	log.Tracef("Sending EvictResponse %+v", response)
	return response, nil
}

// GetTerm gets the current election term
func (s *Server) GetTerm(ctx context.Context, request *api.GetTermRequest) (*api.GetTermResponse, error) {
	log.Tracef("Received GetTermRequest %+v", request)
	in, err := proto.Marshal(&GetTermRequest{})
	if err != nil {
		return nil, err
	}

	out, header, err := s.Query(ctx, "GetTerm", in, request.Header)
	if err != nil {
		return nil, err
	}

	getResponse := &GetTermResponse{}
	if err = proto.Unmarshal(out, getResponse); err != nil {
		return nil, err
	}

	response := &api.GetTermResponse{
		Header: header,
		Term: &api.Term{
			ID:         getResponse.Term.ID,
			Timestamp:  getResponse.Term.Timestamp,
			Leader:     getResponse.Term.Leader,
			Candidates: getResponse.Term.Candidates,
		},
	}
	log.Tracef("Sending GetTermResponse %+v", response)
	return response, nil
}

// Events lists for election change events
func (s *Server) Events(request *api.EventRequest, stream api.LeaderElectionService_EventsServer) error {
	log.Tracef("Received EventRequest %+v", request)
	in, err := proto.Marshal(&ListenRequest{})
	if err != nil {
		return err
	}

	ch := make(chan server.SessionOutput)
	if err := s.CommandStream("Events", in, request.Header, ch); err != nil {
		return err
	}

	for result := range ch {
		if result.Failed() {
			return result.Error
		}

		response := &ListenResponse{}
		if err = proto.Unmarshal(result.Value, response); err != nil {
			return err
		}
		eventResponse := &api.EventResponse{
			Header: result.Header,
			Type:   api.EventResponse_CHANGED,
			Term: &api.Term{
				ID:         response.Term.ID,
				Timestamp:  response.Term.Timestamp,
				Leader:     response.Term.Leader,
				Candidates: response.Term.Candidates,
			},
		}
		log.Tracef("Sending EventResponse %+v", response)
		if err = stream.Send(eventResponse); err != nil {
			return err
		}
	}
	log.Tracef("Finished EventRequest %+v", request)
	return nil
}

// Create opens a new session
func (s *Server) Create(ctx context.Context, request *api.CreateRequest) (*api.CreateResponse, error) {
	log.Tracef("Received CreateRequest %+v", request)
	session, err := s.OpenSession(ctx, request.Header, request.Timeout)
	if err != nil {
		return nil, err
	}
	response := &api.CreateResponse{
		Header: &headers.ResponseHeader{
			SessionID: session,
			Index:     session,
		},
	}
	log.Tracef("Sending CreateResponse %+v", response)
	return response, nil
}

// KeepAlive keeps an existing session alive
func (s *Server) KeepAlive(ctx context.Context, request *api.KeepAliveRequest) (*api.KeepAliveResponse, error) {
	log.Tracef("Received KeepAliveRequest %+v", request)
	if err := s.KeepAliveSession(ctx, request.Header); err != nil {
		return nil, err
	}
	response := &api.KeepAliveResponse{
		Header: &headers.ResponseHeader{
			SessionID: request.Header.SessionID,
		},
	}
	log.Tracef("Sending KeepAliveResponse %+v", response)
	return response, nil
}

// Close closes a session
func (s *Server) Close(ctx context.Context, request *api.CloseRequest) (*api.CloseResponse, error) {
	log.Tracef("Received CloseRequest %+v", request)
	if request.Delete {
		if err := s.Delete(ctx, request.Header); err != nil {
			return nil, err
		}
	} else {
		if err := s.CloseSession(ctx, request.Header); err != nil {
			return nil, err
		}
	}

	response := &api.CloseResponse{
		Header: &headers.ResponseHeader{
			SessionID: request.Header.SessionID,
		},
	}
	log.Tracef("Sending CloseResponse %+v", response)
	return response, nil
}
