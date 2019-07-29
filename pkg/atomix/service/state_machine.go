package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"strings"
	"time"
)

// Context provides information about the context within which a state machine is running
type Context interface {
	// Index returns the current index of the state machine
	Index() uint64

	// Timestamp returns a deterministic, monotonically increasing timestamp
	Timestamp() time.Time

	// OperationType returns the type of the operation currently being executed against the state machine
	OperationType() OperationType
}

// StateMachine applies commands from a protocol to a collection of state machines
type StateMachine interface {
	// Snapshot writes the state machine snapshot to the given writer
	Snapshot(writer io.Writer) error

	// Install reads the state machine snapshot from the given reader
	Install(reader io.Reader) error

	// CanDelete returns a bool indicating whether the node can delete changes up to the given index without affecting
	// the correctness of the state machine
	CanDelete(index uint64) bool

	// Command applies a command to the state machine
	Command(bytes []byte, ch chan<- Output)

	// Query applies a query to the state machine
	Query(bytes []byte, ch chan<- Output)
}

func newOutput(value []byte, err error) Output {
	return Output{
		Value: value,
		Error: err,
	}
}

func newFailure(err error) Output {
	return Output{
		Error: err,
	}
}

// Output is a state machine operation output
type Output struct {
	Value []byte
	Error error
}

// Failed returns a boolean indicating whether the operation failed
func (r Output) Failed() bool {
	return r.Error != nil
}

// Succeeded returns a boolean indicating whether the operation was successful
func (r Output) Succeeded() bool {
	return !r.Failed()
}

// Result is a state machine operation result
type Result struct {
	Output
	Index uint64
}

// newPrimitiveStateMachine returns a new primitive state machine
func NewPrimitiveStateMachine(registry *ServiceRegistry, ctx Context) StateMachine {
	return &primitiveStateMachine{
		ctx:      ctx,
		registry: registry,
		services: make(map[string]*serviceStateMachine),
	}
}

// primitiveStateMachine is the primary state machine for managing primitive services
type primitiveStateMachine struct {
	StateMachine
	ctx      Context
	registry *ServiceRegistry
	services map[string]*serviceStateMachine
}

func (s *primitiveStateMachine) Snapshot(writer io.Writer) error {
	for id, service := range s.services {
		serviceId := &ServiceId{
			Type:      service.Type,
			Name:      strings.Split(id, ":")[0],
			Namespace: strings.Split(id, ":")[1],
		}
		bytes, err := proto.Marshal(serviceId)
		if err != nil {
			return err
		} else {
			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(len(bytes)))
			writer.Write(length)
			writer.Write(bytes)
			service.Snapshot(writer)
		}
	}
	return nil
}

func (s *primitiveStateMachine) Install(reader io.Reader) error {
	s.services = make(map[string]*serviceStateMachine)
	lengthBytes := make([]byte, 4)
	n, err := reader.Read(lengthBytes)
	if err != nil {
		return err
	}
	for n > 0 {
		length := binary.BigEndian.Uint32(lengthBytes)
		bytes := make([]byte, length)
		n, err = reader.Read(bytes)
		if err != nil {
			return err
		}

		serviceId := &ServiceId{}
		if err = proto.Unmarshal(bytes, serviceId); err != nil {
			return err
		}
		service := s.registry.types[serviceId.Type](s.ctx)
		s.services[getServiceName(serviceId)] = newServiceStateMachine(serviceId.Type, service)

		n, err = reader.Read(lengthBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *primitiveStateMachine) Command(bytes []byte, ch chan<- Output) {
	request := &ServiceRequest{}
	err := proto.Unmarshal(bytes, request)
	if err != nil {
		if ch != nil {
			ch <- newFailure(err)
		}
	} else {
		switch r := request.Request.(type) {
		case *ServiceRequest_Command:
			// If the service doesn't exist, create it.
			service, ok := s.services[getServiceName(request.Id)]
			if !ok {
				serviceType := s.registry.getType(request.Id.Type)
				if serviceType == nil {
					if ch != nil {
						ch <- newFailure(errors.New(fmt.Sprintf("unknown service type %s", request.Id.Type)))
						return
					}
				} else {
					service = newServiceStateMachine(request.Id.Type, serviceType(s.ctx))
					s.services[getServiceName(request.Id)] = service
				}
			}

			// Create a channel for the raw service results
			var serviceCh chan Output
			if ch != nil {
				serviceCh = make(chan Output)

				// Start a goroutine to encode the raw service results in a ServiceResponse
				go func() {
					for result := range serviceCh {
						if result.Failed() {
							ch <- result
						} else {
							ch <- newOutput(proto.Marshal(&ServiceResponse{
								Response: &ServiceResponse_Command{
									Command: result.Value,
								},
							}))
						}
					}
				}()
			}

			// Execute the command on the service
			service.Command(r.Command, serviceCh)
		case *ServiceRequest_Create:
			_, ok := s.services[getServiceName(request.Id)]
			if !ok {
				serviceType := s.registry.getType(request.Id.Type)
				if serviceType == nil {
					if ch != nil {
						ch <- newFailure(errors.New(fmt.Sprintf("unknown service type %s", request.Id.Type)))
					}
				} else {
					service := serviceType(s.ctx)
					s.services[getServiceName(request.Id)] = newServiceStateMachine(request.Id.Type, service)

					if ch != nil {
						ch <- newOutput(proto.Marshal(&ServiceResponse{
							Response: &ServiceResponse_Create{
								Create: &CreateResponse{},
							},
						}))
					}
				}
			} else {
				if ch != nil {
					ch <- newOutput(proto.Marshal(&ServiceResponse{
						Response: &ServiceResponse_Create{
							Create: &CreateResponse{},
						},
					}))
				}
			}
		case *ServiceRequest_Delete:
			delete(s.services, getServiceName(request.Id))

			if ch != nil {
				ch <- newOutput(proto.Marshal(&ServiceResponse{
					Response: &ServiceResponse_Delete{
						Delete: &DeleteResponse{},
					},
				}))
			}
		}
	}
}

func getServiceName(id *ServiceId) string {
	return id.Name + "." + id.Namespace
}

func (s *primitiveStateMachine) Query(bytes []byte, ch chan<- Output) {
	request := &ServiceRequest{}
	err := proto.Unmarshal(bytes, request)
	if err != nil {
		if ch != nil {
			ch <- newFailure(err)
		}
	} else {
		switch r := request.Request.(type) {
		case *ServiceRequest_Query:
			service, ok := s.services[getServiceName(request.Id)]
			if !ok {
				if ch != nil {
					ch <- newFailure(errors.New(fmt.Sprintf("unknown service %s", getServiceName(request.Id))))
				}
			} else {
				// Create a channel for the raw service results
				var serviceCh chan Output
				if ch != nil {
					serviceCh := make(chan Output)

					// Start a goroutine to encode the raw service results in a ServiceResponse
					go func() {
						for result := range serviceCh {
							if result.Failed() {
								ch <- result
							} else {
								ch <- newOutput(proto.Marshal(&ServiceResponse{
									Response: &ServiceResponse_Query{
										Query: result.Value,
									},
								}))
							}
						}
					}()
				}

				// Execute the query on the service
				service.Query(r.Query, serviceCh)
			}
		case *ServiceRequest_Metadata:
			services := make([]*ServiceId, 0, len(s.services))
			for id, service := range s.services {
				if r.Metadata.Type == "" || service.Type == r.Metadata.Type {
					services = append(services, &ServiceId{
						Type:      service.Type,
						Name:      strings.Split(id, ":")[0],
						Namespace: strings.Split(id, ":")[1],
					})
				}
			}

			if ch != nil {
				ch <- newOutput(proto.Marshal(&ServiceResponse{
					Response: &ServiceResponse_Metadata{
						Metadata: &MetadataResponse{
							Services: services,
						},
					},
				}))
			}
		}
	}
}

// newServiceStateMachine returns a new wrapped service
func newServiceStateMachine(serviceType string, service Service) *serviceStateMachine {
	return &serviceStateMachine{
		Type:    serviceType,
		service: service,
	}
}

// serviceStateMachine is a typed wrapper around a service
type serviceStateMachine struct {
	StateMachine
	Type    string
	service Service
}

func (s *serviceStateMachine) Snapshot(writer io.Writer) {
	s.service.Snapshot(writer)
}

func (s *serviceStateMachine) Install(reader io.Reader) {
	s.service.Install(reader)
}

func (s *serviceStateMachine) CanDelete(index uint64) bool {
	return s.service.CanDelete(index)
}

func (s *serviceStateMachine) Command(bytes []byte, ch chan<- Output) {
	s.service.Command(bytes, ch)
}

func (s *serviceStateMachine) Query(bytes []byte, ch chan<- Output) {
	s.service.Query(bytes, ch)
}