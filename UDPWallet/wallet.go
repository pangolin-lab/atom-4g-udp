package UDPWallet

import (
	"encoding/json"
	"fmt"
	"github.com/Iuduxras/pangolin-node-4g-udp/service/rpcMsg"
	"github.com/Iuduxras/pangolin-node-4g-udp/utils"
	"github.com/Iuduxras/pangolin-node-4g/account"
	"golang.org/x/crypto/ed25519"
	"net"
	"sync"
	"time"
)

const fixedServerPort = 50998
const fixedClientSendPort = 50996

var once sync.Once

type WConfig struct {
	BCAddr   string
	Cipher   string
	Ip       string
	Mac      string
	NodeAddr string
	NodeIP   string
}

type Wallet struct {
	acc    *account.Account
	Config *WConfig
	Queue  *Queue
	conn   *net.UDPConn
}

type Queue struct {
	queue chan *rpcMsg.UDPRes
	done  chan bool
}

type CmdHandler interface {
	HandleRequireServiceRes(accepted bool, credit int64, msg string)
	HandleChargeRes(number int)
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
			NodeIP: serverIp,
		},
		Queue: &Queue{
			queue: make(chan *rpcMsg.UDPRes, 64),
			done:  make(chan bool),
		},
	}
	return w, nil
}

func (w *Wallet) TestConnection() error {
	ip := net.ParseIP(w.Config.NodeIP)
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: fixedClientSendPort}
	dstAddr := &net.UDPAddr{IP: ip, Port: fixedServerPort}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return err
	}
	w.conn = conn
	if err := w.SendCmdCheck(); err != nil {
		panic(err)
	}
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			panic(err)
		}
		res := &rpcMsg.UDPRes{}
		json.Unmarshal(data[:n], res)
		if res.CmdType == rpcMsg.CmdCheck {
			w.Config.NodeAddr = res.NodeAddr
			break
		}
	}
	return nil
}

func (w *Wallet) Open(handler CmdHandler) {
	//first we have to get nodeAddr, just like handshake
	if err := w.TestConnection(); err != nil {
		panic(err)
	}
	fmt.Println("connect node:" + w.Config.NodeAddr + " successfully, start serving")
	w.Receiving()
	w.Handle(handler)
}

func (w *Wallet) Close() {
	defer func() {
		recover()
	}()
	w.conn.Close()
	w.Queue.done <- true
	close(w.Queue.queue)
}

func (w *Wallet) Send(req *rpcMsg.UDPReq) error {
	data, _ := json.Marshal(req)
	if _, err := w.conn.Write(data); err != nil {
		return err
	}
	fmt.Printf("send cmd type-> %s\n", rpcMsg.TranslateCmd(req.CmdType))
	return nil
}

//listen and receive msg, put msg in queue
func (w *Wallet) Receiving() {
	fmt.Println("get udp listener")
	data := make([]byte, 1024)
	go func() {
		fmt.Println("listening")
		for {
			select {
			case <-w.Queue.done:
				fmt.Println("done receiving")
				return
			default:
				n, remoteAddr, err := w.conn.ReadFromUDP(data)
				if err != nil {
					fmt.Printf("error during read: %s", err)
					continue
				}
				if remoteAddr.IP.String() != w.Config.NodeIP {
					fmt.Println("received message ip:" + remoteAddr.IP.String() + " differ from node ip:" + w.Config.NodeIP)
					continue
				}
				res := &rpcMsg.UDPRes{}
				json.Unmarshal(data[:n], res)
				fmt.Println("Received msg of type" + rpcMsg.TranslateCmd(res.CmdType)+" push in queue")
				w.Queue.queue <- res
			}
		}
	}()
}

func (w *Wallet) SendCmdCheck() error {
	req := &rpcMsg.UDPReq{}
	req.AsCmdCheck(w.Config.BCAddr)
	if err := w.Send(req); err != nil {
		return err
	}
	return nil
}

func (w *Wallet) SendCmdRequireService() error {
	req := &rpcMsg.UDPReq{}
	if err := req.AsCmdRequireService(&rpcMsg.SevReqData{
		Addr: w.Config.BCAddr,
		Ip:   w.Config.Ip,
		Mac:  w.Config.Mac,
	}, w.acc.Key.PriKey); err != nil {
		return err
	}
	if err := w.Send(req); err != nil {
		return err
	}
	return nil
}

func (w *Wallet) SendCmdRecharge(no int) error {
	bill, err := CreatePayBill(string(w.acc.Address), w.Config.NodeAddr, no, w.acc.Key.PriKey)
	if err != nil {
		return err
	}
	fmt.Printf("Create new packet bill:%s for miner:%s", w.Config.NodeAddr, bill.String())
	req := &rpcMsg.UDPReq{}
	if err := req.AsCmdRecharge(bill, w.Config.BCAddr, w.acc.Key.PriKey); err != nil {
		return err
	}
	if err := w.Send(req); err != nil {
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

func (w *Wallet) Handle(handler CmdHandler) {
	for {
		res, more := <-w.Queue.queue
		fmt.Println("handling msg of type: "+ rpcMsg.TranslateCmd(res.CmdType))
		if more {
			switch res.CmdType {
			case rpcMsg.CmdRequireService:
				w.receiveResCmdRequireService(res, handler)
			case rpcMsg.CmdRecharge:
				w.receiveResCmdRecharge(res, handler)
			default:
				fmt.Printf("received unknown cmd :%v", res)
			}
		} else {
			fmt.Println("done handling")
			return
		}
	}
}

func (w *Wallet) VerifyRes(res *rpcMsg.UDPRes) bool {
	//compare nodeAddr of res with nodeAddr record in QConf, and run verify to make sure
	//the msg come from the same 4G node
	if res.NodeAddr != w.Config.NodeAddr {
		fmt.Println("res nodeAddr: " + res.NodeAddr + " do not match with previous record: " + w.Config.NodeAddr)
		return false
	}
	if !res.Verify() {
		fmt.Println("reciveResCmd verify failed")
		return false
	}
	return true
}

func (w *Wallet) receiveResCmdRequireService(res *rpcMsg.UDPRes, handler CmdHandler) {
	if w.VerifyRes(res) {
		//unpack msg
		credit := &rpcMsg.CreditOnNode{}
		if err := json.Unmarshal(res.Msg, credit); err != nil {
			fmt.Printf("unmarshal error: %v\n", err)
		}
		fmt.Println("cmd require service handler params: ",credit.Accepted, credit.Credit, credit.Msg)
		handler.HandleRequireServiceRes(credit.Accepted, credit.Credit, credit.Msg)
	}
}

func (w *Wallet) receiveResCmdRecharge(res *rpcMsg.UDPRes, handler CmdHandler) {
	if w.VerifyRes(res) {
		//unpack msg
		chargeNum := utils.BytesToInt(res.Msg)
		fmt.Println("cmd recharge handler params: ",chargeNum)
		handler.HandleChargeRes(chargeNum)
	}
}
