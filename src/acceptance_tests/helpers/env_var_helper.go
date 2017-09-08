package helpers

import (
	"os"
	"github.com/onsi/ginkgo"
	"fmt"
)

func GetEnvVar(envVarName string) string {
	value, found := os.LookupEnv(envVarName)
	if !found {
		ginkgo.Fail(fmt.Sprintf("Expected to find environment variable, %s", envVarName))
	}
	return value
}
