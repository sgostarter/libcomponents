package account

type Account interface {
	Register(accountName, password string) (uid uint64, err error)
	Login(accountName, password string) (uid uint64, token string, err error)
	Who(token string) (uid uint64, accountName string, err error)
	Logout(token string) error

	SetPropertyData(token string, d interface{}) error
	GetPropertyData(token string, d interface{}) error
}

type Storage interface {
	AddAccount(accountName, hashedPassword string) (uid uint64, err error)
	FindAccount(accountName string) (uid uint64, hashedPassword string, err error)

	AddToken(token string) error
	DelToken(token string) error
	TokenExists(token string) (bool, error)

	SetPropertyData(accountName string, d interface{}) error
	GetPropertyData(accountName string, d interface{}) error
}
