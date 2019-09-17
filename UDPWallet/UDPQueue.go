package UDPWallet

import (
	"encoding/json"
	"fmt"
	"github.com/Iuduxras/pangolin-node-4g-udp/service/rpcMsg"
	"github.com/Iuduxras/pangolin-node-4g-udp/utils"
	"net"
)

/*this message queue is for receiving udp message from node
 */

const fixedServerPort = 50998
const fixedClientPort = 50997

type QueueConfig struct {
	NodeAddr string
	NodeIP   string
}

type Queue struct {
	queue chan *rpcMsg.UDPRes
	done  chan bool
	conf  *QueueConfig
}

type CmdHandler interface {
	HandleRequireServiceRes(accepted bool, credit int64, msg string)
	HandleChargeRes(number int)
}

//listen and receive msg, put msg in queue
func (q *Queue) Receiving() {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: fixedClientPort})
	if err != nil {
		panic(err)
		return
	}
	data := make([]byte, 1024)
	go func() {
		for {
			select {
			case <-q.done:
				fmt.Println("done receiving")
				return
			default:
				n, remoteAddr, err := listener.ReadFromUDP(data)
				if err != nil {
					fmt.Printf("error during read: %s", err)
					continue
				}
				if remoteAddr.IP.String() != q.conf.NodeIP {
					fmt.Println("received message ip:" + remoteAddr.IP.String() + " differ from node ip:" + q.conf.NodeIP)
					continue
				}
				res := &rpcMsg.UDPRes{}
				json.Unmarshal(data[:n], res)
				q.queue <- res
			}

		}
	}()
}

func (q *Queue) Handle(handler CmdHandler) {
	go func() {
		for {
			res, more := <-q.queue
			if more {
				switch res.CmdType {
				case rpcMsg.CmdRequireService:
					q.receiveResCmdRequireService(res, handler)
				case rpcMsg.CmdRecharge:
					q.receiveResCmdRecharge(res, handler)
				default:
					fmt.Printf("received unknown cmd :%v", res)
				}
			} else {
				fmt.Println("done handling")
				return
			}
		}
	}()
}

func (q *Queue) close() {
	q.done <- true
	close(q.queue)
}

func (q *Queue) VerifyRes(res *rpcMsg.UDPRes) bool {
	//compare nodeAddr of res with nodeAddr record in QConf, and run verify to make sure
	//the msg come from the same 4G node
	if res.NodeAddr != q.conf.NodeAddr {
		fmt.Println("res nodeAddr: " + res.NodeAddr + " do not match with previous record: " + q.conf.NodeAddr)
		return false
	}
	if !res.Verify() {
		fmt.Println("reciveResCmd verify failed")
		return false
	}
	return true
}

func (q *Queue) receiveResCmdRequireService(res *rpcMsg.UDPRes, handler CmdHandler) {
	if q.VerifyRes(res) {
		//unpack msg
		credit := rpcMsg.CreditOnNode{}
		if err := json.Unmarshal(res.Msg, credit); err != nil {
			fmt.Printf("unmarshal error: %v\n", err)
		}
		handler.HandleRequireServiceRes(credit.Accepted, credit.Credit, credit.Msg)
	}
}

func (q *Queue) receiveResCmdRecharge(res *rpcMsg.UDPRes, handler CmdHandler) {
	if q.VerifyRes(res) {
		//unpack msg
		chargeNum := utils.BytesToInt(res.Msg)
		handler.HandleChargeRes(chargeNum)
	}
}

//send is not through queue
func (q *Queue) Send(req *rpcMsg.UDPReq) error {
	ip := net.ParseIP(q.conf.NodeIP)
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: fixedServerPort}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	data, _ := json.Marshal(req)
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}
