// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"bufio"
	"io"
	"sync"
)

type OutputType int

const (
	Stdout OutputType = iota
	Stderr
)

type OutputLine struct {
	Type OutputType
	Line string
}

type PipeReader struct {
	StdoutPipe io.ReadCloser
	StderrPipe io.ReadCloser
}

func NewPipeReader(stdout, stderr io.ReadCloser) (*PipeReader, error) {
	return &PipeReader{
		StdoutPipe: stdout,
		StderrPipe: stderr,
	}, nil
}

func (pr *PipeReader) Start() <-chan OutputLine {
	var wg sync.WaitGroup
	wg.Add(2)
	ch := make(chan OutputLine, 100)

	go func() {
		defer wg.Done()
		readLines(pr.StdoutPipe, Stdout, ch)
	}()

	go func() {
		defer wg.Done()
		readLines(pr.StderrPipe, Stderr, ch)
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func readLines(r io.Reader, outputType OutputType, ch chan<- OutputLine) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ch <- OutputLine{
			Type: outputType,
			Line: scanner.Text(),
		}
	}
}
