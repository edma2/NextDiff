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
	"path/filepath"
	"strconv"
	"strings"

	"9fans.net/go/acme"
	"9fans.net/go/plan9"
	"9fans.net/go/plumb"
)

var cwd string

type Loc struct {
	path string
	addr string
}

// a.txt:1,2
func parseLoc(s string) (*Loc, error) {
	split := strings.Split(s, ":")
	if len(split) != 2 {
		return nil, errors.New(fmt.Sprintf("malformed location: %s", s))
	}
	rawPath := split[0]
	path, err := filepath.Abs(rawPath)
	if err != nil {
		return nil, err
	}
	addr := split[1]
	return &Loc{path, addr}, nil
}

// a.txt:1,2 c b.txt:1
func parseLocs(s string) (*Loc, *Loc, error) {
	chunks := strings.Split(strings.TrimSpace(s), " ")
	if len(chunks) != 3 {
		return nil, nil, errors.New(fmt.Sprintf("malformed line: %s", s))
	}
	loc1, err := parseLoc(chunks[0])
	if err != nil {
		return nil, nil, err
	}
	loc2, err := parseLoc(chunks[2])
	if err != nil {
		return nil, nil, err
	}
	return loc1, loc2, nil
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

func plumbFile(loc *Loc) error {
	port, err := plumb.Open("send", plan9.OWRITE)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return err
	}
	defer port.Close()
	attr := plumb.Attribute{"addr", loc.addr, nil}
	msg := plumb.Message{
		Src:  "NextDiff",
		Dst:  "edit",
		Dir:  cwd,
		Type: "text",
		Attr: &attr,
		Data: []byte(loc.path),
	}
	return msg.Send(port)
}

func showOrPlumb(loc *Loc) error {
	w, err := openWin(loc.path)
	if err != nil {
		return err
	}
	if w == nil {
		return plumbFile(loc)
	} else {
		return showAddr(loc.addr, w)
	}
}

func openWin(name string) (*acme.Win, error) {
	wins, err := acme.Windows()
	if err != nil {
		return nil, err
	}
	for _, w := range wins {
		if w.Name == name {
			return acme.Open(w.ID, nil)
		}
	}
	return nil, nil
}

func main() {
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal("error getting current working directory", err)
	}
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
	bytes, _ := w.ReadAll("xdata")
	line := string(bytes)
	loc1, loc2, err := parseLocs(line)
	if err != nil {
		log.Fatal("error parsing locations ", line, err)
	}
	err = showOrPlumb(loc1)
	if err != nil {
		log.Fatal("error plumbing address ", loc1, err)
	}
	err = showOrPlumb(loc2)
	if err != nil {
		log.Fatal("error plumbing address ", loc2, err)
	}
}
