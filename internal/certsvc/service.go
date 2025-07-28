package certsvc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"slices"
	"software.sslmate.com/src/go-pkcs12"
	"ssl-tools/internal/certformat"
	"ssl-tools/internal/certinfo"
	"ssl-tools/internal/models"
	"ssl-tools/internal/options"
	"strings"
)

type CertificateService struct {
	keySize    int
	keyCreated bool
}

func New() *CertificateService {
	return &CertificateService{keySize: 4096, keyCreated: false}
}

func (c *CertificateService) SetKeyLength(bits int) error {
	if bits != 2048 && bits != 3072 && bits != 4096 {
		return fmt.Errorf("unsupported key size %d", bits)
	}
	c.keySize = bits
	return nil
}

func (c *CertificateService) CreateCSR(opts options.CreateOptions) (*models.CSRdto, error) {
	var privateKey *rsa.PrivateKey
	var err error

	if opts.Key != "" {
		privateKey, err = loadPrivateKeyFromFile(opts.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key at %w", err)
		}
	} else {
		if opts.KeySize == 0 {
			opts.KeySize = c.keySize // use the default key size if none given
		}
		privateKey, err = rsa.GenerateKey(rand.Reader, opts.KeySize)
		c.keyCreated = true
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}
	}

	// Create subject
	subj := pkix.Name{
		CommonName:         opts.CommonName,
		Organization:       []string{"San Diego Community College District"},
		OrganizationalUnit: []string{"It"},
		Country:            []string{"US"},
		Province:           []string{"California"},
		Locality:           []string{"San Diego"},
	}

	// Subject alternative names
	sanList := append([]string{opts.CommonName}, opts.SANs...)

	// Create CSR template
	template := x509.CertificateRequest{
		Subject:            subj,
		DNSNames:           sanList,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	// Create the CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating CSR: %v", err)
	}

	// Encode CSR to PEM
	csrPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	// Encode private key to PEM
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Return the result
	return &models.CSRdto{
		RequestData: strings.TrimSpace(string(csrPem)),
		KeyData:     string(keyPem),
		Label:       strings.ReplaceAll(opts.CommonName, "*", "_"),
	}, nil
}

func loadPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read key file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block in key file")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (c *CertificateService) SaveCSRdto(dto *models.CSRdto) error {
	// first save CSR data to file
	csrFile := fmt.Sprintf("%s.csr", dto.Label)
	f, err := os.Create(csrFile)
	check(err)
	defer f.Close()
	byteCount, err := f.Write([]byte(dto.RequestData))
	check(err)
	fmt.Printf("wrote %d bytes to %s\n", byteCount, csrFile)

	// then save key but only if the key was created by service
	if c.keyCreated {
		//write it
		keyFile := fmt.Sprintf("%s.key", dto.Label)
		k, err := os.Create(keyFile)
		check(err)
		defer k.Close()
		keyByteCount, err := k.Write([]byte(dto.KeyData))
		check(err)
		fmt.Printf("wrote %d bytes to %s\n", keyByteCount, keyFile)
	}
	return nil
}

func (c *CertificateService) SavePFXdto(dto *models.PFXdto) error {
	var pfxFile string
	if dto.FileName == "" {
		pfxFile = fmt.Sprintf("%s.pfx", dto.CommonName)
	} else {
		pfxFile = dto.FileName
	}
	f, err := os.Create(pfxFile)
	check(err)
	defer f.Close()
	byteCount, err := f.Write(dto.CertificateData)
	check(err)
	fmt.Printf("wrote %d bytes to %s\n", byteCount, pfxFile)

	return nil
}

func (c *CertificateService) FinishCSR(opts options.FinishOptions) (*models.PFXdto, error) {
	dto := &models.PFXdto{}

	// Step 1: Read certificate file
	certBytes, err := os.ReadFile(opts.Certificate)
	if err != nil {
		dto.CreateMessage = fmt.Sprintf("Failed to read certificate file: %v", err)
		dto.OpCode = 1001
		return dto, nil
	}

	// Step 2: Autofix could be inserted here
	// certBytes = Autofix(certBytes)

	// Step 3: Read key file
	keyBytes, err := os.ReadFile(opts.Key)
	if err != nil {
		dto.CreateMessage = fmt.Sprintf("Failed to read key file: %v", err)
		dto.OpCode = 1002
		return dto, nil
	}

	// Step 4: Parse certificates from PEM
	certs, err := parsePEMCerts(certBytes)
	if err != nil || len(certs) == 0 {
		dto.CreateMessage = "No valid certificates found"
		dto.OpCode = 1003
		return dto, nil
	}

	// Step 5: Identify end-entity certificate
	endEntity := findEndEntityCert(certs)
	if endEntity == nil {
		dto.CreateMessage = "No unique end-entity certificate found"
		dto.OpCode = 1004
		return dto, nil
	}
	// Step 5.5: Save subject in dto
	dto.CommonName = endEntity.Subject.CommonName

	// Step 6: Parse private key
	privKey, err := parseRSAPrivateKey(keyBytes)
	if err != nil {
		dto.CreateMessage = fmt.Sprintf("Failed to parse private key: %v", err)
		dto.OpCode = 1005
		return dto, nil
	}

	// Step 7: Create a certificate with private key
	/* go certificates do not have private keys
	endEntityWithKey, err := endEntity.PrivateKey(privKey)
	if err != nil {
		dto.CreateMessage = fmt.Sprintf("Failed to attach private key: %v", err)
		dto.OpCode = 1006
		return dto, nil
	}
	*/

	// Step 8: Build certificate chain
	chain := []*x509.Certificate{endEntity}
	// Only include a chain if the option is set
	if opts.Chain {
		if !opts.IncludeRoot {
			// if we are not including root certs then filter them out
			certs = filterTrustedCerts(certs)
		}
		for _, cert := range certs {
			if opts.IncludeRoot || !isSelfSigned(cert) {
				if !cert.Equal(endEntity) {
					//chain = append(chain, cert)
					chain = append([]*x509.Certificate{cert}, chain...)
				}
			}
		}
	}

	// Step 9: Encode to PFX
	pfxData, err := encodeToPFX(chain, privKey, opts.Password)
	//fmt.Printf("Using password of [%s]\n", opts.Password)
	if err != nil {
		dto.CreateMessage = fmt.Sprintf("Failed to create PFX: %v", err)
		dto.OpCode = 1007
		return dto, nil
	}

	dto.CertificateData = pfxData
	dto.CreateMessage = "PFX created successfully"
	return dto, nil
}

// helper functions
func filterTrustedCerts(chain []*x509.Certificate) []*x509.Certificate {
	var result []*x509.Certificate
	for _, cert := range chain {
		// only add non-root certs
		if !isTrusted(cert) {
			result = append(result, cert)
		}
	}
	return result
}

func isTrusted(cert *x509.Certificate) bool {
	roots, err := x509.SystemCertPool()
	if err != nil {
		return false
	}

	intermediates := x509.NewCertPool() // empty
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	_, err = cert.Verify(opts)
	return err == nil
}

func parsePEMCerts(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				certs = append(certs, cert)
			}
		}
	}
	return certs, nil
}

func parseRSAPrivateKey(keyData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid PEM RSA private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func findEndEntityCert(certs []*x509.Certificate) *x509.Certificate {
	for _, cert := range certs {
		isIssuer := false
		for _, other := range certs {
			if cert.Subject.String() == other.Issuer.String() && !cert.Equal(other) {
				isIssuer = true
				break
			}
		}
		if !isIssuer {
			return cert
		}
	}
	return nil
}

func isSelfSigned(cert *x509.Certificate) bool {
	return cert.Issuer.String() == cert.Subject.String()
}

func encodeToPFX(certs []*x509.Certificate, key *rsa.PrivateKey, password string) ([]byte, error) {
	// This requires Go's x/crypto/pkcs12 package
	// go get golang.org/x/crypto/pkcs12
	// the above line uses the 'legacy' encoder and is considered 'unsafe' (not sure why it is presented like it is the default)
	//return pkcs12.Encode(rand.Reader, key, certs[0], certs[1:], password)
	if len(certs) == 1 {
		return pkcs12.Modern2023.Encode(key, certs[0], nil, password)
	} else {
		return pkcs12.Modern2023.Encode(key, certs[0], certs[1:], password)
	}
}

/**
	First build the helpers for info.
	1: isBinary (done, using the certformat.CertificateFormat type)
	2: loadBinaryCert(path) and loadPemCert(path)
	3: some type of utility to get extensions and/or ANSI values from certificates.
	   look at the C# PrintCertInfo method at line 943 for all the data I need
**/

func loadPemCertsFromFile(path string) ([]*x509.Certificate, error) {
	certBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Step 4: Parse certificates from PEM
	certs, err := parsePEMCerts(certBytes)
	return certs, nil

}
func loadBinaryCertsFromFile(path string, pass string) ([]*x509.Certificate, error) {
	certBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	_, cert, chain, err := pkcs12.DecodeChain(certBytes, pass)
	if err != nil {
		return nil, err
	}
	certs := []*x509.Certificate{cert}
	for _, link := range chain {
		certs = append(certs, link)
	}
	// reverse the order since pkcs12.DecodeChain gives us a reversed list
	slices.Reverse(certs)
	return certs, nil
}

// info section
func (c *CertificateService) GetInfo(opts options.InfoOptions) error {
	if len(opts.Certificates) > 0 {
		for _, path := range opts.Certificates {
			format := certformat.CertificateFormat.Detect(path)
			switch format {
			case certformat.DER:
				fmt.Println("Binary certificate file found: ", path)
				certs, err := loadBinaryCertsFromFile(path, opts.Password)
				if err == nil {
					if !opts.ShortSummary {
						for _, cert := range certs {
							certinfo.LogCertInfo(cert)
						}
					}
					fmt.Println(strings.Repeat("-", 92))
					fmt.Println("Chain summary")
					for i, cert := range certs {
						certinfo.LogCertSummary(cert, i)
					}
				} else {
					fmt.Println(fmt.Errorf("reading certificates failed: %w", err))
					return err
				}
			case certformat.PEM:
				fmt.Println("PEM certificate file found: ", path)
				certs, err := loadPemCertsFromFile(path)
				if err == nil {
					if !opts.ShortSummary {
						for _, cert := range certs {
							certinfo.LogCertInfo(cert)
						}
					}
					fmt.Println(strings.Repeat("-", 92))
					fmt.Println("Chain summary")
					for i, cert := range certs {
						certinfo.LogCertSummary(cert,i)
					}
				} else {
					fmt.Println(fmt.Errorf("reading certificates failed: %W", err))
					return err
				}
			}
		}
	}
	return nil
}
