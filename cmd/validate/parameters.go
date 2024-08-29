package validate

import "strings"

func parametersFromPath(path string) []string {
	params := make([]string, 0)
	for _, segment := range strings.Split(path, "/") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		if !(strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")) {
			continue
		}
		params = append(params, strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}"))
	}
	return params
}
