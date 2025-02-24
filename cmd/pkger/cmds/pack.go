package cmds

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/markbates/pkger"
	"github.com/markbates/pkger/parser"
	"github.com/markbates/pkger/pkging/pkgutil"
)

const outName = "pkged.go"

type packCmd struct {
	*flag.FlagSet
	help bool
	subs []command
}

func (e *packCmd) Name() string {
	return e.Flags().Name()
}

func (e *packCmd) Exec(args []string) error {
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

	if err := Package(fp, decls); err != nil {
		return err
	}

	return nil
}

func (e *packCmd) Route(args []string) error {
	e.Parse(args)

	if e.help {
		e.Usage()
		return nil
	}

	args = e.Args()

	if len(args) == 0 {
		return e.Exec(args)
	}

	k := args[0]
	for _, c := range e.subs {
		if k == c.Name() {
			args = args[1:]
			for _, a := range args {
				if a == "-h" {
					Usage(os.Stderr, c.Flags())()
					return nil
				}
			}
			return c.Exec(args)
		}
	}

	return fmt.Errorf("unknown arguments: %s", strings.Join(args, " "))
}

func New() (*packCmd, error) {
	c := &packCmd{}

	c.subs = []command{
		&serveCmd{}, &statCmd{}, &infoCmd{}, &pathCmd{}, &parseCmd{}, &listCmd{},
	}
	sort.Slice(c.subs, func(a, b int) bool {
		return c.subs[a].Name() <= c.subs[b].Name()
	})

	c.FlagSet = flag.NewFlagSet("pkger", flag.ExitOnError)
	c.BoolVar(&c.help, "h", false, "prints help information")
	c.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		Usage(os.Stderr, c.FlagSet)()
		for _, s := range c.subs {
			Usage(os.Stderr, s.Flags())()
		}
	}
	return c, nil
}

func (e *packCmd) Flags() *flag.FlagSet {
	if e.FlagSet == nil {
		e.FlagSet = flag.NewFlagSet("", flag.ExitOnError)
	}
	e.Usage = Usage(os.Stderr, e.FlagSet)
	return e.FlagSet
}

func Package(out string, decls parser.Decls) error {
	os.RemoveAll(out)

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	c, err := pkger.Current()
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "package %s\n\n", c.Name)
	fmt.Fprintf(f, "import \"github.com/markbates/pkger\"\n\n")
	fmt.Fprintf(f, "import \"github.com/markbates/pkger/pkging/mem\"\n\n")
	fmt.Fprintf(f, "\nvar _ = pkger.Apply(mem.UnmarshalEmbed([]byte(`")

	if err := pkgutil.Stuff(f, c, decls); err != nil {
		return err
	}

	fmt.Fprintf(f, "`)))\n")

	return nil
}
