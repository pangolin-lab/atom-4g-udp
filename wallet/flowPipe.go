package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/Iuduxras/pangolin-node-4g/network"
	"github.com/Iuduxras/pangolin-node-4g/service/rpcMsg"
	"net"
	"time"
)

type ControlPipe struct {
	reportChan  chan int
	requestBuf  []byte
	responseBuf []byte
	proxyConn   net.Conn
	consume     *network.PipeConn
}

func NewPipe(l net.Conn, r *network.PipeConn, rc chan int, tgt string) *ControlPipe {
	return &ControlPipe{
		requestBuf:  make([]byte, network.BuffSize),
		responseBuf: make([]byte, network.BuffSize),
		proxyConn:   l,
		consume:     r,
		reportChan:  rc,
	}
}


func (p *ControlPipe) expire() {
	p.consume.SetDeadline(time.Now())
	p.proxyConn.SetDeadline(time.Now())
}

func (p *ControlPipe) String() string {
	return fmt.Sprintf("%s<->%s for (%s)",
		p.proxyConn.RemoteAddr().String(),
		p.consume.RemoteAddr().String())
}


func (w *Wallet) pipeHandshake(conn *network.JsonConn, target string) error {

	reqData := &rpcMsg.SevReqData{
		Addr: w.acc.Address.String(),
	}


	data, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("marshal hand shake data err:%v", err)
	}

	sig := w.acc.Sign(data)

	hs := &rpcMsg.BYHandShake{
		CmdType: rpcMsg.CmdPipe,
		Sig:     sig,
		Sev:    reqData,
	}

	if err := conn.WriteJsonMsg(hs); err != nil {
		return fmt.Errorf("write hand shake data err:%v", err)

	}
	ack := &network.ProtonACK{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		return fmt.Errorf("failed to read miner's response :->%v", err)
	}

	if !ack.Success {
		return fmt.Errorf("hand shake to miner err:%s", ack.Message)
	}



	return nil
}
