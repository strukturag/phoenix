package phoenix

import (
	"crypto/tls"
)

func loadTLSConfig(config Config, section string) (*tls.Config, error) {
	certFile, err := config.GetString(section, "certificate")
	if err != nil {
		return nil, err
	}

	keyFile, err := config.GetString(section, "key")
	if err != nil {
		return nil, err
	}

	certificates := make([]tls.Certificate, 1)
	certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create TLS config.
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CipherSuites:             makeDefaultCipherSuites(),
		Certificates:             certificates,
	}
	setTLSMinVersion(config, "https", tlsConfig)
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}
