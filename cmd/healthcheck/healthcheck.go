package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var port = flag.Int("port", 3000, "HTTP port for healthcheck")
	flag.Parse()
	res, err := http.Get(fmt.Sprintf("http://localhost:%d/liveness", *port))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		_ = res.Body.Close()
		_, _ = fmt.Fprintln(os.Stderr, "Healthcheck request not OK: ", res.Status)
		os.Exit(1)
	}
	_ = res.Body.Close()
	os.Exit(0)
}
