package certinfo

import (
	"crypto/x509"
	//"encoding/asn1"
	"encoding/hex"
	"fmt"
	"log"
	"ssl-tools/internal/x509extras"
	"strings"
)

func LogCertSummary(cert *x509.Certificate, index int) {
	log.SetFlags(0)
	log.Printf("[%d] %s", index, cert.Subject.CommonName)
}

// LogCertInfo prints details of the given certificate
func LogCertInfo(cert *x509.Certificate) {
	log.SetFlags(0)
	log.Println(strings.Repeat("-", 92))
	if cert.Subject.CommonName != "" {
		log.Printf("Simple Name: %s", cert.Subject.CommonName)
	}
	log.Printf("Date: %s - %s", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
	log.Printf("Issuer: %s", cert.Issuer.CommonName)
	log.Printf("Issuer DN: %s", cert.Issuer.String())
	log.Printf("Serial Number: %s", cert.SerialNumber.String())
	log.Printf("Thumbprint (SHA-1): %X", certFingerprintSHA1(cert))

	for _, ext := range cert.Extensions {
		oid := ext.Id.String()
		isCritical := ext.Critical
		extName := getFriendlyName(oid)
		if isCritical {
			extName += " (critical)"
		}

		switch oid {
		case "2.5.29.19": // Basic Constraints
			log.Printf("%s: CA=%v", extName, cert.IsCA)

		case "2.5.29.14": // Subject Key Identifier
			log.Printf("SKID%s: %s", criticalSuffix(isCritical), hex.EncodeToString(cert.SubjectKeyId))

		case "2.5.29.35": // Authority Key Identifier
			log.Printf("AKID%s: %s", criticalSuffix(isCritical), hex.EncodeToString(cert.AuthorityKeyId))

		case "2.5.29.17": // Subject Alternative Name
			log.Printf("%s", extName)
			for _, dns := range cert.DNSNames {
				log.Printf("  DNS: %s", dns)
			}

		case "2.5.29.15": // Key Usage
			log.Printf("%s: %s", extName, keyUsageString(cert.KeyUsage))

		case "2.5.29.37": // Extended Key Usage
			log.Printf("%s: %s", extName, extKeyUsageString(cert.ExtKeyUsage))
		case "1.3.6.1.5.5.7.1.1": // AIA
			aia, err := x509extras.ParseAIA(ext.Value)
			if err == nil {
				log.Printf("%s", getFriendlyName(oid)) // still works
				for _, ad := range aia {
					log.Printf("  %s: %s", x509extras.FriendlyAccessMethod(ad.Method), ad.URI)
				}
			} else {
				log.Printf("%s: failed to parse (%v)", getFriendlyName(oid), err)
			}
		}
	}

	log.Println()
}

func certFingerprintSHA1(cert *x509.Certificate) []byte {
	fp := cert.Signature
	if len(fp) > 20 {
		fp = fp[:20]
	}
	return fp
}

func criticalSuffix(critical bool) string {
	if critical {
		return " (critical)"
	}
	return ""
}

func getFriendlyName(oid string) string {
	switch oid {
	case "2.5.29.19":
		return "Basic Constraints"
	case "2.5.29.14":
		return "Subject Key Identifier"
	case "2.5.29.35":
		return "Authority Key Identifier"
	case "2.5.29.17":
		return "Subject Alternative Name"
	case "2.5.29.15":
		return "Key Usage"
	case "2.5.29.37":
		return "Enhanced Key Usage"
	case "1.3.6.1.5.5.7.1.1":
		return "Authority Information Access"
	default:
		return "Unknown OID: " + oid
	}
}

func keyUsageString(ku x509.KeyUsage) string {
	var usages []string
	if ku&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if ku&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if ku&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if ku&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if ku&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if ku&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if ku&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if ku&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if ku&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}
	return strings.Join(usages, ", ")
}

func extKeyUsageString(usages []x509.ExtKeyUsage) string {
	var result []string
	for _, usage := range usages {
		switch usage {
		case x509.ExtKeyUsageServerAuth:
			result = append(result, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			result = append(result, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			result = append(result, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			result = append(result, "EmailProtection")
		case x509.ExtKeyUsageTimeStamping:
			result = append(result, "TimeStamping")
		default:
			result = append(result, fmt.Sprintf("Unknown (%d)", usage))
		}
	}
	return strings.Join(result, ", ")
}
