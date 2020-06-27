package gather

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cloudfoundry/mysql-metrics/agent"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DatabaseClient
type DatabaseClient interface {
	ShowGlobalStatus() (map[string]string, error)
	ShowGlobalVariables() (map[string]string, error)
	ShowSlaveStatus() (map[string]string, error)
	HeartbeatStatus() (map[string]string, error)
	ServicePlansDiskAllocated() (map[string]string, error)
	IsAvailable() bool
	IsFollower() (bool, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Stater
type Stater interface {
	Stats(path string) (bytesFree, bytesTotal, inodesFree, inodesTotal uint64, err error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CpuStater
type CpuStater interface {
	GetPercentage() (int, error)
}

type Gatherer struct {
	client          DatabaseClient
	stater          Stater
	cpuStater       CpuStater
	previousQueries int
}

func NewGatherer(client DatabaseClient, stater Stater, cpuStater CpuStater) *Gatherer {
	return &Gatherer{
		client:          client,
		stater:          stater,
		cpuStater:       cpuStater,
		previousQueries: -1,
	}
}

func (g Gatherer) BrokerStats() (map[string]string, error) {
	return g.client.ServicePlansDiskAllocated()
}

func (g Gatherer) CPUStats() (map[string]string, error) {
	percentage, err := g.cpuStater.GetPercentage()
	if err != nil {
		return nil, err
	}
	return map[string]string{"cpu_utilization_percent": strconv.Itoa(percentage)}, err
}
func (g Gatherer) DiskStats() (map[string]string, error) {
	bytesFreePersistent, bytesTotalPersistent, inodesFreePersistent, inodesTotalPersistent, err := g.stater.Stats("/var/vcap/store")
	if err != nil {
		return nil, err
	}

	bytesFreeEphemeral, bytesTotalEphemeral, inodesFreeEphemeral, inodesTotalEphemeral, err := g.stater.Stats("/var/vcap/data")
	if err != nil {
		return nil, err
	}

	persistentDiskUsedBytes := bytesTotalPersistent - bytesFreePersistent
	ephemeralDiskUsedBytes := bytesTotalEphemeral - bytesFreeEphemeral
	persistentInodesUsed := inodesTotalPersistent - inodesFreePersistent
	ephemeralInodesUsed := inodesTotalEphemeral - inodesFreeEphemeral
	return map[string]string{
		"persistent_disk_used":                strconv.FormatUint(persistentDiskUsedBytes/1024, 10),
		"persistent_disk_free":                strconv.FormatUint(bytesFreePersistent/1024, 10),
		"persistent_disk_used_percent":        strconv.FormatUint(g.calculateWholePercent(persistentDiskUsedBytes, bytesTotalPersistent), 10),
		"persistent_disk_inodes_used":         strconv.FormatUint(persistentInodesUsed, 10),
		"persistent_disk_inodes_free":         strconv.FormatUint(inodesFreePersistent, 10),
		"persistent_disk_inodes_used_percent": strconv.FormatUint(g.calculateWholePercent(persistentInodesUsed, inodesTotalPersistent), 10),
		"ephemeral_disk_used":                 strconv.FormatUint(ephemeralDiskUsedBytes/1024, 10),
		"ephemeral_disk_free":                 strconv.FormatUint(bytesFreeEphemeral/1024, 10),
		"ephemeral_disk_used_percent":         strconv.FormatUint(g.calculateWholePercent(ephemeralDiskUsedBytes, bytesTotalEphemeral), 10),
		"ephemeral_disk_inodes_used":          strconv.FormatUint(ephemeralInodesUsed, 10),
		"ephemeral_disk_inodes_free":          strconv.FormatUint(inodesFreeEphemeral, 10),
		"ephemeral_disk_inodes_used_percent":  strconv.FormatUint(g.calculateWholePercent(ephemeralInodesUsed, inodesTotalEphemeral), 10),
	}, nil
}

func (g Gatherer) BackupStats() (map[string]string, error) {
	clientPair, err := tls.X509KeyPair([]byte(`-----BEGIN CERTIFICATE-----
MIIDIDCCAgigAwIBAgIUZJCYM4PWTCW1x1jHTlxpMfm9350wDQYJKoZIhvcNAQEL
BQAwEDEOMAwGA1UEAxMFdWFhQ0EwHhcNMjAwNjI1MTMxNTU2WhcNMjEwNjI1MTMx
NTU2WjATMREwDwYDVQQDEwhhZGJyLWFwaTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBALY+zaNHCP+shB9Os0aKvMydZ8rhppPLo6cUn4JSXEt5HQmVlafk
JXGaSaliFxg9nxDuml6URoT1r7tt09KLxtMvPkF1xvMNgBK/1/WS4/NUKrjmmo9r
zTLLN7bTSqrh5Gl1nRLNnF03WXkf/iXw+IOKXEtX1g/qBgE2Snt768/OP9Uyr9Cd
l2mwtFNOxrSuZCtW50GFixXymHWl0q8aAvI5BUthxEqNOVARBrpxjG2aqDtocg1I
TyWvAVHs2Jfd0eiZ+73xRvWNwcRjaeEyiKp7isYTlm/zGvWPlAYtMBOSD7bXCEd5
ha6ySCqo/nmLqkdAyeX9Xy9mHCo927O4JNcCAwEAAaNvMG0wHQYDVR0OBBYEFAE5
Gu3KlQN4NNdNak25eLYhf9+rMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcD
ATAfBgNVHSMEGDAWgBSnYUI04LGRK2xXFw1u6gMRTCz8RDAMBgNVHRMBAf8EAjAA
MA0GCSqGSIb3DQEBCwUAA4IBAQAAg5Nw+vUAmqyKjHIwNpQhpBrJXjTDu429sBhN
U21jw5yu8xxgREJquQOMXwcVdnRJkd7J+RY487b7uT9eNvNSMYtHjmJ40rcwJ540
od5MuU6LNu4DHdxwQkZXfzKo4Hojvtk15ojkReCh6+IlGWqV5R7m4E0cYpfE3Wur
eIw+W3FRMEqT6oLaLZbcaSFiqrk6l1NuPoXiKWNWOI4fplOyCfim6ducC7dSrYwZ
sAZnFP6JFQhSeRZ4Eq5RIJfdsFQgLFGuMt/BsvXrMJdpy++VzmBuJBTYepIEgVkR
4AXWle7KD58ILR+QQttrzGCx+saK3FvB03mPtqAjWDISSmJZ
-----END CERTIFICATE-----`),
		[]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAtj7No0cI/6yEH06zRoq8zJ1nyuGmk8ujpxSfglJcS3kdCZWV
p+QlcZpJqWIXGD2fEO6aXpRGhPWvu23T0ovG0y8+QXXG8w2AEr/X9ZLj81QquOaa
j2vNMss3ttNKquHkaXWdEs2cXTdZeR/+JfD4g4pcS1fWD+oGATZKe3vrz84/1TKv
0J2XabC0U07GtK5kK1bnQYWLFfKYdaXSrxoC8jkFS2HESo05UBEGunGMbZqoO2hy
DUhPJa8BUezYl93R6Jn7vfFG9Y3BxGNp4TKIqnuKxhOWb/Ma9Y+UBi0wE5IPttcI
R3mFrrJIKqj+eYuqR0DJ5f1fL2YcKj3bs7gk1wIDAQABAoIBACZbYomOfmDdjPNP
66CZu7HvITTMuHQb5KZaS1YxAnbqF0f4oUZ0WMMnv4A0gnraIVv9dCUa3RrH4QQ4
UeBbBe5V36sEYGqweTe5A/mOQIbqvJEtC/PKsyYRKnlC0FDV+W6O0A/wkYv3BdBe
AO44YP3Chblf09CGp1vi8ts5kkCqowdvU+14FoyvJFLLC9sgmc4fmxzKD+koabH6
ZdDt6souYkDIKWadZOUvRozeqF0B9oCNx0yHFkDpFd5hrRYb4UnrbHzFM8daVvYK
/Rj3mudxFrlxUstLMRWt0gjyOK5WLRX8spoC89mL0Vay5FO/nW0tj2boAUgiUczy
OwJjU1kCgYEA9hKox+Dn/2xC4IXKd0X9Td+K08LhBld/5MffOA5o09C7CLU62BKK
v4gPJpR5vAsDurHxqL5F6rQ+3iwHuGZKz1+kx3x2zsrFbKRl6XUXS70cwM/K+2eX
aNhzUlyLbhL7GEhtUT7ShhrWkS4XWVJ6lW7MCH8qBeZugqlkM7EhbE8CgYEAvZj1
eE4QvSjWB4K73+MDlc4DFP/9VyYFvfG4ZCsMcjsSuTQMXFmWabwzSSoTONWwdyx1
EoUhgIhtSuJr1xnm++55YukhvrXrsRbcI25Pq76z3ygICGbC3F3iehZ1tr6ybTwH
K/3eR8FfS1UE4WAenggblcTb3TQAnIhpodY+dPkCgYAbeSfY8RZV4StyT920BV9r
k1q3m9lt0NUZoOseIhW4GGTZawp/10ogajtuzkLtKLmo3XcipOO/eZJPUdEm2Fzf
3EjUcOP+4Iq8P3qVXxpTvXB5YnnCKeWwsgHmHyj+CCZ6ppN177KngFWWbfPzaA8B
ohYrmK8Da5/I/MqQLuWRZwKBgQCnwuFo2wJ6rdh7+tTMbO2uLwSRH1WGOFGaWXkD
wQeZR+XwVDqfuHGcC3gBxCYQAxzKxl6szXnwZkb2nNQ5F2VIBCIKQCiovAXZw1V0
UFZUrEAyNBSvgmXnYXdU+eycj64Hc7cQ2OhG67arTIYt+cP9p0TpR7AX0by8xQNa
vNy02QKBgB+3Q6jtcGpBgQIuiTbsMwf5GvII7ZiFVj6H/74WApTmv2RKtvwCsnAW
+JAVu3IVuCgtjuRXF7gzqymGsHOdKw4zQbjrtvm3NC06JytRl/o23SRvG8ubEenL
vJwJWqjqPrgss47cILuN3NRCdTxvkL/H+v0Ax9blsr+mDLTqHV9V
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		return nil, fmt.Errorf("failed to create key pair: %v", err)
	}

	caCertPool, err := x509.SystemCertPool()

	if err != nil {
		return nil, err
	}

	caCertPool.AppendCertsFromPEM([]byte(`-----BEGIN CERTIFICATE-----
MIIDLTCCAhWgAwIBAgIUDrr+DzmERHy8VwiV5nhH2esdUUEwDQYJKoZIhvcNAQEL
BQAwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tMB4XDTIw
MDYyNTEyNTk1MVoXDTIxMDYyNTEyNTk1MVowJjEkMCIGA1UEAxMbZG0tcm9vdC5k
ZWRpY2F0ZWQtbXlzcWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAwnHJTuW0JceyB9o+Pf3luPeffGNVgxHE2em0u2Zq0qnqAyR28gvzZsZVrdYz
d/pbJv3jGV+2a1AWWbBZYYLk1m5zgC/jV9258ja02LxDQIqsPJtcdp5RY09gkBCJ
zmqNCf4UkQRAbWDH2kW5PaXGAEnxPb9loYkFedxGydUY6AXL4JTCzc3v+kJRgY5z
CGQAicszKucXA7A0RHb+mMDs0J/n34/Rualw4j2oeM0J3qcFZnbM563cHzxjkTZr
GwvFDIeTxAed1e2ONSPKhvFr9PdRqSxWxWyofxHUGtDiv1o0yFRY/n4PiSDKWlNR
+IUUTPcAS801WHHM/rwrauuC7wIDAQABo1MwUTAdBgNVHQ4EFgQUuirhS+YmA4bX
8X/iVFRb61fY0V8wHwYDVR0jBBgwFoAUuirhS+YmA4bX8X/iVFRb61fY0V8wDwYD
VR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAC/btdzmk0U7mwEDSdSyh
0Y531nbJMOzaAk4i/mx9uHnfd1wfif2jRS0HuS42j7zDcn4oU5NsmwYnrALLgBDj
SczXsKnQMX9fwgXb35Q1bQarBE5i+kEfARAEPvSBiFCUjcn3st6wr9nCnnajp/fH
5+OU+U6li9soVJzkZXg3EussaEorUeNvos2W84uOhkHA3RjrGSgMwsGgClKqaZm9
AAHANvxrSWZ9h751rULz8x3ANXwzqqPkOq+C38qEs5og/iXLQK5coHq1Nwr3/+iH
Ga/ERVIbKoM4fV5yqqTZxz4QkgajAqJI8ywAU0Ang/MwZWJMCyJC92S2lgRimKA2
Fw==
-----END CERTIFICATE-----`))

	agentClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{clientPair},
			},
		},
	}

	a := agent.New(agentClient)

	_, _, lastSuccessfulBackupTime, err := a.Status("https://f72e0ce1-f5ff-4cec-ac30-b0e95c6fef18.mysql.service.internal:5000")
	if err != nil {
		return nil, err
	}

	resp, err := http.Get("http://localhost:1234/nexttime")

	if err != nil {
		return nil, err
	}

	nextTime, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"last_successful_backup_time": time.Unix(lastSuccessfulBackupTime, 0).String(),
		"next_backup_time": string(nextTime),
	}, nil
}


func (Gatherer) calculateWholePercent(numerator, denominator uint64) uint64 {
	numeratorFloat := float64(numerator)
	denominatorFloat := float64(denominator)
	return uint64((numeratorFloat / denominatorFloat) * 100)
}

func (g Gatherer) IsDatabaseAvailable() bool {
	return g.client.IsAvailable()
}

func (g Gatherer) IsDatabaseFollower() (bool, error) {
	return g.client.IsFollower()
}

func (g *Gatherer) DatabaseMetadata() (globalStatus map[string]string, globalVariables map[string]string, err error) {
	globalStatus, err = g.client.ShowGlobalStatus()
	if err != nil {
		return nil, nil, err
	}

	globalVariables, err = g.client.ShowGlobalVariables()
	if err != nil {
		return nil, nil, err
	}

	currentQueries := -1

	if currentQueriesString, ok := globalStatus["queries"]; ok {
		var err error
		if currentQueries, err = strconv.Atoi(currentQueriesString); err != nil {
			globalStatus["queries_delta"] = "0"
		}
	} else {
		globalStatus["queries_delta"] = "0"
	}

	if g.previousQueries != -1 {
		if currentQueries-g.previousQueries >= 0 {
			globalStatus["queries_delta"] = strconv.Itoa(currentQueries - g.previousQueries)
		}
	}

	g.previousQueries = currentQueries

	return
}

func (g Gatherer) FollowerMetadata() (slaveStatus map[string]string, heartbeatStatus map[string]string, err error) {
	slaveStatus, err = g.client.ShowSlaveStatus()
	if err != nil {
		return nil, nil, err
	}

	heartbeatStatus, err = g.client.HeartbeatStatus()
	if err != nil {
		return slaveStatus, nil, err
	}

	return
}
