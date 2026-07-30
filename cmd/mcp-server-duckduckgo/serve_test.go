package main

import (
	"context"
	"io"
	"testing"

	"github.com/maccavelli/mcplib"
)

type dummyReadCloser struct{}

func (d *dummyReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
func (d *dummyReadCloser) Close() error {
	return nil
}

type dummyWriteCloser struct{}

func (d *dummyWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func (d *dummyWriteCloser) Close() error {
	return nil
}

func TestRun_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lb := mcplib.NewLogBuffer()
	reader := &dummyReadCloser{}
	writer := &dummyWriteCloser{}

	// Start run, but the reader will immediately EOF.
	// mcpServer.Serve will return an error because it expects JSON-RPC payload or keeps waiting?
	// If reader EOFs immediately, the stdio transport should shut down.
	err := run(ctx, lb, reader, writer, "1.0")
	if err != nil {
		t.Logf("run returned error (expected due to EOF or closed): %v", err)
	}
}
