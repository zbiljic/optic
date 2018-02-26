package metric

import (
	"regexp"
	"strings"
)

var (
	metricNameRegexp = regexp.MustCompile(`[a-zA-Z]+[a-zA-Z0-9_:]*`)

	// *sanitizer regexp is for sanitizing:
	//   - tag keys
	//   - tag values
	//   - fields keys
	prefixSanitizer = regexp.MustCompile(`^[^a-zA-Z_]`)
	bodySanitizer   = regexp.MustCompile(`[^a-zA-Z0-9_]`)

	// name*sanitizer regexp is for sanitizing metric names only.
	namePrefixSanitizer = regexp.MustCompile(`^[^a-zA-Z_:]`)
	nameBodySanitizer   = regexp.MustCompile(`[^a-zA-Z0-9_:]`)

	// stringFieldEscaper is for escaping string field values only.
	stringFieldEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

func MetricNameValid(name string) bool {
	return metricNameRegexp.MatchString(name)
}

func sanitize(s string, t string) string {
	switch t {
	case "tagkey":
		return bodySanitizer.ReplaceAllString(
			prefixSanitizer.ReplaceAllString(s, "_"),
			"_",
		)
	case "name", "fieldkey":
		return nameBodySanitizer.ReplaceAllString(
			namePrefixSanitizer.ReplaceAllString(s, "_"),
			"_",
		)
	case "fieldval":
		return stringFieldEscaper.Replace(s)
	}
	return s
}
