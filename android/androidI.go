package androidLib

//package main

import "C"
import (
	"fmt"
	"github.com/Iuduxras/atom-4g-udp/UDPWallet"
	"github.com/Iuduxras/atom-4g-udp/ethereum"
	"github.com/Iuduxras/pangolin-node-4g-udp/account"
	"github.com/btcsuite/btcutil/base58"
)

const Separator = "@@@"

var wallet *UDPWallet.Wallet
var isOpen *bool

func init() {
	f := false
	isOpen = &f
}

type Handler interface {
	UDPWallet.CmdHandler
}

//consumer setup
func InitWallet(addr, cipher, ip, mac, serverIp, password string) bool {
	w, err := UDPWallet.NewWallet(addr, cipher, ip, mac, serverIp, password)
	if err != nil {
		fmt.Println("init wallet failed")
		panic(err)
	}
	wallet = w
	if err := w.TestConnection(); err != nil {
		wallet = nil
		f := false
		isOpen = &f
		fmt.Printf("init fail error : %v\n",err)
		return false
	} else {
		return true
	}
}

func Communicating(handler Handler) {
	t := true
	isOpen = &t
	fmt.Printf("wallet status : %v\n", *isOpen)
	wallet.Open(handler)
}

func StopConsuming() {
	fmt.Println("closing")
	if *isOpen {
		wallet.SendCmdClose()
	}
	f := false
	isOpen = &f
	wallet.Close()
	fmt.Println("wallet closed")
}

func StopTrail(){
	fmt.Println("stop trail")
	f := false
	isOpen = &f
	wallet.Close()
	fmt.Println("wallet closed")
}

func ApplyTrail() {
	if err := wallet.SendCmdTrail(); err != nil {
		fmt.Printf("error sending apply for trail %v\n", err)
	}
}

func Query() {
	if *isOpen {
		if err := wallet.SendCmdRequireService(); err != nil {
			fmt.Printf("send require serivce error: %v\n", err)
		}
	} else {
		fmt.Println("wallet is closed")
	}

}

func Recharge(no int) {
	if *isOpen {
		if err := wallet.SendCmdRecharge(no); err != nil {
			fmt.Printf("send recharge error: %v\n", err)
		}
	} else {
		fmt.Println("wallet is closed")
	}
}

func VerifyAccount(addr, cipher, password string) bool {
	if _, err := account.AccFromString(addr, cipher, password); err != nil {
		fmt.Println("Valid Account:", err)
		return false
	}
	return true
}

func CreateAccount(password string) string {

	key, err := account.GenerateKey(password)
	if err != nil {
		return ""
	}
	address := key.ToNodeId().String()
	cipherTxt := base58.Encode(key.LockedKey)

	return address + Separator + cipherTxt
}

func IsProtonAddress(address string) bool {
	return account.ID(address).IsValid()
}

func LoadEthAddrByProtonAddr(protonAddr string) string {
	return ethereum.BoundEth(protonAddr)
}

func EthBindings(ETHAddr string) string {
	ethB, no := ethereum.BasicBalance(ETHAddr)
	if ethB == nil {
		return ""
	}

	return fmt.Sprintf("%f"+Separator+"%d",
		ethereum.ConvertByDecimal(ethB),
		no)
}

func CreateEthAccount(password, directory string) string {
	return ethereum.CreateEthAccount2(password, directory)
}

func VerifyEthAccount(cipherTxt, pwd string) bool {
	return ethereum.VerifyEthAccount(cipherTxt, pwd)
}

func BindProtonAddress(protonAddr, cipherKey, password string) string {
	tx, err := ethereum.Bind(protonAddr, cipherKey, password)
	if err != nil {
		fmt.Printf("\nBind proton addr(%s) err:%s", protonAddr, err)
		return err.Error()
	}
	return tx
}
func UnbindProtonAddress(protonAddr, cipherKey, password string) string {
	tx, err := ethereum.Unbind(protonAddr, cipherKey, password)
	if err != nil {
		fmt.Printf("\nBind proton addr(%s) err:%s", protonAddr, err)
		return err.Error()
	}
	return tx
}

//////////////////for test///////////////////
//
//type fakeHandler struct {
//
//}
//
//func (f *fakeHandler) HandleRequireServiceRes(accepted bool, credit int64, msg string){
//	fmt.Println(accepted,credit,msg)
//}
//func (f *fakeHandler) HandleChargeRes(number int){
//	fmt.Println(number)
//}
//
//func main(){
//	InitWallet(
//		"YPGmpwh8Ev4eKmBhTvidBqgUvk4sgNJqipvQShtfR7vVYk",
//		"4aLvNMdFyJy6wHsKZJMC1r2m6NzEBWu5sNPzqGhFyXhJwwY43unxijWGbKGZWqzJdZnvLSzqdtZqscVRHbz1hj5Yd9JdxG3wYv7FEqV57ZqNa",
//		"127.0.0.1",
//		"00:e0:4c:36:0a:2c",
//		"127.0.0.1",
//		"123")
//
//	handler := fakeHandler{}
//	wallet.Open(&handler)
//}
//////////////////////////////////////////////
