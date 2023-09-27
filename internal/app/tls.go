package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	serialNumber = 1
	ip4GrayZone  = 127
	yearsGrant   = 1
	RSALen       = 4096
	CertsPerm    = 0600
)

func CreateCertificates(logger *zap.SugaredLogger) error {
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(serialNumber),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Shortener"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(ip4GrayZone, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(yearsGrant, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, RSALen)
	if err != nil {
		logger.Fatal(err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		logger.Fatal(err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	certFile, err := os.OpenFile("./certs/cert.pem", os.O_WRONLY|os.O_CREATE, CertsPerm)
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		if err := certFile.Close(); err != nil {
			logger.Fatal(err)
		}
	}()

	if err := pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return fmt.Errorf("error creating cert file: %w", err)
	}

	rsaFile, err := os.OpenFile("./certs/private.pem", os.O_WRONLY|os.O_CREATE, CertsPerm)
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		if err := rsaFile.Close(); err != nil {
			logger.Fatal(err)
		}
	}()

	if err := pem.Encode(rsaFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("error creating RSA private key: %w", err)
	}

	return nil
}
