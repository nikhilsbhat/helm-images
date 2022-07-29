package pkg

import (
	"errors"
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

	if errStr == "" {
		return nil
	}

	return errors.New(errStr)
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

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// The gist that helped out https://gist.github.com/niski84/a6a3b825b6704cc2cbfd39c97b89e640
func findKey(obj interface{}, key string) (interface{}, bool) {
	mobj, ok := obj.(map[string]interface{})
	if !ok {
		return nil, false
	}

	for k, v := range mobj {
		if k == key {
			return v, true
		}

		if m, ok := v.(map[string]interface{}); ok {
			if res, ok := findKey(m, key); ok {
				return res, true
			}
		}

		if va, ok := v.([]interface{}); ok {
			for _, a := range va {
				if res, ok := findKey(a, key); ok {
					return res, true
				}
			}
		}
	}

	return nil, false
}

func getUniqueSlice(slice []string) []string {
	for slc := 0; slc < len(slice); slc++ {
		if find(slice[slc+1:], slice[slc]) {
			slice = append(slice[:slc], slice[slc+1:]...)
			slc--
		}
	}
	return slice
}

func getUniqEntries(slice []string) []string {
	for slc := 0; slc < len(slice); slc++ {
		if contains(slice[slc+1:], slice[slc]) {
			slice = append(slice[:slc], slice[slc+1:]...)
			slc--
		}
	}
	return slice
}

func contains(slice []string, image string) bool {
	for _, slc := range slice {
		if slc == image {
			return true
		}
	}
	return false
}
