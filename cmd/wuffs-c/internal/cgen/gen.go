// Copyright 2017 The Wuffs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

package main

// gen.go converts base.* to data.go.
//
// Invoke it via "go generate".

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

const columns = 1024

func main() {
	if err := main1(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func main1() error {
	out := &bytes.Buffer{}
	out.WriteString("// Code generated by running \"go generate\". DO NOT EDIT.\n")
	out.WriteString("\n")
	out.WriteString("// Copyright 2017 The Wuffs Authors.\n")
	out.WriteString("//\n")
	out.WriteString("// Licensed under the Apache License, Version 2.0 (the \"License\");\n")
	out.WriteString("// you may not use this file except in compliance with the License.\n")
	out.WriteString("// You may obtain a copy of the License at\n")
	out.WriteString("//\n")
	out.WriteString("//    https://www.apache.org/licenses/LICENSE-2.0\n")
	out.WriteString("//\n")
	out.WriteString("// Unless required by applicable law or agreed to in writing, software\n")
	out.WriteString("// distributed under the License is distributed on an \"AS IS\" BASIS,\n")
	out.WriteString("// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n")
	out.WriteString("// See the License for the specific language governing permissions and\n")
	out.WriteString("// limitations under the License.\n")
	out.WriteString("\n")
	out.WriteString("package cgen\n")
	out.WriteString("\n")

	if err := genBase(out); err != nil {
		return err
	}
	if err := genTemplates(out); err != nil {
		return err
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return err
	}
	return ioutil.WriteFile("data.go", formatted, 0644)
}

func genBase(out *bytes.Buffer) error {
	files := []struct {
		filename, varname string
	}{
		{"base-header.h", "baseHeader"},
		{"base-impl.h", "baseImpl"},
	}

	for _, f := range files {
		in, err := ioutil.ReadFile(f.filename)
		if err != nil {
			return err
		}

		const afterEditing = "// After editing this file,"
		if !bytes.HasPrefix(in, []byte(afterEditing)) {
			return fmt.Errorf("%s's contents do not start with %q", f.filename, afterEditing)
		}
		if i := bytes.Index(in, []byte("\n\n")); i >= 0 {
			in = in[i+2:]
		}

		fmt.Fprintf(out, "const %s = \"\" +\n", f.varname)
		for len(in) > 0 {
			s := in
			if len(s) > columns {
				s = s[:columns]
			}
			in = in[len(s):]
			fmt.Fprintf(out, "%q +\n", s)
		}
		out.WriteString("\"\"\n\n")
	}
	return nil
}

type template struct {
	name       string
	args       [][2]string
	format     string
	formatArgs []string
}

func genTemplates(out *bytes.Buffer) error {
	templates, err := readTemplates()
	if err != nil {
		return err
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].name < templates[j].name
	})
	for _, t := range templates {
		out.WriteString("\n")
		fmt.Fprintf(out, "type template_args_%s struct{\n", t.name)
		for _, a := range t.args {
			key, val := stripq(a[0]), a[1]
			switch val {
			case "%s":
				fmt.Fprintf(out, "%s string\n", key)
			}
		}
		fmt.Fprintf(out, "}\n\n")
		fmt.Fprintf(out, "func template_%s(b *buffer, args template_args_%s) error {\n", t.name, t.name)
		fmt.Fprintf(out, "b.printf(%q,\n", t.format)
		for _, a := range t.formatArgs {
			fmt.Fprintf(out, "args.%s,\n", stripq(a))
		}
		fmt.Fprintf(out, ")\n")
		fmt.Fprintf(out, "return nil\n")
		fmt.Fprintf(out, "}\n")
	}
	return nil
}

func readTemplates() ([]template, error) {
	f, err := os.Open("templates.h")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	templates := []template(nil)
	name := ""
	args := [][2]string(nil)
	argsMap := map[string]string{}
	format := []byte(nil)
	formatArgs := []string(nil)

	r := bufio.NewScanner(f)
	for r.Scan() {
		s := r.Text()
		if len(s) == 0 || strings.HasPrefix(s, "//") {
			continue
		}

		if name == "" {
			line := s
			const prefix = "template "
			if !strings.HasPrefix(s, prefix) {
				return nil, fmt.Errorf("bad template line %q", line)
			}
			s = s[len(prefix):]
			if i := strings.IndexByte(s, '('); i < 0 {
				return nil, fmt.Errorf("bad template line %q", line)
			} else {
				name = s[:i]
				s = s[i+1:]
			}
			if i := strings.IndexByte(s, ')'); i < 0 {
				return nil, fmt.Errorf("bad template line %q", line)
			} else {
				s = s[:i]
			}
			for _, x := range strings.Split(s, ",") {
				x = strings.TrimSpace(x)
				key, val := "", ""
				switch {
				case strings.HasPrefix(x, "string "):
					key = x[len("string "):]
					val = "%s"
				}
				if key == "" || argsMap[key] != "" {
					return nil, fmt.Errorf("bad template line %q", line)
				}
				args = append(args, [2]string{key, val})
				argsMap[key] = val
			}
			continue
		}

		if s == "}" {
			templates = append(templates, template{
				name:       name,
				args:       args,
				format:     string(format),
				formatArgs: formatArgs,
			})
			name = ""
			args = nil
			argsMap = map[string]string{}
			format = nil
			formatArgs = nil
			continue
		}

		s = strings.TrimSpace(s)
		if len(s) == 0 || strings.HasPrefix(s, "//") {
			continue
		}
		s = strings.Replace(s, "%", "%%", -1)

		for _, a := range args {
			key, val := a[0], a[1]
			for {
				i := strings.Index(s, key)
				if i < 0 {
					break
				}
				s = s[:i] + val + s[i+len(key):]
				formatArgs = append(formatArgs, key)
			}
		}
		format = append(format, s...)
		format = append(format, '\n')
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	return templates, nil
}

func stripq(s string) string {
	if len(s) >= 2 && s[0] == 'q' && s[len(s)-1] == 'q' {
		return s[1 : len(s)-1]
	}
	return s
}
