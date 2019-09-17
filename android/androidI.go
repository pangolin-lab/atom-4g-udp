package androidLib

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

type Handler interface {
	UDPWallet.CmdHandler
}


//consumer setup
func InitConsumer(addr, cipher, url, ip,mac ,serverIp, password string){
	w,err := UDPWallet.NewWallet(addr,cipher,ip,mac ,serverIp,password)
	if err!=nil{
		fmt.Println("init wallet failed")
		panic(err)
	}
	wallet = w
}


func Consuming(handler Handler){
	wallet.Open(handler)
}

func StopConsuming(){
	wallet.Close()
}

func Query(){
	wallet.SendCmdRequireService()
}

func Recharge(no int){
	wallet.SendCmdRecharge(no)
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
