package Client

import (
	. "../Log"
	"../Pack"
	"bufio"
	"io"
	"net"
	"time"
)

type Connector struct {
	addr       string
	conn       net.Conn
	readWriter *bufio.ReadWriter
	refresh    chan string
	stat       bool
}

func (connector *Connector) Handle(conn net.Conn) (err error) {
	connector.conn = conn
	connector.init()
	defer func() {
		_ = connector.clearReadBuffer()
		_ = connector.readWriter.Flush()
		_ = conn.Close()
	}()
	err = connector.checkAccess()
	if err != nil {
		return err
	}
	connector.preAction()
	defer connector.postAction()
	go connector.connectionHeartBeats()
	for connector.stat {
		connector.loop()
	}
	return err
}

func (connector *Connector) init() {
	connector.addr = connector.conn.RemoteAddr().String()
	connector.readWriter = bufio.NewReadWriter(bufio.NewReader(connector.conn), bufio.NewWriter(connector.conn))
	connector.stat = true
	connector.refresh = make(chan string, 1)

	connector.extraInit()
}

func (connector *Connector) extraInit() {}

func (connector *Connector) checkAccess() error { return nil }

func (connector *Connector) preAction() {}

func (connector *Connector) postAction() {}

func (connector *Connector) loop() {}

func (connector *Connector) connectionHeartBeats() {
	for {
		select {
		case <-connector.refresh:
			if !connector.stat {
				return
			}
		case <-time.After(time.Minute):
			connector.stat = false
			break
		}
	}
}

func (connector *Connector) writeRepeat(packet Pack.Packet, t time.Duration) (err error) {
	var ch = make(chan string, 1)
	go func() {
		var count int
		for count < 3 && len(ch) < 1 {
			_ = connector.conn.SetWriteDeadline(time.Now().Add(t))
			_, err = connector.readWriter.WriteString(string(packet))
			if err != nil {
				Log.Println(err)
			}
			err = connector.readWriter.Flush()
			if err != nil && err != io.EOF {
				Log.Println(err)
				count++
			} else {
				break
			}
		}
		ch <- ""
		_ = connector.conn.SetWriteDeadline(time.Time{})
	}()
	select {
	case <-ch:
		connector.refresh <- ""
		_ = connector.conn.SetReadDeadline(time.Time{})
		return nil
	case <-time.After(t):
		ch <- ""
		_ = connector.conn.SetReadDeadline(time.Time{})
		defer func() {
			time.Sleep(time.Second)
			<-ch
		}()
		return io.EOF
	}
}

func (connector *Connector) clearReadBuffer() error {
	var n = connector.readWriter.Reader.Buffered()
	var _, err = connector.readWriter.Discard(n)
	return err
}
