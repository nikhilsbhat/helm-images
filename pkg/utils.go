package pkg

import (
	"fmt"
	"os"
	"strings"
)

type ValueFiles []string

func (v *ValueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *ValueFiles) Valid() error {
	errStr := ""

	for _, valuesFile := range *v {
		if strings.TrimSpace(valuesFile) != "-" {
			if _, err := os.Stat(valuesFile); os.IsNotExist(err) {
				errStr += err.Error()
			}
		}
	}

	if len(errStr) == 0 {
		return nil
	}

	//nolint:goerr113
	return fmt.Errorf("%s", errStr)
}

func (v *ValueFiles) Type() string {
	return "ValueFiles"
}

func (v *ValueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}

	return nil
}

func GetUniqEntries(slice []string) []string {
	for slc := 0; slc < len(slice); slc++ {
		if Contains(slice[slc+1:], slice[slc]) {
			slice = append(slice[:slc], slice[slc+1:]...)
			slc--
		}
	}

	return slice
}

func Contains(slice []string, image string) bool {
	for _, slc := range slice {
		if slc == image {
			return true
		}
	}

	return false
}
