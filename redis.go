package leiorm

// RedisArgs is a helper for constructing Redis command arguments from structured values.
type RedisArgs []interface{}

func (args RedisArgs) Add(value ...interface{}) RedisArgs {
	return append(args, value...)
}

type RedisCommand struct {
	Cmd  string
	Args RedisArgs
}

type RedisCommands []*RedisCommand

func (rcs RedisCommands) Add(cmds ...*RedisCommand) RedisCommands {
	return append(rcs, cmds...)
}

type RedisClienter interface {
	// Do sends a command to the server and returns the received reply.
	Do(cmd string, args ...interface{}) (interface{}, error)

	// Send writes the command to the client's output buffer.
	Send(cmd string, args ...interface{}) error

	// Flush flushes the output buffer to the Redis server.
	Flush() error

	// Receive receives a single reply from the Redis server
	Receive() (reply interface{}, err error)
}

func DoRedisCommands(rd RedisClienter, rcs RedisCommands) {
	for _, rc := range rcs {
		rd.Send(rc.Cmd, rc.Args...)
		//rd.Do(rc.Cmd, rc.Args...)
	}
	rd.Flush()
}
