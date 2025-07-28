package certformat

import (
	"bufio"
	"os"
)

// Format represents the type of certificate encoding.
type Format int

const (
	Unknown Format = iota
	PEM
	DER
)

// CertificateFormat is a namespace-like value to access detection logic.
var CertificateFormat formatDetector

type formatDetector struct{}

// Detect reads the given file path and tries to determine if it's DER or PEM.
func (formatDetector) Detect(path string) Format {
	isBinary, err := isBinaryData(path)
	if err != nil {
		return Unknown
	}
	if isBinary {
		return DER
	}
	return PEM
}

// isBinaryData checks for non-text control characters.
func isBinaryData(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.Size() == 0 {
		return false, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break // EOF or read error
		}
		if (b > 0 && b < 8) || (b > 13 && b < 26) {
			return true, nil
		}
	}
	return false, nil
}

