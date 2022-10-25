package resources

import (
	"fmt"
	"strings"

	f "github.com/fauna/faunadb-go/v5/faunadb"
)

var BlacklistedResourceNames = []string{"events", "sets", "self", "documents", "_"}

func CheckNameNotBlacklisted(name string, resourceType string) error {
	nameTrimmed := strings.TrimSpace(name)
	for _, blacklistedResourceName := range BlacklistedResourceNames {
		if nameTrimmed == blacklistedResourceName {
			return fmt.Errorf("The name of a Fauna '%s' cannot be '%s'.", resourceType, name)
		}
	}

	return nil
}

func GetProperty[U any](obj f.ObjectV, propName string, def U) (U, bool) {
	if value, ok := obj[propName]; ok {
		return ParseFaunaValue[U](value), true
	}

	return def, false
}

func ParseFaunaValue[U any](obj f.Value) U {
	var parsed U
	obj.Get(&parsed)
	return parsed
}
