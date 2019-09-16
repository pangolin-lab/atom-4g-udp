package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/Iuduxras/pangolin-node-4g-udp/account"
	"github.com/Iuduxras/pangolin-node-4g-udp/network"
	"github.com/Iuduxras/pangolin-node-4g-udp/service/rpcMsg"
	"golang.org/x/crypto/ed25519"
	"log"
	"net"
	"sync"
	"syscall"
	"time"
)

type WConfig struct {
	BCAddr     string
	Cipher     string
	SettingUrl string
	ServerId   *ServeNodeId
	Saver      func(fd uintptr)
	Ip         string
	Mac        string
}

var CreditLocal int64

const MaxLocalConn = 1 << 10
const PipeDialTimeOut = time.Second * 2
const RechargeTimeInterval = time.Minute * 5

func (c *WConfig) String() string {
	return fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++\n"+
		"+\t BCAddr:%s\n"+
		"+\t Ciphere:%s\n"+
		"+\tSettingUrl:%s\n"+
		"+\tServerId:%s\n"+
		"++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
		c.BCAddr,
		c.Cipher,
		c.SettingUrl,
		c.ServerId.String())
}

type PacketBucket struct {
	sync.RWMutex
	token  chan int
	//unpaid int    //unpaid is removed because it should be handled out of library
}

type Wallet struct {
	*PacketBucket
	*InternetAddress
	acc          *account.Account
	sysSaver     func(fd uintptr)
	payConn      *network.JsonConn
	checkConn    *network.JsonConn
	aesKey       account.PipeCryptKey
	minerID      account.ID
	minerAddr    []byte
	minerNetAddr string
}

type InternetAddress struct {
	IP  string
	Mac string
}

func NewWallet(conf *WConfig, password string) (*Wallet, error) {
	acc, err := account.AccFromString(conf.BCAddr, conf.Cipher, password)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n Unlock client success:%s Selected miner id:%s",
		conf.BCAddr, conf.ServerId.String())
	w := &Wallet{
		acc:          acc,
		minerID:      conf.ServerId.ID,
		sysSaver:     conf.Saver,
		minerNetAddr: NetAddrFixedPort(conf.ServerId.IP),
		PacketBucket: &PacketBucket{
			token: make(chan int, MaxLocalConn),
		},
		InternetAddress: &InternetAddress{
			IP:  conf.Ip,
			Mac: conf.Mac,
		},
	}
	w.minerAddr = make([]byte, len(w.minerID))
	copy(w.minerAddr, []byte(w.minerID))
	//TODO:: to be checked
	if err := w.acc.Key.GenerateAesKey(&w.aesKey, w.minerID.ToPubKey()); err != nil {
		return nil, err
	}

	if err := w.setPayChannel(); err != nil {
		log.Println("Create payment channel err:", err)
		return nil, err
	}

	if err := w.setCheckChannel(); err != nil {
		log.Println("Create check channel err:", err)
		return nil, err
	}
	fmt.Printf("\nCreate payment channel and check channel success:%s", w.ToString())
	return w, nil
}


//pay channel
func (w *Wallet) setPayChannel() error {
	fmt.Printf("\ncreatePayChannel Wallet socks ID addr:%s ", w.minerNetAddr)
	conn, err := getOuterConn(w, w.minerNetAddr)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(w.acc.Key.PriKey, []byte(w.acc.Address))
	hs := &rpcMsg.BYHandShake{
		CmdType:  rpcMsg.CmdRecharge,
		Sig:      sig,
		UserAddr: w.acc.Address.String(),
	}
	jsonConn := &network.JsonConn{Conn: conn}
	if err := jsonConn.Syn(hs); err != nil {
		return err
	}
	w.payConn = jsonConn
	return nil
}

//check channel (require service)
func (w *Wallet) setCheckChannel() error {
	fmt.Printf("\ncreateCheckChannel Wallet socks ID addr:%s ", w.minerNetAddr)
	conn, err := getOuterConn(w, w.minerNetAddr)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(w.acc.Key.PriKey, []byte(w.acc.Address))
	hs := &rpcMsg.BYHandShake{
		CmdType:  rpcMsg.CmdRequireService,
		Sig:      sig,
		UserAddr: w.acc.Address.String(),
	}
	jsonConn := &network.JsonConn{Conn: conn}
	if err := jsonConn.Syn(hs); err != nil {
		return err
	}
	w.checkConn = jsonConn
	return nil
}

func (w *Wallet) Finish() {
	w.payConn.Close()
	w.checkConn.Close()
}

func getOuterConn(w *Wallet, addr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: PipeDialTimeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			if w.sysSaver != nil {
				return c.Control(w.sysSaver)
			}
			return nil
		},
	}
	return d.Dial("tcp", addr)
}

func GetOuterConnSimple(addr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: PipeDialTimeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			return nil
		},
	}
	return d.Dial("tcp", addr)
}

func (w *Wallet) ToString() string {
	return fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++\n"+
		"+\t account:%s\n"+
		"+\t minerID:%s\n"+
		"+\t Address:%s\n"+
		"++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
		w.acc.Address,
		string(w.minerID),
		w.minerNetAddr)
}

func (w *Wallet) Running(done chan error) {

	//loop:
	//for {
	//	select {
	//	case err := <-done:
	//		fmt.Printf("\nwallet closed by out controller:%s", err.Error())
	//		break loop
	//	}
	//}
}

func (w *Wallet) Query() (string,error) {
	srv := &rpcMsg.SevReqData{
		Addr: w.acc.Address.String(),
		Ip:   w.IP,
		Mac:  w.Mac,
	}

	data, err := json.Marshal(srv)
	if err != nil {
		return "",err
	}
	sig := ed25519.Sign(w.acc.Key.PriKey, data)
	request := &rpcMsg.CreditQuery{
		Sig:        sig,
		SevReqData: srv,
	}
	if d,err := w.checkConn.SynResRaw(request); err != nil {

		return "",err
	}else{
		fmt.Printf("query answer is : %s",d)
		return d,nil
	}
}


func (w *Wallet) Recharge(no int) error {

	minerAddr := string(w.minerAddr)
	bill, err := CreatePayBill(string(w.acc.Address), minerAddr, no, w.acc.Key.PriKey)
	if err != nil {
		return err
	}

	fmt.Printf("Create new packet bill:%s for miner:%s", minerAddr, bill.String())

	if err := w.payConn.Syn(bill); err != nil {
		fmt.Printf("\nwallet write back bill msg err:%v", err)
		return err
	}

	fmt.Printf("recharge success:%d", no)
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

