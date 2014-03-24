package blueprint

import (
	"fmt"
	htemp "html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TODO: Lazy load or recompile based on fsnotify
// TODO: Find templates by pattern

var blueprint = struct {
	templates map[string]*htemp.Template
	fmap      htemp.FuncMap
}{
	templates: make(map[string]*htemp.Template),
	fmap:      nil,
}

func CompileDir(master, dir string, extra ...string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("Problem with %s", dir)
	}
	common := []string{master}
	common = append(common, extra...)

	for _, f := range files {
		if !isOneOf(f.Name(), common...) && !f.IsDir() {
			t := htemp.New(master)
			if blueprint.fmap != nil {
				t.Funcs(blueprint.fmap)
			}
			if _, err = t.ParseFiles(prependEach(dir, append(common, f.Name())...)...); err == nil {
				blueprint.templates[f.Name()] = t
			}
		}
	}
	return err
}

func MustCompileDir(master, dir string, extra ...string) {
	if err := CompileDir(master, dir, extra...); err != nil {
		panic(err.Error())
	}
}

func Execute(wr io.Writer, name string, data interface{}) error {
	t, ok := blueprint.templates[name]
	if !ok {
		return fmt.Errorf("Template %s not found", name)
	}
	return t.Execute(wr, data)
}

func Map(name string, value interface{}) {
	if blueprint.fmap == nil {
		blueprint.fmap = make(htemp.FuncMap)
	}
	blueprint.fmap[name] = value
}

func SaveGenerated(dir string) {
	wd, _ := os.Getwd()
	abswd, _ := filepath.Abs(wd)
	absdir, _ := filepath.Abs(dir)
	if abswd == absdir {
		panic("Can't delete wd")
	}
	if err := os.RemoveAll(dir); err != nil {
		panic(err.Error())
	}
	if err := os.Mkdir(dir, 0666); err != nil {
		panic(err.Error())
	}
	for k, v := range blueprint.templates {
		f, err := os.Create(filepath.Join(dir, k))
		if err != nil {
			panic(err.Error())
		}
		v.Execute(f, nil)
		f.Close()
	}
}

func isOneOf(v string, cmp ...string) bool {
	for _, s := range cmp {
		if v == s {
			return true
		}
	}
	return false
}

func prependEach(prefix string, strs ...string) (res []string) {
	for _, s := range strs {
		res = append(res, filepath.Join(prefix, s))
	}
	return
}
