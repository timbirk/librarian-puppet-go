package librarianpuppetgo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

type eachOpts struct {
	prefix, suffix string
	body           string
}

type params struct {
	Name      string
	Ref       string
	RefSemver string
}

func (g *Git) Each(path string, cmds []string, opts eachOpts) {
	mods := parse(path)
	out := os.Stdout
	for _, mod := range mods {
		c, err := makeEachArgs(cmds, mod)
		if err != nil {
			log.Fatalln(err)
		}
		p, err := replaceWithMod(opts.prefix, mod)
		if err != nil {
			log.Fatalln(err)
		}
		s, err := replaceWithMod(opts.suffix, mod)
		if err != nil {
			log.Fatalln(err)
		}
		b := bytes.NewBuffer([]byte{})
		x := bytes.NewBuffer([]byte{})
		if err := run3(b, x, mod.Dest(), c[0], c[1:]); err != nil {
			fmt.Fprintf(out, "# Failed to run `%v` in %v\n", c, mod.Dest())
			fmt.Fprintf(os.Stderr, "%v", x)
			continue
		}

		fmt.Fprint(out, p)
		if opts.body == "" {
			fmt.Fprint(out, b)
		} else {
			v := struct {
				params
				Value string
			}{params{mod.name, mod.Ref(), mod.RefSemver()}, b.String()}
			s, err := replaceWith(opts.body, v)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Fprint(out, s)
		}
		fmt.Fprint(out, s)
	}
}

func replaceWithMod(t string, m Mod) (string, error) {
	return replaceWith(t, params{
		m.name,
		m.Ref(),
		m.RefSemver(),
	})
}

func replaceWith(templ string, v interface{}) (string, error) {
	t, err := template.New("").Parse(templ)
	if err != nil {
		return "", err
	}
	b := bytes.NewBuffer([]byte{})
	t.Execute(b, v)
	s := strings.Replace(b.String(), "\\n", "\n", -1)
	s = strings.Replace(s, "\\t", "\t", -1)
	return s, err
}

func makeEachArgs(args []string, m Mod) ([]string, error) {
	c := make([]string, len(args))
	for i, e := range args {
		s, err := replaceWithMod(e, m)
		if err != nil {
			return []string{e}, err
		}
		c[i] = s
	}
	return c, nil
}
