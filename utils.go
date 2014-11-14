package wade

import neturl "net/url"

// UrlQuery adds query arguments (?arg1=value1&arg2=value2...)
// specified in the given map args to a given url and returns the new result
func UrlQuery(url string, args map[string][]string) string {
	qs := neturl.Values(args).Encode()
	if qs == "" {
		return url
	}

	return url + "?" + qs
}
