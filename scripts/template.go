package main

import (
	"log"
	"os"
	templ "text/template"

	"github.com/pkg/errors"
)

var VersionString = "unset"

func main() {
	// If no template exists, use the existing buildpack.toml
	if _, err := os.Stat("buildpack.toml.tmpl"); os.IsNotExist(err) {
		os.Exit(0)
	}

	v := struct {
		Version string
	}{
		Version: VersionString,
	}

	template := templ.Must(templ.ParseFiles("buildpack.toml.tmpl"))
	f, err := os.OpenFile("buildpack.toml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open buildpack.toml"))
	}

	err = template.Execute(f, v)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to write template to buildpack.toml"))
	}

}
