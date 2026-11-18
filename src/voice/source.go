package voice

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SourceOptions struct {
	FFmpeg string
	Start  time.Duration
}

type source struct {
	io.ReadCloser
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func OpenSource(ctx context.Context, input string, options SourceOptions) (io.ReadCloser, error) {
	if input == "" || strings.HasPrefix(input, "-") {
		return nil, errors.New("audio input must be a file path or URL")
	}
	if options.Start < 0 {
		return nil, errors.New("audio start offset must not be negative")
	}
	binaryPath := options.FFmpeg
	if binaryPath == "" {
		binaryPath = "ffmpeg"
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-nostdin", "-protocol_whitelist", "file,http,https,tcp,tls,crypto"}
	if options.Start > 0 {
		args = append(args, "-ss", strconv.FormatFloat(options.Start.Seconds(), 'f', 3, 64))
	}
	args = append(args, "-i", input, "-vn", "-f", "s16le", "-ar", strconv.Itoa(SampleRate), "-ac", strconv.Itoa(Channels), "pipe:1")
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	configure(cmd)
	cmd.WaitDelay = time.Second
	cmd.Stderr = io.Discard
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		return nil, errors.New("audio decoder could not start")
	}
	return &source{ReadCloser: stdout, cmd: cmd, cancel: cancel}, nil
}

func (s *source) Close() error {
	s.cancel()
	_ = s.ReadCloser.Close()
	_ = s.cmd.Wait()
	return nil
}
