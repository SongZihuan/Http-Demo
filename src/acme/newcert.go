package acme

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"fmt"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"net"
	"os"
	"path"
	"time"
)

func saveAccount(dir string, email string, reg *registration.Resource) error {
	filepath := path.Join(dir, fmt.Sprintf("%s.account", email))

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(reg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, buff.Bytes(), 0644)
}

func loadAccount(dir string, email string) (*registration.Resource, error) {
	filepath := path.Join(dir, fmt.Sprintf("%s.account", email))
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var reg registration.Resource
	dec := gob.NewDecoder(file)

	err = dec.Decode(&reg)
	if err != nil {
		return nil, err
	}

	return &reg, nil
}

func newCert(dir string, email string, httpsAddress string, domain string) (crypto.PrivateKey, *certificate.Resource, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	user := NewUser(email, privateKey)

	config := lego.NewConfig(user)
	config.Certificate.KeyType = certcrypto.RSA4096
	config.Certificate.Timeout = 30 * 24 * time.Hour
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	iface, port, err := net.SplitHostPort(httpsAddress)
	if err != nil {
		return nil, nil, err
	}

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer(domain, port))
	if err != nil {
		return nil, nil, err
	}

	regOption := registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	}

	reg, err := loadAccount(path.Join(dir, "account"), email)
	if err != nil {
		// 尝试注册
		reg, err = client.Registration.Register(regOption)
		if err != nil {
			return nil, nil, err
		}

		err = saveAccount(dir, email, reg)
		if err != nil {
			return nil, nil, err
		}
	}
	user.setRegistration(reg)

	if domain == "" {
		domain = iface
	}

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, certificates, nil
}

func getCert(resource *certificate.Resource) (*x509.Certificate, error) {
	block, _ := pem.Decode(resource.Certificate)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}

func writerWithDate(baseDir string, resource *certificate.Resource) error {
	cert, err := getCert(resource)
	if err != nil {
		return err
	}

	domain := cert.Subject.CommonName
	if domain == "" && len(cert.DNSNames) == 0 {
		return fmt.Errorf("no domains in certificate")
	}
	domain = cert.DNSNames[0]

	year := fmt.Sprintf("%d", cert.NotBefore.Year())
	month := fmt.Sprintf("%d", cert.NotBefore.Month())
	day := fmt.Sprintf("%d", cert.NotBefore.Day())

	dir := path.Join(baseDir, domain, year, month, day)

	err = os.WriteFile(path.Join(dir, FilePrivateKey), resource.PrivateKey, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dir, FileCertificate), resource.Certificate, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dir, FileIssuerCertificate), resource.IssuerCertificate, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dir, FileCSR), resource.CSR, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func writer(dir string, resource *certificate.Resource) error {
	err := os.WriteFile(path.Join(dir, FilePrivateKey), resource.PrivateKey, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dir, FileCertificate), resource.Certificate, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dir, FileIssuerCertificate), resource.IssuerCertificate, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dir, FileCSR), resource.CSR, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
