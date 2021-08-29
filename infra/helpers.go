package infra

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net"
	"sync"
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

// NewIPPool creates new IP pool
func NewIPPool(startIP net.IP) *IPPool {
	return &IPPool{
		currentIP: startIP,
	}
}

// IPPool generates IPs from pool
type IPPool struct {
	mu        sync.Mutex
	currentIP net.IP
}

// Next returns next free IP from pool
func (p *IPPool) Next() (net.IP, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentIP[len(p.currentIP)-1] == 0xfe {
		return nil, errors.New("no more IPs available")
	}
	p.currentIP[len(p.currentIP)-1]++
	ip := make([]byte, len(p.currentIP))
	copy(ip, p.currentIP)
	return ip, nil
}
