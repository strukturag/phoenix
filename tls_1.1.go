// Copyright 2016 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !go1.2

package phoenix

import (
	"crypto/tls"
)

func setTLSMinVersion(config Config, section string, tlsConfig *tls.Config) {
	// NOTE(lcooper): We cannot support this on Go 1.1.
}

func makeDefaultCipherSuites() []uint16 {
	// Go 1.1 is missing the following suites:
	//  ECDHE_RSA_WITH_AES_128_GCM_SHA256
	//  ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	//  ECDHE_RSA_WITH_AES_128_CBC_SHA
	//	ECDHE_ECDSA_WITH_AES_128_CBC_SHA
	//  ECDHE_ECDSA_WITH_AES_256_CBC_SHA
	// Still no RC4 support.
	return []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	}
}
