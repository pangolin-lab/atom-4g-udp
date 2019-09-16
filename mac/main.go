package main

import "C"
import (
	"encoding/base64"
	"fmt"
	"github.com/Iuduxras/atom-4g/Service4G"
	"github.com/Iuduxras/atom-4g/wallet"
	"github.com/Iuduxras/pangolin-node-4g/account"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

var proxyConfTest = &Service4G.ConsumerConfig{
	WConfig: &wallet.WConfig{
		BCAddr:     "YPDsDm5RBqhA14dgRUGMjE4SVq7A3AzZ4MqEFFL3eZkhjZ",
		Cipher:     "GffT4JanGFefAj4isFLYbodKmxzkJt9HYTQTKquueV8mypm3oSicBZ37paYPnDscQ7XoPa4Qgse6q4yv5D2bLPureawFWhicvZC5WqmFp9CGE",
		SettingUrl: "https://raw.githubusercontent.com/proton-lab/quantum/master/seed_debug.quantum",
		Saver:      nil,
	},
	BootNodes: "YPBzFaBFv8ZjkPQxtozNQe1c9CvrGXYg4tytuWjo9jiaZx@192.168.30.12",
}

func createKs() {
	ks := keystore.NewKeyStore("bin", keystore.StandardScryptN, keystore.StandardScryptP)
	password := "secret"
	account, err := ks.NewAccount(password)
	if err != nil {
		panic(err)
	}

	fmt.Println(account.Address.Hex()) // 0x20F8D42FB0F667F2E53930fed426f225752453b3
	fmt.Println(account.URL.Path)
	fmt.Println(account.URL.Scheme)
}

func importKs() {
	file := "bin/UTC--2019-07-15T12-05-57.402709000Z--48abf79312d973b55841e48fb1e4872953d43946"
	ks := keystore.NewKeyStore("bin", keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	password := "secret"
	account, err := ks.Import(jsonBytes, password, password)
	if err != nil {
		panic(err)
	}

	fmt.Println(account.Address.Hex()) // 0x20F8D42FB0F667F2E53930fed426f225752453b3

	if err := os.Remove(file); err != nil {
		panic(err)
	}
}
func test12() {
	acc, err := account.AccFromString("YPDV86j2ZTnFivpC44FtpocyYgtqPJ5R5NC5EcRcyhprTs",
		"3U1V26zuBSgW6mudv7aZACkK8q75XEf936qWfhRRvKEHqTrQmmk726464tRnSLXYPUgqyvWADG4DPtqE3Y2Va4qo9ivvRTbz2jnikpdhj6Feuz", "123")
	if err != nil {
		panic(err)
	}
	print(acc)
}

func main() {
	test4G()
}

func test4Glocal(){
	var conf = &wallet.WConfig{
		BCAddr:     "YPGmpwh8Ev4eKmBhTvidBqgUvk4sgNJqipvQShtfR7vVYk",
		Cipher:     "4aLvNMdFyJy6wHsKZJMC1r2m6NzEBWu5sNPzqGhFyXhJwwY43unxijWGbKGZWqzJdZnvLSzqdtZqscVRHbz1hj5Yd9JdxG3wYv7FEqV57ZqNa",
		SettingUrl: "",
		Ip:			"127.0.0.1",
		Mac: 		"00:e0:4c:36:0a:2c",
		Saver:      nil,
		ServerId: &wallet.ServeNodeId{
			ID: account.ID("YP6ywypy2P3dMRYCG2V1PZMCeuT8mUpgr1Sapo9XhhLRki"),
			IP: "127.0.0.1",
		},
	}
	w, err := wallet.NewWallet(conf, "123")
	if err != nil {
		panic(err)
	}

	proxy, e := Service4G.NewConsumer(":51080", w)
	if e != nil {
		panic(err)
	}

	proxy.Consuming()
}


func test4G(){
	var conf = &wallet.WConfig{
		BCAddr:     "YPGmpwh8Ev4eKmBhTvidBqgUvk4sgNJqipvQShtfR7vVYk",
		Cipher:     "4aLvNMdFyJy6wHsKZJMC1r2m6NzEBWu5sNPzqGhFyXhJwwY43unxijWGbKGZWqzJdZnvLSzqdtZqscVRHbz1hj5Yd9JdxG3wYv7FEqV57ZqNa",
		SettingUrl: "",
		Ip:			"172.16.100.163",
		Mac: 		"38:f9:d3:8c:21:f4",
		Saver:      nil,
		ServerId: &wallet.ServeNodeId{
			ID: account.ID("YPDFSEKYU3tYfpvxER3JDTqDMb4vB4SWqawTRmaa3jnnuA"),
			IP: "172.16.100.1",
		},
	}
	w, err := wallet.NewWallet(conf, "123")
	if err != nil {
		panic(err)
	}
	proxy, e := Service4G.NewConsumer(":51080", w)
	if e != nil {
		panic(err)
	}

	proxy.Query()

	proxy.Consuming()
}


func test10() {
	fmt.Println(publicsuffix.EffectiveTLDPlusOne("1-apple.com.tw"))
}

func test11() {

	//failed:
	//YPCr9KRE3tRXaKMb388A5gEjFqK3u4sAo9EBLK7tc94xwh
	//3HvcAKMmKT6hEEgpYo4Sf1TNRAewbZtyqTkopC9G4E6nv89vqkiq1ft5Rzf7pmim3b4ZxXaEu1bR8yGzJUM8865mNoX2FEkmaJsGKvSfHYMyu5
	//success:
	//YPDsDm5RBqhA14dgRUGMjE4SVq7A3AzZ4MqEFFL3eZkhjZ
	//GffT4JanGFefAj4isFLYbodKmxzkJt9HYTQTKquueV8mypm3oSicBZ37paYPnDscQ7XoPa4Qgse6q4yv5D2bLPureawFWhicvZC5WqmFp9CGE
	var conf = &wallet.WConfig{
		BCAddr:     "YPDsDm5RBqhA14dgRUGMjE4SVq7A3AzZ4MqEFFL3eZkhjZ",
		Cipher:     "GffT4JanGFefAj4isFLYbodKmxzkJt9HYTQTKquueV8mypm3oSicBZ37paYPnDscQ7XoPa4Qgse6q4yv5D2bLPureawFWhicvZC5WqmFp9CGE",
		SettingUrl: "",
		Saver:      nil,
		ServerId: &wallet.ServeNodeId{
			ID: account.ID("YP9K8VVHqLzi75tvDs3xLPm3FJ8mN623EKd7XpyNCuJW9D"),
			IP: "192.168.30.13",
		},
	}
	w, err := wallet.NewWallet(conf, "12345678")
	if err != nil {
		panic(err)
	}

	proxy, e := Service4G.NewConsumer(":51080", w)
	if e != nil {
		panic(err)
	}

	proxy.Consuming()
}

func test9() {
	tt, err := base64.StdEncoding.DecodeString("")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(tt))
}

func test8() {
	ip, subNet, _ := net.ParseCIDR("58.248.0.0/13")
	mask := subNet.Mask

	srcIP := net.ParseIP("58.251.82.180")
	maskIP := srcIP.Mask(mask)

	fmt.Println(ip.String(), maskIP.String())
	fmt.Println(string(maskIP), string(ip), maskIP.String() == ip.String(), subNet.Contains(srcIP))
}

func test7() {
	str := "zoFqdxIrIwcRWPyALfi7yCVvJagI6hE86K3KNc0ioPxsSJWqYa2A5QWTxfO8fUq5GyDJeCfOjnyNxZsFFmav2KE4z5FsoMeUIbNTjwiFMqeqzObr1JKJi+l/wybgKEfZ0ijbMGaynfEIWbFPlIKxYc1YkZdHcKzeG6yWNxXCtXEK1JJ7pbo9DRcaOWuj2xFBD/Dnasizc7fJOPnPy2JROHmlDyajxz/UavGjFNAmBh5iegAisNexrSoGihG/r5GiY9xP1wCP860nC3RWN6Sxzbb7fCZJvqKuXuPCm8d6KjyrXV7v0PPlrhFfekdviE0dg4f2h/ZGN4dZ4rq7N+qxCw=="
	tt, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}

	domainArr := strings.Split(string(tt), "\n")
	fmt.Println("len:", len(domainArr), len(str))
	for idx, dom := range domainArr {
		fmt.Println(idx, dom)
	}
}

func test6() {
	ptr, _ := net.LookupAddr("155.138.201.205")
	for _, ptrvalue := range ptr {
		fmt.Println(ptrvalue)
	}
}

func test5() {
	domains := []string{
		"192.168.0.1",
		"amazon.co.uk",
		"books.amazon.co.uk",
		"www.books.amazon.co.uk",
		"amazon.com",
		"",
		"example0.debian.net",
		"example1.debian.org",
		"",
		"golang.dev",
		"golang.net",
		"play.golang.org",
		"gophers.in.space.museum",
		"",
		"0emm.com",
		"a.0emm.com",
		"b.c.d.0emm.com",
		"",
		"there.is.no.such-tld",
		"",
		// Examples from the PublicSuffix function's documentation.
		"foo.org",
		"foo.co.uk",
		"foo.dyndns.org",
		"foo.blogspot.co.uk",
		"cromulent",
	}

	for _, domain := range domains {
		if domain == "" {
			fmt.Println(">")
			continue
		}

		eTLD, _ := publicsuffix.EffectiveTLDPlusOne(domain)
		fmt.Printf("> %24s%16s \n", domain, eTLD)
		//eTLD, icann := publicsuffix.PublicSuffix(domain)

		// Only ICANN managed domains can have a single label. Privately
		// managed domains must have multiple labels.
		//manager := "Unmanaged"
		//if icann {
		//	manager = "ICANN Managed"
		//} else if strings.IndexByte(eTLD, '.') >= 0 {
		//	manager = "Privately Managed"
		//}
		//
		//fmt.Printf("> %24s%16s  is  %s\n", domain, eTLD, manager)
	}
}

func test4() {
	resp, err := http.Get("https://raw.githubusercontent.com/proton-lab/quantum/master/gfw.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	buf, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Println("Update GFW list err:", e)
		return
	}

	domains, err := base64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		fmt.Println("Update GFW list err:", e)
		return
	}
	fmt.Println(string(domains))
}

func test3() {
	l, _ := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 51415,
	})
	fmt.Println(l.Addr().String())
}
func test1() {
	decodeBytes, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		panic(err)
	}

	ip4 := &layers.IPv4{}
	tcp := &layers.TCP{}
	udp := &layers.UDP{}
	dns := &layers.DNS{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, ip4, tcp, udp, dns)
	decodedLayers := make([]gopacket.LayerType, 0, 4)
	if err := parser.DecodeLayers(decodeBytes, &decodedLayers); err != nil {
		panic(err)
	}

	for _, typ := range decodedLayers {
		switch typ {
		case layers.LayerTypeDNS:

			for _, ask := range dns.Questions {
				fmt.Printf("	question:%s-%s-%s\n", ask.Name, ask.Class.String(), ask.Type.String())
			}

			for _, as := range dns.Answers {
				fmt.Println("	Answer:", as.String())
			}
			break
		case layers.LayerTypeIPv4:
			fmt.Println("	IPV4", ip4.SrcIP, ip4.DstIP)
			break
		case layers.LayerTypeTCP:
			fmt.Println("	TCP", tcp.SrcPort, tcp.DstPort)
			break
		case layers.LayerTypeUDP:
			fmt.Println("	UDP", udp.SrcPort, udp.DstPort)
			break
		}
	}
}
