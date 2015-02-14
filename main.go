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
	"9fans.net/go/plan9"
	"9fans.net/go/plumb"
)

// a.txt:1,2 c b.txt:1
// => [a.txt 1,2 b.txt 1]
func parseLocs(s string) (string, string, string, string, error) {
	chunks := strings.Split(strings.TrimSpace(s), " ")
	if len(chunks) != 3 {
		return "", "", "", "", errors.New(fmt.Sprintf("malformed line: %s", s))
	}
	loc1 := chunks[0]
	loc2 := chunks[2]
	loc1split := strings.Split(loc1, ":")
	if len(loc1split) != 2 {
		return "", "", "", "", errors.New(fmt.Sprintf("malformed line: %s", s))
	}
	loc2split := strings.Split(loc2, ":")
	if len(loc2split) != 2 {
		return "", "", "", "", errors.New(fmt.Sprintf("malformed line: %s", s))
	}
	return loc1split[0], loc1split[1], loc2split[0], loc2split[0], nil
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

func plumbFile(file, addr string) error {
	port, err := plumb.Open("send", plan9.OWRITE)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return err
	}
	cwd, _ := os.Getwd()
	defer port.Close()
	attr := plumb.Attribute{"addr", addr, nil}
	msg := plumb.Message{
		Src:  "NextDiff",
		Dst:  "edit",
		Dir:  cwd,
		Type: "text",
		Attr: &attr,
		Data: []byte(file),
	}
	return msg.Send(port)
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
	f1, a1, f2, a2, err := parseLocs(string(line))
	err = plumbFile(f1, a1)
	if err != nil {
		log.Fatal("error plumbing address ", a1, err)
	}
	err = plumbFile(f2, a2)
	if err != nil {
		log.Fatal("error plumbing address ", a2, err)
	}
}
