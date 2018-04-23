package fio

import (
	"log"
	"strings"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/bashexecuter"
)

// RunSmall Executes small fio file located in /app/fio
func RunFioAndGenPlot(size, mode, bsize string) error {
	var res string
	var err error
	if bsize == "notSet" {
		res, err = bashexecuter.Execute("/app/fio/fio-" + mode + "-" + size + ".sh")
	} else {
		res, err = bashexecuter.Execute("/app/fio/fio-" + mode + "-" + size + ".sh " + bsize)
	}
	log.Println("/app/fio/fio-" + mode + "-" + size + ".sh " + bsize)
	if err != nil && strings.Contains(res, "fail") {
		logging.WithError("BA-OPERATOR-FIO-SMALL-001", err).Panicln(res)
		return err
	}
	return FioGenPlot()
}

// FioGenPlot creates plots from fio bw files and moves them to /app/pictures
func FioGenPlot() error {
	res, err := bashexecuter.Execute("/app/fio/fiogenplot.sh")
	if err != nil && strings.Contains(res, "fail") {
		logging.WithError("BA-OPERATOR-FIO-GENPLOT-001", err).Panicln(res)
	}
	return err
}
