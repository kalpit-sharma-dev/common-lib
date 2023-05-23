<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Redis :
    This is the wrapper library on top of "https://github.com/go-redis/redis".

### Third-Party Libraties
  - **Link** : https://github.com/go-redis/redis
  - **License** : [BSD 2-Clause "Simplified" License] (https://github.com/go-redis/redis/blob/master/LICENSE)
  - **Description** : Go library for Redis Client

### Use 

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/redis"
```

**Create Client**
```go
    client := redis.GetService(&redis.Config{
			ServerAddress: cfg.RedisClientConfig.RedisServerAddress,
			Password:      cfg.RedisClientConfig.Password,
			DB:            cfg.RedisClientConfig.DB,
		})
 
    client.Init()
```

***Note *** : If we did not call client.Init() after creating the client then the library will throw panic


**Configuration**

```go
type Config struct {
	// host:port address.
	ServerAddress string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int
	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration
}

```

**Use Client**
```go
	// Set Command 
	client.Set("Key", "Value")
	// Set With Expiry Command 
	client.SetWithExpiry("Key", "Value", Expiry Time)
	// Get values of keys
	client.Get("Key")
	//Delete Key
	client.Delete("key")
	//Delete multiple Keys
	client.Delete("key1","key2")
	// Increment key
	client.Incr("key")
	// Increment key by count
	client.IncrBy("key", count)
	// Decrement key
	client.Decr("key")
	// Decrement key by count
	client.DecrBy("key", count)
	// Receive all keys according to specified pattern
	client.Keys("pattern")
	// Expire key
	client.Expire("key", expiry time)
	// Get ttl of the key 
	client.TTL("key")
```
**Use Sorted Set**
```go
	/*Redis Sorted Sets are similar to Redis Sets with the unique feature of values stored in a set.
	The difference is, every member of a Sorted Set is associated with a score, 
	that is used in order to take the sorted set ordered, from the smallest to the greatest score.*/



	/*ZAdd :Adds all the specified members with the specified scores to the sorted set stored at key. 
	It is possible to specify multiple score / member pairs. If a specified member is already a member of the sorted set, 
	the score is updated and the element reinserted at the right position to ensure the correct ordering.

	If key does not exist, a new sorted set with the specified members as sole members is created, 
	like if the sorted set was empty. If the key exists but does not hold a sorted set, an error is returned.*/

	client.ZAdd("key", client.High, "member")

	/*ZRange :Returns the specified range of elements in the sorted set stored at key.
	 The elements are considered to be ordered from the lowest to the highest score. 
	 Lexicographical order is used for elements with equal score.*/

	client.ZRange("key",0,-1) 

	/*ZRem: Removes the specified members from the sorted set stored at key. 
	Non existing members are ignored.

	An error is returned when key exists and does not hold a sorted set.*/
	
	client.ZRem("key", "member")

	/*Exist: Returns if key exists*/
	
	client.Exists("key")
	
```

**Use redis Set**
```go
	/* If key does not exist, a new set with the specified members as sole members is created, like if the sorted set was empty. 
	The number of members added or removed is returned as int64 value. */

	//Add member/members to a set
	client.SAdd("setName",[]string{"member1","member2"})
	//Remove member/members from a set
	client.SRem("setName",[]string{"member1","member2"})
	//Check Existence of a member in a set
	client.SIsMember("setName","value")
	//Get List of members from a set
	client.SMembers("setName")
	// Perform union on sets with specified keys and store it in destination
	client.SUnionStore("DestinationSetName","Set1","Set2",...)
	
```

**Close CLient**
```go
    client.Close()
```

**Example**
    Please refer "devicestatenotifier" package in "platform-agent-service".
    Link : https://github.com/ContinuumLLC/platform-agent-service/tree/master/src/devicestatenotifier

**Wiki** : https://continuum.atlassian.net/wiki/spaces/EN/pages/1672348148/Device+Down+-+Redis+-+SDD

### Contribution

Any changes in this package should be communicated to Juno Agent and Common framework Team