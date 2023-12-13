package util

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	HostRegex       = regexp.MustCompile(`^(?:[\w-]+\.)+\w+$`)
	IpAddressRegex  = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
	UriRegex        = regexp.MustCompile(`^/?(?:[\w-]+/)*[\w-]+/?(?:\?(?:[\w-]+=[\w-]+)(?:&[\w-]+=[\w-]+)*)?$`)
	HttpMethodRegex = regexp.MustCompile(fmt.Sprintf(`^(%s)$`, strings.Join([]string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}, "|")))
)
