// Code generated by goagen v1.2.0-dirty, DO NOT EDIT.
//
// API "amc": Application Resource Href Factories
//
// Command:
// $ goagen
// --design=github.com/citrusleaf/amc/api_design
// --out=$(GOPATH)/src/github.com/citrusleaf/amc/temp
// --version=v1.2.0-dirty

package app

import (
	"fmt"
	"strings"
)

// ConnectionHref returns the resource href.
func ConnectionHref(connID interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v", paramconnID)
}

// ModuleHref returns the resource href.
func ModuleHref(connID, name interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramname := strings.TrimLeftFunc(fmt.Sprintf("%v", name), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/modules/%v", paramconnID, paramname)
}
