// TODO: implement the mysql engine
package store

const (
	ENGINE_MYSQL = "mysql"
)

type mysqlEngine struct{}

func (m *mysqlEngine) Init() (servers Servers, users Users, allServers map[string]int64)     { return }
func (m *mysqlEngine) LoadConfig(s string)                                                   {}
func (m *mysqlEngine) WriteUser(username string, u *User) (err error)                        { return }
func (m *mysqlEngine) BatchWritePingRets(server, location string, prs []PingRet) (err error) { return }

// func (m *mysqlEngine) AppendPingRet(server, location string, pr PingRet) (err error)         { return }

func newMysqlEngine() StoreEngine { return new(mysqlEngine) }

func init() {
	Register(ENGINE_MYSQL, newMysqlEngine)
}
