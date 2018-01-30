package main

import (
    "bytes"
    "compress/flate"
    "log"
	"net"
	"bufio"
)

const resp = `<html><body>A website (with custom TCP server) in 1 IPv4 datagram!
<br/><br/>
To make the biggest-smallest site, we consider the most limiting step among TCP segments, IP datagrams, Ethernet frames, HTTP response...
<br/><br/>
In <a href="https://tools.ietf.org/html/rfc6691">RFC 6691</a>, the max size (in bytes) for a TCP segment is the IP
max datagram size (576) minus 40 for TCP&IP headers, leaving 536 for HTTP & content.
<br/><br/>
In <a href="https://www.w3.org/Protocols/rfc2616/rfc2616-sec6.html">RFC 2616</a> the minimal valid HTTP 1.1 response with DEFLATE compression
is: HTTP/1.1 200 OK\r\nContent-Encoding: deflate\r\n\r\n: 46 bytes for HTTP leaves 490 for content.
<br/><br/>
This text & HTML is 481 bytes, for 527 total bytes! Visit <a href="https://aadrake.com/">aadrake.com</a> for more!
</body></html>`

func main() {
    ip := net.IPv4(0, 0, 0, 0)
    tcpaddr := net.TCPAddr{IP: ip, Port: 80}
    ln, _ := net.ListenTCP("tcp", &tcpaddr)
	httpResponse := []byte("HTTP/1.1 200 OK\r\nContent-Encoding: deflate\r\n\r\n")
	limit := 536  // This is the limit of bytes we can send in one datagram

    var buf bytes.Buffer
    zw, err := flate.NewWriter(&buf, flate.BestCompression)
    if err != nil {
        log.Fatal(err)
    }
    _, err = zw.Write([]byte(resp))
    if err != nil {
        log.Fatal(err)
    }
    zw.Close()
    compressedBody := buf.Next(buf.Len())

    if len(httpResponse) + len(compressedBody) > limit {
		log.Fatalf("too many bytes: %v\n", len(append(httpResponse, compressedBody...)))
    }

    for {
        conn, _ := ln.AcceptTCP()
        conn.SetNoDelay(true) // Silly windows are silly.  See Nagling.
        go func(){
            defer conn.Close()
			_, _, _ = bufio.NewReader(conn).ReadLine()  // This is used in the stdlib HTTP server
            conn.Write(append(httpResponse, compressedBody...))
        }()
    }
}
