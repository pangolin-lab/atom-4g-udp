package wallet

type UserWallet interface {
	Running(done chan error)
	Finish()
	Query()(string,error)
	Recharge(no int) error
}