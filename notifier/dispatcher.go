package notifier

import (
	"bufio"
	"encoding/json"
	"net"
)

type Handler func(tip, title, info string)

type Message struct {
	Tooltip string `json:"tooltip"`
	Title   string `json:"title"`
	Info    string `json:"info"`
}

type Transport interface {
	Observe(address string, handler Handler) error
	Stop() error
}

type SocketTransport struct {
	running bool
	conn    net.Conn
}

func (st *SocketTransport) Observe(address string, handler Handler) error {
	st.running = true
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	st.conn = conn
	buf := bufio.NewReader(st.conn)
	for {
		str, err := buf.ReadString('\n')
		if !st.running {
			return nil
		}
		if len(str) > 0 {
			if err := st.handleLine(str, handler); err != nil {
				continue
			}
		}
		if err != nil {
			return err
		}
	}
}

func (st *SocketTransport) Stop() error {
	st.running = false
	return st.conn.Close()
}

func (st SocketTransport) handleLine(s string, handler Handler) error {
	var m Message
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return err
	}
	handler(m.Tooltip, m.Title, m.Info)
	return nil
}
