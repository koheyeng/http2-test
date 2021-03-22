package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func main() {

	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal("load key pair error:", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2"},
	}

	l, err := tls.Listen("tcp", ":3000", tlsCfg)

	if err != nil {
		log.Fatal("listen for tls over tcp error:", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept connection error:", err)
		}

		const preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
		b := make([]byte, len(preface))
		if _, err := io.ReadFull(conn, b); err != nil {
			log.Fatal("read from conn error:", err)
		}
		if string(b) != preface {
			log.Fatal("invalid preface:", string(b))
		}

		framer := http2.NewFramer(conn, conn)

		frames, err := readFrames(framer)
		if err != nil {
			log.Fatal("read frames error:", err)
		}
		var streamID uint32
		var data []byte
		for _, frame := range frames {
			if headersframe, ok := frame.(*http2.HeadersFrame); ok {
				streamID = headersframe.StreamID
			}
			if headersframe, ok := frame.(*http2.DataFrame); ok {
				data = headersframe.Data()
			}
		}
		fmt.Printf("StreamID: %v, Message: %s\n", streamID, string(data))

		framer.WriteRawFrame(http2.FrameSettings, 0, 0, []byte{})

		readFrames(framer)

		pic, err := ioutil.ReadFile("image.png")
		if err != nil {
			log.Fatal(err)
		}
		hbuf := bytes.NewBuffer([]byte{})
		henc := hpack.NewEncoder(hbuf)
		henc.WriteField(hpack.HeaderField{Name: ":status", Value: "200"})
		henc.WriteField(hpack.HeaderField{Name: "content-length", Value: strconv.Itoa(len(pic))})
		henc.WriteField(hpack.HeaderField{Name: "content-type", Value: "image/png"})

		fmt.Printf("Encoded Header: %d Byte\n", len(hbuf.Bytes()))

		err = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: streamID, BlockFragment: hbuf.Bytes(), EndHeaders: true})
		if err != nil {
			log.Fatal("write headers error: ", err)
		}

		for _, chunk := range chunkBy(pic, 1024) {
			framer.WriteData(streamID, false, chunk)
		}
		framer.WriteData(streamID, true, []byte{})
	}
}

func readFrames(framer *http2.Framer) ([]http2.Frame, error) {
	frames := make([]http2.Frame, 0)
	for {
		frame, err := framer.ReadFrame()
		if err != nil {
			return frames, err
		}
		frames = append(frames, frame)
		if frame.Header().Flags.Has(http2.FlagDataEndStream) {
			return frames, nil
		}
	}
}

func chunkBy(items []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}
