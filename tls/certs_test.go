package tls

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"
)

// TestGenerateCACert tests the GenerateCACert function.
// write a test for GenerateCACert
func TestGenerateCACert(t *testing.T) {
	// generate a new CA certificate
	certPEM, keyPEM, err := GenerateCACert(pkix.Name{CommonName: "rqlite.io"}, 0, time.Hour, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// decode the certificate and private key
	cert, _ := pem.Decode(certPEM)
	if cert == nil {
		t.Fatal("failed to decode certificate")
	}

	key, _ := pem.Decode(keyPEM)
	if err != nil {
		t.Fatal("failed to decode key")
	}

	// parse the certificate and private key
	certParsed, err := x509.ParseCertificate(cert.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	keyParsed, err := x509.ParsePKCS1PrivateKey(key.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	// verify the certificate and private key
	if certParsed.Subject.CommonName != "rqlite.io" {
		t.Fatal("certificate subject is not correct")
	}

	if !certParsed.IsCA {
		t.Fatal("certificate is not a CA")
	}

	if certParsed.PublicKey.(*rsa.PublicKey).N.Cmp(keyParsed.N) != 0 {
		t.Fatal("certificate and private key do not match")
	}
}

func TestGenerateCASignedCert(t *testing.T) {
	caCert, caKey := mustGenerateCACert(pkix.Name{CommonName: "ca.rqlite"})

	// generate a new certificate signed by the CA
	certPEM, keyPEM, err := GenerateCert(pkix.Name{CommonName: "test"}, 365*24*time.Hour, 2048, caCert, caKey)
	if err != nil {
		t.Fatal(err)
	}

	// write certPEM and keyPEM to files
	ioutil.WriteFile("cert.pem", certPEM, 0644)

	cert, _ := pem.Decode(certPEM)
	if cert == nil {
		panic("failed to decode certificate")
	}

	key, _ := pem.Decode(keyPEM)
	if key == nil {
		panic("failed to decode key")
	}

	// parse the certificate and private key
	parsedCert, err := x509.ParseCertificate(cert.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x509.ParsePKCS1PrivateKey(key.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	// verify the certificate is signed by the CA
	if err := parsedCert.CheckSignatureFrom(caCert); err != nil {
		t.Fatal(err)
	}

	// verify the certificate is valid for the correct duration
	if parsedCert.NotBefore.After(time.Now()) {
		t.Fatal("certificate is not valid yet")
	}
	if parsedCert.NotAfter.Before(time.Now()) {
		t.Fatal("certificate is expired")
	}

	// verify the certificate is valid for the correct subject
	if parsedCert.Subject.CommonName != "test" {
		t.Fatal("certificate has incorrect subject")
	}

	// verify the certificate is valid for the correct key usage
	if parsedCert.KeyUsage != (x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature) {
		t.Fatalf("certificate has incorrect key usage, exp %v, got %v", x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature, parsedCert.KeyUsage)
	}

	// verify the certificate is valid for the correct extended key usage
	if len(parsedCert.ExtKeyUsage) != 1 || parsedCert.ExtKeyUsage[0] != x509.ExtKeyUsageServerAuth {
		t.Fatal("certificate has incorrect extended key usage")
	}

	// verify the certificate is valid for the correct basic constraints
	if parsedCert.IsCA {
		t.Fatal("certificate has incorrect basic constraints")
	}
}

// mustGenerateCACert generates a new CA certificate and private key. It is used for testing only.
func mustGenerateCACert(name pkix.Name) (*x509.Certificate, *rsa.PrivateKey) {
	certPEM, keyPEM, err := GenerateCACert(name, 0, time.Hour, 2048)
	if err != nil {
		panic(err)
	}
	cert, _ := pem.Decode(certPEM)
	if cert == nil {
		panic("failed to decode certificate")
	}

	key, _ := pem.Decode(keyPEM)
	if key == nil {
		panic("failed to decode key")
	}

	parsedCert, err := x509.ParseCertificate(cert.Bytes)
	if err != nil {
		panic(err)
	}
	parsedKey, err := x509.ParsePKCS1PrivateKey(key.Bytes)
	if err != nil {
		panic(err)
	}

	return parsedCert, parsedKey
}
