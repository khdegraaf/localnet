package infra

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"text/template"
	"time"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/lib/retry"
)

// PreprocessApp runs all the operations required to prepare app to be deployed
func PreprocessApp(ctx context.Context, ip net.IP, app AppBase) error {
	data := struct {
		IP net.IP
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

// HealthCheckCapable represents application exposing health check endpoint
type HealthCheckCapable interface {
	// HealthCheck runs single health check
	HealthCheck(ctx context.Context) error
}

// WaitUntilHealthy waits until app is healthy or context expires
func WaitUntilHealthy(ctx context.Context, app HealthCheckCapable) error {
	return retry.Do(ctx, time.Second, func() error {
		return app.HealthCheck(ctx)
	})
}
