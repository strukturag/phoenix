// Copyright 2016 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.2

package phoenix

import (
	"crypto/tls"
)

func setTLSMinVersion(config Config, section string, tlsConfig *tls.Config) {
	// Default to SSL3.
	minVersion := tls.VersionSSL30
	minVersionString, err := config.GetString(section, "minVersion")
	if err == nil {
		switch minVersionString {
		case "TLSv1":
			minVersion = tls.VersionTLS10
		case "TLSv1.1":
			minVersion = tls.VersionTLS11
		case "TLSv1.2":
			minVersion = tls.VersionTLS12
		}
	}
	tlsConfig.MinVersion = uint16(minVersion)
}

func makeDefaultCipherSuites() []uint16 {
	// Default cipher suites - no RC4.
	return []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	}
}
