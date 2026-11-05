package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/klauspost/compress/zstd"
)

const (
	CompressionZstdStream = "zstd-stream"
)

type packetReader interface {
	next(*packet) error
	close()
}

type plainReader struct{ ws *websocket.Conn }

func (r plainReader) next(p *packet) error { return r.ws.ReadJSON(p) }
func (r plainReader) close()               {}

type zstdReader struct {
	ws      *websocket.Conn
	pipe    *io.PipeReader
	decoder *zstd.Decoder
	json    *json.Decoder
	mu      sync.Mutex
	readErr error
}

func newZstdReader(ws *websocket.Conn) (*zstdReader, error) {
	pr, pw := io.Pipe()
	decoder, err := zstd.NewReader(pr, zstd.WithDecoderConcurrency(1), zstd.WithDecoderLowmem(true), zstd.WithDecoderMaxWindow(64<<20))
	if err != nil {
		_ = pr.Close()
		return nil, err
	}
	r := &zstdReader{ws: ws, pipe: pr, decoder: decoder, json: json.NewDecoder(decoder)}
	go func() {
		for {
			kind, data, err := ws.ReadMessage()
			if err != nil {
				r.mu.Lock()
				r.readErr = err
				r.mu.Unlock()
				_ = pw.CloseWithError(err)
				return
			}
			if kind != websocket.BinaryMessage {
				err = errors.New("Micup Gateway sent a text frame on a compressed connection")
				r.mu.Lock()
				r.readErr = err
				r.mu.Unlock()
				_ = pw.CloseWithError(err)
				return
			}
			if _, err = pw.Write(data); err != nil {
				return
			}
		}
	}()
	return r, nil
}

func (r *zstdReader) next(p *packet) error {
	if err := r.json.Decode(p); err != nil {
		r.mu.Lock()
		readErr := r.readErr
		r.mu.Unlock()
		if readErr != nil {
			return readErr
		}
		return err
	}
	return nil
}

func (r *zstdReader) close() {
	_ = r.pipe.Close()
	r.decoder.Close()
}

type zstdWriter struct {
	buf     bytes.Buffer
	encoder *zstd.Encoder
}

func newZstdWriter() (*zstdWriter, error) {
	w := &zstdWriter{}
	encoder, err := zstd.NewWriter(&w.buf, zstd.WithEncoderConcurrency(1), zstd.WithWindowSize(1<<20))
	if err != nil {
		return nil, err
	}
	w.encoder = encoder
	return w, nil
}

func (w *zstdWriter) frame(message []byte) ([]byte, error) {
	w.buf.Reset()
	if _, err := w.encoder.Write(message); err != nil {
		return nil, err
	}
	if err := w.encoder.Flush(); err != nil {
		return nil, err
	}
	return append([]byte(nil), w.buf.Bytes()...), nil
}
