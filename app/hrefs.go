// Code generated by goagen v1.2.0-dirty, DO NOT EDIT.
//
// API "amc": Application Resource Href Factories
//
// Command:
// $ goagen
// --design=github.com/citrusleaf/amc/api_design
// --out=$(GOPATH)/src/github.com/citrusleaf/amc
// --version=v1.2.0-dirty

package app

import (
	"fmt"
	"strings"
)

// BackupHref returns the resource href.
func BackupHref(connID, backupID interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	parambackupID := strings.TrimLeftFunc(fmt.Sprintf("%v", backupID), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/backups/%v", paramconnID, parambackupID)
}

// ConnectionHref returns the resource href.
func ConnectionHref(connID interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v", paramconnID)
}

// DbUserHref returns the resource href.
func DbUserHref(connID, username interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramusername := strings.TrimLeftFunc(fmt.Sprintf("%v", username), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/users/%v", paramconnID, paramusername)
}

// IndexHref returns the resource href.
func IndexHref(connID, name interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramname := strings.TrimLeftFunc(fmt.Sprintf("%v", name), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/indexes/%v", paramconnID, paramname)
}

// LogicalNamespaceHref returns the resource href.
func LogicalNamespaceHref(connID, namespace interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramnamespace := strings.TrimLeftFunc(fmt.Sprintf("%v", namespace), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/logical-namespaces/%v", paramconnID, paramnamespace)
}

// ModuleHref returns the resource href.
func ModuleHref(connID, name interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramname := strings.TrimLeftFunc(fmt.Sprintf("%v", name), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/modules/%v", paramconnID, paramname)
}

// NamespaceHref returns the resource href.
func NamespaceHref(connID, node, namespace interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramnode := strings.TrimLeftFunc(fmt.Sprintf("%v", node), func(r rune) bool { return r == '/' })
	paramnamespace := strings.TrimLeftFunc(fmt.Sprintf("%v", namespace), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/nodes/%v/namespaces/%v", paramconnID, paramnode, paramnamespace)
}

// NodeHref returns the resource href.
func NodeHref(connID, node interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramnode := strings.TrimLeftFunc(fmt.Sprintf("%v", node), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/nodes/%v", paramconnID, paramnode)
}

// SetHref returns the resource href.
func SetHref(connID, node, namespace, setName interface{}) string {
	paramconnID := strings.TrimLeftFunc(fmt.Sprintf("%v", connID), func(r rune) bool { return r == '/' })
	paramnode := strings.TrimLeftFunc(fmt.Sprintf("%v", node), func(r rune) bool { return r == '/' })
	paramnamespace := strings.TrimLeftFunc(fmt.Sprintf("%v", namespace), func(r rune) bool { return r == '/' })
	paramsetName := strings.TrimLeftFunc(fmt.Sprintf("%v", setName), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/connections/%v/nodes/%v/namespaces/%v/sets/%v", paramconnID, paramnode, paramnamespace, paramsetName)
}

// XdrHref returns the resource href.
func XdrHref(node interface{}) string {
	paramnode := strings.TrimLeftFunc(fmt.Sprintf("%v", node), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/api/v1/xdr/%v", paramnode)
}
