package UDPWallet

import (
	"encoding/json"
	"fmt"
	"github.com/Iuduxras/pangolin-node-4g-udp/service/rpcMsg"
	"github.com/Iuduxras/pangolin-node-4g/account"
	"golang.org/x/crypto/ed25519"
	"sync"
	"time"
)

var once sync.Once

type WConfig struct {
	BCAddr string
	Cipher string
	Ip     string
	Mac    string
}

type Wallet struct {
	acc    *account.Account
	Config *WConfig
	Queue  *Queue
}

func NewWallet(addr, cipher, ip, mac, serverIp, password string) (*Wallet, error) {
	acc, err := account.AccFromString(addr, cipher, password)
	if err != nil {
		return nil, err
	}
	w := &Wallet{
		acc: acc,
		Config: &WConfig{
			BCAddr: addr,
			Cipher: cipher,
			Ip:     ip,
			Mac:    mac,
		},
		Queue: &Queue{
			queue: make(chan *rpcMsg.UDPRes, 64),
			done: make(chan bool),
			conf: &QueueConfig{
				NodeAddr: "",
				NodeIP:   serverIp,
			},
		},
	}
	//first we have to get nodeAddr, just like handshake
	if err := w.SendCmdCheck(); err != nil {
		panic(err)
	}
	for {
		m := <-w.Queue.queue
		if m.CmdType == rpcMsg.CmdCheck {
			w.Queue.conf.NodeAddr = m.NodeAddr
			break
		}
	}
	return w, nil
}


func (w *Wallet)Open(handler CmdHandler){
	fmt.Println("wallet open for consuming")
	w.Queue.Receving()
	w.Queue.Handle(handler)
}

func (w *Wallet)Close(){
	w.Queue.close()
}

func (w *Wallet) SendCmdCheck() error {
	req := &rpcMsg.UDPReq{}
	req.AsCmdCheck(w.Config.BCAddr)
	if err := w.Queue.Send(req);err!=nil{
		return err
	}
	return nil
}

func (w *Wallet) SendCmdRequireService() error {
	req:=&rpcMsg.UDPReq{}
	if err:=req.AsCmdRequireService(&rpcMsg.SevReqData{
		Addr: w.Config.BCAddr,
		Ip:   w.Config.Ip,
		Mac:  w.Config.Mac,
	},w.acc.Key.PriKey);err!=nil{
		return err
	}
	if err := w.Queue.Send(req);err!=nil{
		return err
	}
	return nil
}

func (w *Wallet) SendCmdRecharge(no int) error{
	bill, err := CreatePayBill(string(w.acc.Address), w.Queue.conf.NodeAddr, no, w.acc.Key.PriKey)
	if err != nil {
		return err
	}
	fmt.Printf("Create new packet bill:%s for miner:%s", w.Queue.conf.NodeAddr, bill.String())
	req:=&rpcMsg.UDPReq{}
	if err:=req.AsCmdRecharge(bill,w.Config.BCAddr,w.acc.Key.PriKey);err!=nil{
		return err
	}
	if err := w.Queue.Send(req);err!=nil{
		return err
	}
	return nil
}

func CreatePayBill(user, miner string, usage int, priKey ed25519.PrivateKey) (*rpcMsg.UserCreditPay, error) {
	pay := &rpcMsg.CreditPayment{
		UserAddr:    user,
		MinerAddr:   miner,
		PacketUsage: usage,
		PayTime:     time.Now(),
	}

	data, err := json.Marshal(pay)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(priKey, data)

	return &rpcMsg.UserCreditPay{
		UserSig:       sig,
		CreditPayment: pay,
	}, nil
}