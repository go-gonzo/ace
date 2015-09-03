package ace

// github.com/yosssi/gcss binding for Gonzo.

import (
	"bytes"
	html "html/template"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/omeid/gonzo/context"

	"github.com/omeid/gonzo"

	"github.com/yosssi/ace"
)

type (
	Options ace.Options
)

func Compile(options Options, data interface{}) gonzo.Stage {
	return func(ctx context.Context, in <-chan gonzo.File, out chan<- gonzo.File) error {

		options := ace.Options(options)

		fs := []*ace.File{}

		for {
			select {
			case file, ok := <-in:
				if !ok {
					return nil
				}

				buf := new(bytes.Buffer)
				_, err := buf.ReadFrom(file)
				file.Close()
				if err != nil {
					ctx.Error(err)
					continue
				}

				s, err := file.Stat()
				if err != nil {
					return err
				}

				name := s.Name() //Probably filepath.Rel(file.Dir, file.Path) ??
				f := ace.NewFile(name, buf.Bytes())
				source := ace.NewSource(
					ace.NewFile("", nil),
					f,
					fs,
				)

				fs = append(fs, f)

				r, err := ace.ParseSource(source, &options)
				if err != nil {
					ctx.Error(err)
					continue
				}

				t, err := ace.CompileResultWithTemplate(html.New(name), r, &options)
				if err != nil {
					ctx.Error(err)
					continue
				}

				if filepath.Base(file.FileInfo().Name())[0] != '_' {
					buf := new(bytes.Buffer)
					err = t.Execute(buf, data)
					if err != nil {
						ctx.Error(err)
						continue
					}

					file = gonzo.NewFile(ioutil.NopCloser(buf), file.FileInfo())
					file.FileInfo().SetSize(int64(buf.Len()))
					name = strings.TrimSuffix(file.FileInfo().Name(), ".ace") + ".html"
					file.FileInfo().SetName(name)
				}

				out <- file

			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		}
	}
}
