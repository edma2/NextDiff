// The script will highlight each line, avoiding the cursor jump to
// highlighted regions as happens by plumbing or right-clicking on a
// file pattern. Though the cursor jump will occur the first time the
// diffed files are opened, for subsequent execution of NextDiff for
// diffs within the same files the cursor will remain over the NextDiff
// command and the highlighted regions will change with file scrolling
// to show at least part of the changed regions.
//
// When the files first open you'll still need to manually arrange the
// files side by side. There is no acme API for window placement.
// However, the command will save some amount of scrolling, clicking,
// and mouse movement within the adiff output.
//
// http://ipn.caerwyn.com/2009/05/lab-95-acme-side-by-side-diff.html
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"9fans.net/go/acme"
)

// a.txt:1,2 c b.txt:1
func parseAddrs(s string) (string, string, error) {
	ss := strings.Split(strings.TrimSpace(s), " ")
	if len(ss) != 3 {
		return "", "", errors.New(fmt.Sprintf("malformed line: %s", s))
	}
	return ss[0], ss[2], nil
}

func setAddrToDot(w *acme.Win) error {
	_, _, err := w.ReadAddr() // first read is bogus
	if err != nil {
		return err
	}
	return w.Ctl("addr=dot\n")
}

func showAddr(addr string, w *acme.Win) error {
	err := w.Addr(addr)
	if err != nil {
		return err
	}
	err = w.Ctl("dot=addr\n")
	if err != nil {
		return err
	}
	return w.Ctl("show\n")
}

func plumbAddr(a string) error {
	return nil
}

func main() {
	id, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		log.Fatal("error getting winid", err)
	}
	w, err := acme.Open(id, nil)
	if err != nil {
		log.Fatal("error opening acme win", err)
	}
	defer w.CloseFiles()
	err = setAddrToDot(w)
	if err != nil {
		log.Fatal("error setting address to dot", err)
	}
	err = showAddr(`/^[^\-<>].*\n/`, w)
	if err != nil {
		log.Fatal("error searching window", err)
	}
	line, _ := w.ReadAll("xdata")
	a1, a2, err := parseAddrs(string(line))
	if err != nil {
		log.Fatal("error parsing address", err)
	}
	err = plumbAddr(a1)
	if err != nil {
		log.Fatal("error plumbing address ", a1, err)
	}
	err = plumbAddr(a2)
	if err != nil {
		log.Fatal("error plumbing address ", a2, err)
	}
}
