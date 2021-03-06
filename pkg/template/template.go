package template

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/tmrts/boilr/pkg/boilr"
	"github.com/tmrts/boilr/pkg/prompt"
	"github.com/tmrts/boilr/pkg/util/osutil"
	"github.com/tmrts/boilr/pkg/util/stringutil"
)

// Interface is contains the behavior of boilr templates.
type Interface interface {
	// Executes the template on the given target directory path.
	Execute(string) error

	// If used, the template will execute using default values.
	UseDefaultValues()

	// Returns the metadata of the template.
	Info() Metadata
}

func (t dirTemplate) Info() Metadata {
	return t.Metadata
}

// Get retrieves the template from a path.
func Get(path string) (Interface, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// TODO make context optional
	ctxt, err := func(fname string) (map[string]interface{}, error) {
		f, err := os.Open(fname)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}

			return nil, err
		}
		defer f.Close()

		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(buf, &metadata); err != nil {
			return nil, err
		}

		return metadata, nil
	}(filepath.Join(absPath, boilr.ContextFileName))

	metadataExists, err := osutil.FileExists(filepath.Join(absPath, boilr.TemplateMetadataName))
	if err != nil {
		return nil, err
	}

	md, err := func() (Metadata, error) {
		if !metadataExists {
			return Metadata{}, nil
		}

		b, err := ioutil.ReadFile(filepath.Join(absPath, boilr.TemplateMetadataName))
		if err != nil {
			return Metadata{}, err
		}

		var m Metadata
		if err := json.Unmarshal(b, &m); err != nil {
			return Metadata{}, err
		}

		return m, nil
	}()

	return &dirTemplate{
		Context:  ctxt,
		FuncMap:  FuncMap,
		Path:     filepath.Join(absPath, boilr.TemplateDirName),
		Metadata: md,
	}, err
}

type dirTemplate struct {
	Path     string
	Context  map[string]interface{}
	FuncMap  template.FuncMap
	Metadata Metadata

	alignment         string
	ShouldUseDefaults bool
}

func (t *dirTemplate) UseDefaultValues() {
	t.ShouldUseDefaults = true
}

func (t *dirTemplate) BindPrompts() {
	if t.ShouldUseDefaults {
		for s, v := range t.Context {
			t.FuncMap[s] = func() interface{} {
				switch v := v.(type) {
				// First is the default value if it's a slice
				case []interface{}:
					return v[0]
				}

				return v
			}
		}
	} else {
		for s, v := range t.Context {
			t.FuncMap[s] = prompt.New(s, v)
		}
	}
}

// Execute fills the template with the project metadata.
func (t *dirTemplate) Execute(dirPrefix string) error {
	t.BindPrompts()

	// TODO create io.ReadWriter from string
	// TODO refactor name manipulation
	// TODO trim leading or trailing whitespaces
	return filepath.Walk(t.Path, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Path relative to the root of the template directory
		oldName, err := filepath.Rel(t.Path, filename)
		if err != nil {
			return err
		}

		buf := stringutil.NewString("")

		// TODO translate errors into meaningful ones
		fnameTmpl := template.Must(template.
			New("file name template").
			Option(Options...).
			Funcs(FuncMap).
			Parse(oldName))

		if err := fnameTmpl.Execute(buf, nil); err != nil {
			return err
		}

		newName := buf.String()

		target := filepath.Join(dirPrefix, newName)

		if info.IsDir() {
			if err := os.MkdirAll(target, 0744); err != nil {
				return err
			}
		} else {
			fi, err := os.Lstat(filename)
			if err != nil {
				return err
			}

			// Delete target file if it exists
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return err
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, fi.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			contentsTmpl := template.Must(template.
				New("file contents template").
				Option(Options...).
				Funcs(FuncMap).
				ParseFiles(filename))

			fileTemplateName := filepath.Base(filename)

			if err := contentsTmpl.ExecuteTemplate(f, fileTemplateName, nil); err != nil {
				return err
			}
		}

		return nil
	})
}
