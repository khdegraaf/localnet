package infra

import (
	"bytes"
	"context"
	"io/ioutil"
	"text/template"

	"github.com/ridge/must"
)

// PreprocessApp runs all the operations required to prepare app to be deployed
func PreprocessApp(ctx context.Context, ip string, app AppBase) error {
	data := struct {
		IP string
	}{
		IP: ip,
	}

	for i, arg := range app.Args {
		tpl := template.Must(template.New("").Parse(arg))
		buf := &bytes.Buffer{}
		must.OK(tpl.Execute(buf, data))
		app.Args[i] = buf.String()
	}

	for _, file := range app.Files {
		if file.Preprocess {
			tpl := template.Must(template.New("").Parse(string(file.Content)))
			buf := &bytes.Buffer{}
			must.OK(tpl.Execute(buf, data))
			file.Content = buf.Bytes()
		}
		must.OK(ioutil.WriteFile(file.Path, file.Content, 0o600))
	}

	if app.PreFunc != nil {
		return app.PreFunc(ctx)
	}
	return nil
}
