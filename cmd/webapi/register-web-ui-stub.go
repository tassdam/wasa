//go:build !webui

package main

import (
	"net/http"
)

func registerWebUI(hdl http.Handler) (http.Handler, error) {
	return hdl, nil
}
