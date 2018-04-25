package testutil

import (
	"net/url"
	"strconv"
	"strings"

	. "github.com/onsi/gomega"
)

func ParseURL(urlstr string) (string, uint) {
	url, err := url.Parse(urlstr)
	Expect(err).NotTo(HaveOccurred())

	split := strings.Split(url.Host, ":")
	host := split[0]

	portInt, err := strconv.Atoi(split[1])
	Expect(err).NotTo(HaveOccurred())
	port := uint(portInt)

	return host, port
}
