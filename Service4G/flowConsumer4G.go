package Service4G

import (
	"fmt"
	"github.com/Iuduxras/atom-4g/wallet"
)

type Consumer4G struct {
	Done   chan error
	Wallet wallet.UserWallet
}

func NewConsumer(addr string, w wallet.UserWallet) (*Consumer4G, error) {
	ap := &Consumer4G{
		Wallet: w,
		Done:   make(chan error),
	}
	return ap, nil
}

func (c4g *Consumer4G) Consuming() {

	//loop:
	for {
		select {
		case err := <-c4g.Done:
			fmt.Printf("Consumer4G exit for:%s", err.Error())
			//break loop
			c4g.Finish()
			return
		}
	}

	fmt.Println("consumer closed bypass loop")
	defer c4g.Finish()
}

func (c4g *Consumer4G) Finish() {

	if c4g.Wallet != nil {
		c4g.Wallet.Finish()
	}
}

func (c4g *Consumer4G) Query() string {
	if r, e := c4g.Wallet.Query(); e != nil {
		c4g.Done <- e
		return ""
	} else {
		return r
	}
}

func (c4g *Consumer4G) Recharge(no int) error {
	if err := c4g.Wallet.Recharge(no); err != nil {
		c4g.Done <- err
		return err
	} else {
		return nil
	}
}
