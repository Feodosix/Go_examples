package main

import (
	"io"
)

type otpReader struct {
	r    io.Reader
	prng io.Reader
}

func (or *otpReader) Read(p []byte) (int, error) {
	n, err := or.r.Read(p)
	if n > 0 {
		prngData := make([]byte, n)
		_, err := io.ReadFull(or.prng, prngData)
		if err != nil {
			return 0, err
		}

		for i := 0; i < n; i++ {
			p[i] ^= prngData[i]
		}
	}
	return n, err
}

type otpWriter struct {
	w    io.Writer
	prng io.Reader
}

func (ow *otpWriter) Write(p []byte) (int, error) {
	prngData := make([]byte, len(p))
	_, err := io.ReadFull(ow.prng, prngData)
	if err != nil {
		return 0, err
	}

	for i := 0; i < len(p); i++ {
		prngData[i] ^= p[i]
	}

	n, err := ow.w.Write(prngData)
	if err != nil {
		return n, err
	}

	return len(p), nil
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &otpReader{r: r, prng: prng}
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &otpWriter{w: w, prng: prng}
}
