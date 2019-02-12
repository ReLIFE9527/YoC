package Client

import (
	. "../Log"
	"../Pack"
	"bufio"
	"io"
	"net"
	"strings"
	"time"
)

type Connector struct {
	addr       string
	conn       net.Conn
	readWriter *bufio.ReadWriter
}

func (connector *Connector) handleConnection(conn net.Conn) (err error) { return err }

func (connector *Connector) connectionHeartBeats(flag *bool, actionRefresh chan string) {
	for {
		select {
		case <-actionRefresh:
			if *flag {
				return
			}
		case <-time.After(time.Minute):
			*flag = true
			break
		}
	}
}

func (connector *Connector) writeRepeat(packet Pack.Packet, t time.Duration) (err error) {
	var ch = make(chan string, 1)
	go func() {
		_, _ = connector.readWriter.WriteString(string(packet))
		err = connector.readWriter.Flush()
		if err != nil {
			Log.Println(err)
		}
		var count int
	repeat:
		_ = connector.conn.SetReadDeadline(time.Now().Add(t))
		str, err := connector.readWriter.ReadString(Pack.TailByte)
		stream, _ := Pack.DePackString(packet)
		if strings.Contains(str, "done") && Pack.IsStreamValid([]string{"read"}, stream) {
			count += 2
		}
		count++
		if err != nil && err != io.EOF && count < 3 && len(ch) < 1 {
			goto repeat
		}
		ch <- ""
	}()
	select {
	case <-ch:
		_ = connector.conn.SetReadDeadline(time.Time{})
		return nil
	case <-time.After(t):
		ch <- ""
		_ = connector.conn.SetReadDeadline(time.Time{})
		return io.EOF
	}
}
