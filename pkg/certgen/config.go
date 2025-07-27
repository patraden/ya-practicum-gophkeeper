// Copyright (c) 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package certgen

import (
	"fmt"
	"time"
)

// Config defines input parameters for generating a certificate.
//
// Fields like Host, OrgName, and CommonName are used in the certificate's
// Subject and SANs. The configuration also allows specifying usage
// (client/server/CA) and key algorithm preferences.
type Config struct {
	Host       string        // Comma-separated list of DNS names, IPs, emails, or URIs for SANs.
	OrgName    string        // Organization name for the certificate subject.
	CommonName string        // Optional CommonName, used for client and CA certs.
	IsClient   bool          // If true, enables client authentication usage.
	IsCA       bool          // If true, generates a certificate authority (CA).
	ValidFrom  time.Time     // Start date of certificate validity.
	ValidFor   time.Duration // Duration for which the certificate is valid.
	Ed25519    bool          // If true, use Ed25519 instead of ECDSA.
	ECDSACurve string        // Curve name: P224, P256, P384, or P521.
	CACertPath string        // Optional: path to CA cert for signing.
	CAKeyPath  string        // Optional: path to CA key for signing.
}

type TimeFlag struct {
	T time.Time
}

func (tf *TimeFlag) Set(value string) error {
	t, err := time.Parse("Jan 2 15:04:05 2006", value)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	tf.T = t

	return nil
}

func (tf *TimeFlag) String() string {
	return tf.T.Format("Jan 2 15:04:05 2006")
}
