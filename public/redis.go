package public

import (
	"gateway/utils"
	"github.com/garyburd/redigo/redis"
)

func RedisConfPipeline(pip ...func(c redis.Conn)) error {
	c, err := utils.RedisConnFactory("default")
	if err != nil {
		return err
	}
	defer c.Close()
	for _, f := range pip {
		f(c)
	}
	c.Flush()
	return nil
}

func RedisConfDo(commandName string, args ...interface{}) (interface{}, error) {
	c, err := utils.RedisConnFactory("default")
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.Do(commandName, args...)
}
