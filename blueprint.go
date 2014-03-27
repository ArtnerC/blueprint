package blueprint

import (
	"fmt"
	htemp "html/template"
	"io"
	"io/ioutil"
	logger "log"
	"os"
	"path/filepath"
	"sync"
)

// TODO: Lazy load or recompile based on fsnotify
// TODO: Find templates by pattern
// TODO: Localized template support

var log = logger.New(os.Stderr, "[blueprint] ", 0)

type Blueprint struct {
	templates map[string]*htemp.Template
	fmap      htemp.FuncMap
	mutex     sync.RWMutex
	dir       string
	master    string
	extra     []string
}

var bp = &Blueprint{
	templates: make(map[string]*htemp.Template),
	fmap:      htemp.FuncMap{},
	mutex:     sync.RWMutex{},
}

func (b *Blueprint) set(name string, t *htemp.Template) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.templates[name] = t
}

func (b *Blueprint) exists(name string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	_, ok := b.templates[name]
	return ok
}

func (b *Blueprint) execute(wr io.Writer, name string, data interface{}) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	t, ok := b.templates[name]
	if !ok {
		return fmt.Errorf("execute template (%s): not found", name)
	}
	return t.Execute(wr, data)
}

func (b *Blueprint) mapFunc(name string, value interface{}) {
	b.fmap[name] = value
}

func CompileDir(master, dir string, extra ...string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("compile dir (%s): %v", dir, err)
	}
	bp.dir = dir
	// if bp.dir, err = filepath.Abs(dir); err != nil {
	// 	log.Printf("compile dir: filepath.Abs failed: %s", err.Error())
	// }

	bp.master = master
	bp.extra = []string{}
	common := []string{master}
	for _, e := range extra {
		if _, err := os.Stat(filepath.Join(dir, e)); err == nil {
			common = append(common, e)
			bp.extra = append(bp.extra, e)
		} else {
			log.Printf("compile dir: extra file %s not found", e)
		}
	}
	common = prependEach(dir, common...)

	for _, f := range files {
		name := f.Name()
		fullName := filepath.Join(dir, name)

		if !isOneOf(fullName, common...) && !f.IsDir() {
			log.Printf("Building template %s with %v", name, common)
			t := htemp.New("")

			//Map template functions
			fmap := htemp.FuncMap{
				"templateName": func() string { return name },
				"isTemplate":   func(s string) bool { return s == name },
				"htmlComment":  funcHtmlComment,
			}
			for k, v := range bp.fmap {
				fmap[k] = v
			}
			t.Funcs(fmap)

			//Parse the templates
			if _, err = t.ParseFiles(append(common, fullName)...); err == nil {
				if t = t.Lookup(master); t != nil {
					bp.set(name, t)
				}
			}
		}
	}
	return err
}

func CompileTemplate(name, master, dir string, extra ...string) error {
	common := []string{master}
	common = append(common, extra...)
	common = prependEach(dir, common...)

	log.Printf("Building template %s with %v", name, common)
	fullName := filepath.Join(dir, name)
	t := htemp.New("")

	//Map template functions
	fmap := htemp.FuncMap{
		"templateName": func() string { return name },
		"isTemplate":   func(s string) bool { return s == name },
		"htmlComment":  funcHtmlComment,
	}
	for k, v := range bp.fmap {
		fmap[k] = v
	}
	t.Funcs(fmap)

	//Parse the templates
	var err error
	if _, err = t.ParseFiles(append(common, fullName)...); err == nil {
		if t = t.Lookup(master); t != nil {
			bp.set(name, t)
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
	return bp.execute(wr, name, data)
}

func Map(name string, value interface{}) {
	bp.mapFunc(name, value)
}

func SaveGenerated(dir string) {
	wd, _ := os.Getwd()
	abswd, _ := filepath.Abs(wd)
	absdir, _ := filepath.Abs(dir)
	if abswd == absdir {
		panic("Can't delete wd")
	}
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		panic(err.Error())
	}
	if err := os.Mkdir(dir, 0666); err != nil {
		panic(err.Error())
	}

	bp.mutex.RLock()
	defer bp.mutex.RUnlock()
	for k, v := range bp.templates {
		f, err := os.Create(filepath.Join(dir, k))
		if err != nil {
			panic(err.Error())
		}
		v.Execute(f, nil)
		f.Close()
	}
}

func SaveTemplate(name, dir string) {
	fullName := filepath.Join(dir, name)
	if err := os.Remove(fullName); err != nil && !os.IsNotExist(err) {
		panic(err.Error())
	}
	f, err := os.Create(fullName)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	bp.execute(f, name, nil)
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
