package integration_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/cloudfoundry/mysql-diag/config"
	"github.com/cloudfoundry/mysql-diag/internal/testing/docker"
)

// tlsTestServerName is the DNS SAN on the server cert. It differs from the TCP address
// (127.0.0.1) so that verify_identity must use the configured ServerName — not the IP —
// to pass x509 hostname verification.
const tlsTestServerName = "mysql-tls-test.internal"

var _ = Describe("mysql-diag TLS config", Ordered, Label("integration"), func() {
	var (
		containerID string
		mysqlPort   uint64
		caPEM       []byte
	)

	BeforeAll(func() {
		var certPEM, keyPEM []byte
		var err error
		caPEM, certPEM, keyPEM, err = generateTLSCerts(tlsTestServerName)
		Expect(err).NotTo(HaveOccurred())

		certDir, err := os.MkdirTemp("", "mysql-tls-certs-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(certDir) })

		for name, data := range map[string][]byte{
			"ca.pem":          caPEM,
			"server-cert.pem": certPEM,
			"server-key.pem":  keyPEM,
		} {
			Expect(os.WriteFile(filepath.Join(certDir, name), data, 0644)).To(Succeed())
		}

		containerID, err = docker.CreateContainer(docker.ContainerSpec{
			Image:          "percona/percona-server:8.0",
			Ports:          []string{"3306/tcp"},
			HealthCmd:      "mysqladmin ping --host=127.0.0.1",
			HealthInterval: "3s",
			Env:            []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1"},
			Volumes:        []string{certDir + ":/etc/mysql/ssl"},
			Args: []string{
				"--ssl-ca=/etc/mysql/ssl/ca.pem",
				"--ssl-cert=/etc/mysql/ssl/server-cert.pem",
				"--ssl-key=/etc/mysql/ssl/server-key.pem",
				"--require-secure-transport=ON",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(docker.RemoveContainer(containerID)).To(Succeed()) })

		Expect(docker.StartContainer(containerID)).To(Succeed())

		containerPort, err := docker.ContainerPort(containerID, "3306/tcp")
		Expect(err).NotTo(HaveOccurred())

		mysqlPort, err = strconv.ParseUint(containerPort, 10, 16)
		Expect(err).NotTo(HaveOccurred())

		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caPEM)
		Expect(mysql.RegisterTLSConfig("healthcheck", &tls.Config{
			RootCAs:    caPool,
			ServerName: tlsTestServerName,
		})).To(Succeed())
		healthDB, err := sql.Open("mysql", fmt.Sprintf("root@tcp(127.0.0.1:%d)/?tls=healthcheck", mysqlPort))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(healthDB.Close)

		Eventually(healthDB.Ping, "3m", "2s").Should(Succeed(),
			"timed out waiting for MySQL to accept connections")
	})

	// writeConfig marshals a minimal mysql-diag config file pointing at the running container.
	writeConfig := func(ca, serverName string) string {
		cfg := config.MysqlConfig{
			Username:   "root",
			Password:   "",
			Port:       uint(mysqlPort),
			CA:         ca,
			ServerName: serverName,
			Nodes:      []config.MysqlNode{{Host: "127.0.0.1", Name: "mysql", UUID: "test"}},
		}

		data, err := yaml.Marshal(map[string]config.MysqlConfig{"mysql": cfg})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		dir, err := os.MkdirTemp("", "mysql-diag-cfg-*")
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(dir) })

		path := filepath.Join(dir, "config.yml")
		ExpectWithOffset(1, os.WriteFile(path, data, 0644)).To(Succeed())
		return path
	}

	It("connects in required mode — TLS encrypted, no certificate verification", func() {
		cfg, err := config.LoadFromFile(writeConfig("", ""))
		Expect(err).NotTo(HaveOccurred())

		db := cfg.Mysql.Connection(cfg.Mysql.Nodes[0])
		DeferCleanup(db.Close)
		Expect(db.Ping()).To(Succeed())
	})

	It("connects in verify_identity mode — verifies CA chain AND hostname", func() {
		cfg, err := config.LoadFromFile(writeConfig(string(caPEM), tlsTestServerName))
		Expect(err).NotTo(HaveOccurred())

		db := cfg.Mysql.Connection(cfg.Mysql.Nodes[0])
		DeferCleanup(db.Close)
		Expect(db.Ping()).To(Succeed())
	})
})

// generateTLSCerts produces a self-signed CA and a server certificate bearing dnsName as
// the only DNS SAN. The server cert has no IP SANs, so TLS hostname verification against
// an IP address (127.0.0.1) will only succeed when the client explicitly sets ServerName
// to dnsName.
func generateTLSCerts(dnsName string) (caPEM, certPEM, keyPEM []byte, err error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	if err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: caDER}); err != nil {
		return
	}
	caPEM = bytes.Clone(buf.Bytes())
	buf.Reset()

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: dnsName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		DNSNames:     []string{dnsName},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return
	}
	if err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: serverDER}); err != nil {
		return
	}
	certPEM = bytes.Clone(buf.Bytes())
	buf.Reset()

	keyDER, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		return
	}
	err = pem.Encode(&buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyPEM = buf.Bytes()
	return
}
