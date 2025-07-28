package x509extras

import (
	"crypto/x509"
	"errors"
)

// SortCertificateChain tries to sort a slice of x509 certificates from leaf to root.
// It uses the Subject/Issuer relationship to build the correct order.
// It returns an error if the chain cannot be fully constructed.
func SortCertificateChain(certs []*x509.Certificate) ([]*x509.Certificate, error) {
	if len(certs) == 0 {
		return nil, nil
	}

	// Build a map from Subject → cert
	subjectMap := make(map[string]*x509.Certificate)
	for _, c := range certs {
		subjectMap[c.Subject.String()] = c
	}

	// Find the leaf cert (the one whose Issuer is not in the list of subjects or self-signed)
	var leaf *x509.Certificate
	for _, c := range certs {
		if c.Issuer.String() == c.Subject.String() {
			continue // skip self-signed roots
		}
		if _, ok := subjectMap[c.Issuer.String()]; !ok {
			leaf = c
			break
		}
	}
	if leaf == nil {
		return nil, errors.New("could not identify leaf certificate")
	}

	// Walk the chain
	var sorted []*x509.Certificate
	current := leaf
	sorted = append(sorted, current)
	for {
		if current.Issuer.String() == current.Subject.String() {
			break // reached self-signed root
		}
		parent, ok := subjectMap[current.Issuer.String()]
		if !ok {
			return nil, errors.New("incomplete chain: missing issuer for " + current.Subject.String())
		}
		sorted = append(sorted, parent)
		current = parent
	}

	return sorted, nil
}

