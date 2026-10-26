package session

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"

	"mml-server/internal/mapping"
)

const maxRequestBytes = 4096

var errLineTooLong = errors.New("request line exceeds 4096 bytes")

func Handle(ctx context.Context, conn net.Conn, mapper *mapping.Table) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log.Printf("session connected: %s", remote)
	defer log.Printf("session disconnected: %s", remote)

	stop := context.AfterFunc(ctx, func() {
		conn.Close()
	})
	defer stop()

	r := bufio.NewReader(conn)
	for {
		line, err := readRequestLine(r)
		if err != nil {
			if errors.Is(err, errLineTooLong) {
				log.Printf("session %s: line too long", remote)
			} else if errors.Is(err, io.EOF) {
				// client closed
			} else {
				log.Printf("session %s: read error: %v", remote, err)
			}
			return
		}
		resp := mapper.Lookup(line)
		if _, err := conn.Write([]byte(resp + "\r\n")); err != nil {
			log.Printf("session %s: write error: %v", remote, err)
			return
		}
	}
}

func readRequestLine(r *bufio.Reader) (string, error) {
	var buf []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0xFF {
			if err := skipIAC(r); err != nil {
				return "", err
			}
			continue
		}
		if b == '\n' {
			if n := len(buf); n > 0 && buf[n-1] == '\r' {
				buf = buf[:n-1]
			}
			return string(buf), nil
		}
		buf = append(buf, b)
		if len(buf) > maxRequestBytes {
			return "", errLineTooLong
		}
	}
}

func skipIAC(r *bufio.Reader) error {
	cmd, err := r.ReadByte()
	if err != nil {
		return err
	}
	if cmd >= 251 && cmd <= 254 {
		_, err = r.ReadByte()
		return err
	}
	return nil
}
