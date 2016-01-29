package phoenix

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/strukturag/httputils"
)

type httpService struct {
	*httputils.Server
}

func newHTTPService(logger *log.Logger, handler http.Handler, addr string, readtimeout, writetimeout int, tlsConfig *tls.Config) Service {
	server := &httputils.Server{
		Server: http.Server{
			Addr:           addr,
			Handler:        handler,
			ReadTimeout:    time.Duration(readtimeout) * time.Second,
			WriteTimeout:   time.Duration(writetimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
			TLSConfig:      tlsConfig,
		},
		Logger: logger,
	}
	return &httpService{server}
}

func (service *httpService) OnStart(container Container) (err error) {
	container.Printf("Starting %s server on %s", service.protocol(), service.addr())

	if service.TLSConfig == nil {
		err = service.Listen()
	} else {
		err = service.ListenTLSWithConfig(service.TLSConfig)
	}
	return
}

func (service *httpService) OnStop(container Container) {
	container.Printf("Stopped %s server on %s", service.protocol(), service.addr())
}

func (service *httpService) addr() string {
	return service.Server.Server.Addr
}

func (service *httpService) protocol() string {
	if service.Server.Server.TLSConfig != nil {
		return "HTTPS"
	}
	return "HTTP"
}
