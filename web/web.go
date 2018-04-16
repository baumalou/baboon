package web

import (
	"flag"
	"log"
	"net/http"

	"git.workshop21.ch/workshop21/ba/operator/configuration"
)

func Serve(config *configuration.Config) {
	port := config.WebPort
	directory := config.WebPath
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(directory)))

	log.Printf("Serving %s on HTTP port: %s\n", directory, port)
	log.Println(http.ListenAndServe(":"+port, nil))
}
