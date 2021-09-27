package main

import (
	"context"
	"log"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type poloniex struct {
	conn net.Conn
}

func newPoloniexConnectWs(ctx context.Context, url string) (*poloniex, error) {
	conn, _, _, err := ws.DefaultDialer.Dial(ctx, poloniexWSUrl)
	if err != nil {
		return nil, err
	}

	p := new(poloniex)
	p.conn = conn

	return p, nil
}

func (p *poloniex) doRead(ctx context.Context, parseCh chan<- []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg, _, err := wsutil.ReadServerData(p.conn)
		if err != nil {
			log.Printf("read server data error: %v", err)
			continue
		}
		parseCh <- msg
	}
}

func (p *poloniex) sendCommand(cmd []byte) error {
	err := wsutil.WriteClientMessage(p.conn, ws.OpText, cmd)
	log.Printf("sending cmd: %v", string(cmd))
	if err != nil {
		return err
	}

	return nil
}

func (p *poloniex) close() error {
	return p.conn.Close()
}