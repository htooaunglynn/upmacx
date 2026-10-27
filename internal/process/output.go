package process

import (
	"io"
	"os"
)

// outputPipe keeps os/exec from waiting for inherited descendant pipes before
// returning the direct child's status. That status must be observable promptly
// to propagate foreground SIGINT/SIGTERM to the run controller.
type outputPipe struct {
	reader, writer *os.File
	target         io.Writer
	done           chan error
}

func newOutput(target io.Writer) (io.Writer, *outputPipe, error) {
	if target == nil {
		return nil, nil, nil
	}
	if f, ok := target.(*os.File); ok {
		return f, nil, nil
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	p := &outputPipe{reader: reader, writer: writer, target: target, done: make(chan error, 1)}
	return writer, p, nil
}
func (p *outputPipe) close() {
	if p != nil {
		_ = p.reader.Close()
		_ = p.writer.Close()
	}
}
func (p *outputPipe) start(cancel func()) {
	if p == nil {
		return
	}
	_ = p.writer.Close()
	go func() {
		_, err := io.Copy(p.target, p.reader)
		_ = p.reader.Close()
		if err != nil {
			cancel()
		}
		p.done <- err
	}()
}
func (p *outputPipe) wait() error {
	if p == nil {
		return nil
	}
	return <-p.done
}
