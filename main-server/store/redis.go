// TODO: implement the redis engine
package store

const (
	ENGINE_REDIS = "redis"
)

type redisEngine struct{}

func (r *redisEngine) Init() (servers Servers, users Users, allServers map[string]int64)     { return }
func (r *redisEngine) LoadConfig(s string)                                                   {}
func (r *redisEngine) WriteUser(username string, u *User) (err error)                        { return }
func (r *redisEngine) AppendPingRet(server, location string, pr PingRet) (err error)         { return }
func (r *redisEngine) BatchWritePingRets(server, location string, prs []PingRet) (err error) { return }

func newRedisEngine() StoreEngine { return new(redisEngine) }

func init() {
	Register(ENGINE_MYSQL, newMysqlEngine)
}
