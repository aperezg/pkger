package cmds

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/markbates/pkger"
	"github.com/markbates/pkger/internal/takeon/github.com/markbates/hepa"
	"github.com/markbates/pkger/internal/takeon/github.com/markbates/hepa/filters"
	"github.com/markbates/pkger/parser"
)

type listCmd struct {
	*flag.FlagSet
	help bool
	json bool
	subs []command
}

func (e *listCmd) Name() string {
	return e.Flags().Name()
}

func (e *listCmd) Exec(args []string) error {
	e.Parse(args)

	if e.help {
		e.Usage()
		return nil
	}

	args = e.Args()

	info, err := pkger.Current()
	if err != nil {
		return err
	}

	fp := info.FilePath(outName)
	os.RemoveAll(fp)

	decls, err := parser.Parse(info)
	if err != nil {
		return err
	}

	jay := struct {
		ImportPath string         `json:"import_path"`
		Files      []*parser.File `json:"files"`
	}{
		ImportPath: info.ImportPath,
	}

	files, err := decls.Files()
	if err != nil {
		return err
	}
	jay.Files = files

	if e.json {
		bb := &bytes.Buffer{}

		enc := json.NewEncoder(bb)
		enc.SetIndent("", " ")
		if err := enc.Encode(jay); err != nil {
			return err
		}

		hep := hepa.New()
		hep = hepa.With(hep, filters.Home())
		hep = hepa.With(hep, filters.Golang())

		b, err := hep.Filter(bb.Bytes())
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		return err
	}

	fmt.Println(jay.ImportPath)
	for _, f := range jay.Files {
		fmt.Println(" >", f.Path)
	}
	return nil
}

func (e *listCmd) Flags() *flag.FlagSet {
	if e.FlagSet == nil {
		e.FlagSet = flag.NewFlagSet("list", flag.ExitOnError)
		e.BoolVar(&e.json, "json", false, "prints in JSON format")
	}
	e.Usage = Usage(os.Stderr, e.FlagSet)
	return e.FlagSet
}
