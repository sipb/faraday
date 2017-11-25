package testkeyutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"
)

func GenerateTLSKeypairForTests(t *testing.T, commonname string, dns []string, ips []net.IP, parent *x509.Certificate, parentkey *rsa.PrivateKey) (*rsa.PrivateKey, *x509.Certificate) {
	return GenerateTLSKeypairForTests_WithTime(t, commonname, dns, ips, parent, parentkey, time.Now(), time.Hour)
}

func GenerateTLSKeypairForTests_WithTime(t *testing.T, commonname string, dns []string, ips []net.IP, parent *x509.Certificate, parentkey *rsa.PrivateKey, issueat time.Time, duration time.Duration) (*rsa.PrivateKey, *x509.Certificate) {
	key, err := rsa.GenerateKey(rand.Reader, 512) // NOTE: this is LAUGHABLY SMALL! do not attempt to use this in production.
	if err != nil {
		t.Fatal("Could not generate TLS keypair: " + err.Error())
	}

	serialNumber, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		t.Fatal("Could not generate TLS keypair: " + err.Error())
	}

	extKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

	certTemplate := &x509.Certificate{
		SignatureAlgorithm: x509.SHA256WithRSA,

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: extKeyUsage,

		BasicConstraintsValid: true,
		IsCA:       true,
		MaxPathLen: 1,

		SerialNumber: serialNumber,

		NotBefore: issueat,
		NotAfter:  issueat.Add(duration),

		Subject:     pkix.Name{CommonName: commonname},
		DNSNames:    dns,
		IPAddresses: ips,
	}

	if parent == nil {
		parent = certTemplate
		parentkey = key
	}

	signed_cert, err := x509.CreateCertificate(rand.Reader, certTemplate, parent, key.Public(), parentkey)
	if err != nil {
		t.Fatal("Could not generate TLS keypair: " + err.Error())
	}
	cert, err := x509.ParseCertificate(signed_cert)
	if err != nil {
		t.Fatal("Could not generate TLS keypair: " + err.Error())
	}
	return key, cert
}
