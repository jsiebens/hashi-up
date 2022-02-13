package cmd

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

	"github.com/muesli/coral"
)

func TlsCommands() *coral.Command {
	tlsCmd := baseCommand("tls")
	tlsCmd.Short = "Builtin helpers for creating certificates"
	tlsCmd.Long = "Builtin helpers for creating certificates"
	tlsCmd.AddCommand(certCommands())
	return tlsCmd
}

func certCommands() *coral.Command {
	certCmd := baseCommand("cert")
	certCmd.AddCommand(createCertificateCommand())
	return certCmd
}

func createCertificateCommand() *coral.Command {

	var hosts []string

	command := &coral.Command{
		Use:          "create",
		SilenceUsage: true,
	}

	command.Flags().StringSliceVar(&hosts, "host", []string{}, "Hostnames and IPs to generate a certificate for")

	command.RunE = func(command *coral.Command, args []string) error {

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

		if err != nil {
			return err
		}

		keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment

		notBefore := time.Now()
		notAfter := notBefore.Add(365 * 24 * time.Hour)

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			return err
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			NotBefore:    notBefore,
			NotAfter:     notAfter,

			Subject: pkix.Name{
				CommonName: "hashi-up!",
			},

			KeyUsage:              keyUsage,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		for _, h := range hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
		if err != nil {
			return err
		}

		certOut, err := os.Create("server.pem")
		if err != nil {
			return err
		}
		if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
			return err
		}
		if err := certOut.Close(); err != nil {
			return err
		}
		fmt.Println("wrote server.pem")

		keyOut, err := os.OpenFile("server-key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return err
		}
		if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
			return err
		}
		if err := keyOut.Close(); err != nil {
			return err
		}

		fmt.Println("wrote server-key.pem")

		return nil
	}

	return command
}
