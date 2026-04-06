package pkg

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/errors"
)

type ValueFiles []string

func (v *ValueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *ValueFiles) Valid() error {
	var errBuilder strings.Builder

	for _, valuesFile := range *v {
		if strings.TrimSpace(valuesFile) != "-" {
			if _, err := os.Stat(valuesFile); os.IsNotExist(err) {
				errBuilder.WriteString(err.Error())
			}
		}
	}

	if errBuilder.Len() == 0 {
		return nil
	}

	return &errors.ImageError{Message: errBuilder.String()}
}

func (v *ValueFiles) Type() string {
	return "ValueFiles"
}

func (v *ValueFiles) Set(value string) error {
	for filePath := range strings.SplitSeq(value, ",") {
		*v = append(*v, filePath)
	}

	return nil
}

func GetUniqEntries(slice []string) []string {
	encountered := map[string]bool{}
	result := make([]string, 0)

	for _, val := range slice {
		if !encountered[val] {
			encountered[val] = true

			result = append(result, val)
		}
	}

	return result
}

func Contains(slice []string, image string) bool {
	return slices.Contains(slice, image)
}
