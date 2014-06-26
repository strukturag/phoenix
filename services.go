package phoenix

import (
	"errors"
	"strings"
	"sync"
)

// Service represents a resource whose lifecycle should be managed by a Runtime.
//
// Typically this would be an exclusive resource such as a socket, database file,
// or shared memory segment.
type Service interface {
	// Start runs the main loop of the Service. It is expected to block until
	// Stop is called or the execution of the service is complete.
	//
	// Undefined behavior will result if errors are returned during shutdown,
	// such errors shall be returned by Stop.
	Start() error

	// Stop shall terminate execution of Start and may return any errors
	// reported by cleanup of resources used by the Service.
	Stop() error
}

// startHandler shall be considered undocumented until further notice.
type startHandler interface {
	OnStart(Container) error
}

// stopHandler shall be considered undocumented until further notice.
type stopHandler interface {
	OnStop(Container)
}

type serviceManager struct {
	Container
	services []Service
}

func newServiceManager(container Container) *serviceManager {
	return &serviceManager{
		container,
		make([]Service, 0, 1),
	}
}

func (manager *serviceManager) AddService(service Service) {
	manager.services = append(manager.services, service)
}

func (manager *serviceManager) Start() error {
	if len(manager.services) <= 0 {
		return errors.New("No services were registered")
	}

	running := &sync.WaitGroup{}
	fail := make(chan error)

	for _, service := range manager.services {
		running.Add(1)
		go func(srv Service) {
			defer running.Done()

			if handler, ok := srv.(startHandler); ok {
				if err := handler.OnStart(manager); err != nil {
					fail <- err
					return
				}
			}

			if err := srv.Start(); err != nil {
				manager.Printf("Error while listening %s\n", err)
				fail <- err
			} else if handler, ok := srv.(stopHandler); ok {
				handler.OnStop(manager)
			}
		}(service)
	}

	done := make(chan bool)
	go func() {
		running.Wait()
		close(done)
	}()

	var err error
	select {
	case <-done:
		// All ok.
	case err = <-fail:
		// At least one has failed.
		close(fail)
	}

	return err
}

func (manager *serviceManager) Stop() error {
	faults := &stopError{}
	for _, service := range manager.services {
		if err := service.Stop(); err != nil {
			faults.AddError(err)
		}
	}
	return faults.AsError()
}

type stopError struct {
	errors []error
}

func (stop *stopError) AddError(err error) {
	stop.errors = append(stop.errors, err)
}

func (stop *stopError) Error() string {
	msgs := make([]string, 0, len(stop.errors))
	for _, err := range stop.errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

func (stop *stopError) AsError() error {
	if len(stop.errors) == 0 {
		return nil
	}
	return stop
}
