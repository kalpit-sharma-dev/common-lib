package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) *redis.Client {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client
}

func Test_redisFunctionality(t *testing.T) {
	z := Z{Score: 0, Member: "abc"}
	z1 := Z{Score: 0, Member: "xyz"}

	t.Run("Add_member_to_a_sorted_set,_or_update_its_score_if_it_already_exists", func(t *testing.T) {

		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		c.ZAdd("k1", z, z1)
		strarray, err := c.ZRange("k1", 0, -1)
		if err != nil {
			t.Errorf("clientImpl.ZAdd() error = %v, wantErr %s", err, "nil")
			return
		}
		if strarray[0] != z.Member || strarray[1] != z1.Member {
			t.Errorf("clientImpl.ZAdd() = %v and %v, want %v and %v", z.Member, z1.Member, strarray[0], strarray[1])
		}
		_, err = c.ZRem("k1", z1.Member)

		if err != nil {
			t.Errorf("error of ZRem() while removing member; error = %v, wantErr %s", err, "nil")
			return
		}
		strarray, err = c.ZRange("k1", 0, -1)
		if strarray[0] != z.Member || len(strarray) != 1 {
			t.Errorf("after removal of element length of array is not equal to 1 or got value is %v but want value %v", strarray[0], z.Member)
		}
		output, err := c.Exists("k1")
		if err != nil {
			t.Errorf("error of ZRem() while removing member; error = %v, wantErr %s", err, "nil")
			return
		}
		if output != 1 {
			t.Errorf("No of element is set with k1 key is equal to: %v but no of elements want: 1", output)
		}
	})
}

func Test_redisSetFunctionality(t *testing.T) {
	members := []string{"member1"}
	members2 := []string{"member2", "member3"}

	t.Run("Add_member_to_a_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//adding member to set
		result, err := c.SAdd("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 1 {
			t.Errorf("clientImpl.SAdd() = %v , want = %v", result, 1)
		}
	})

	t.Run("Get_members_list_in_a_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//adding member to set
		_, err := c.SAdd("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		//fetching list of member
		result, err := c.SMembers("TestKey")
		if err != nil {
			t.Errorf("clientImpl.SMembers() error = %v, wantErr = %s", err, "nil")
			return
		}
		assert.Equal(t, members, result)
	})

	t.Run("Add_multiple_members_to_a__set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		result, err := c.SAdd("TestKey", members2)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 2 {
			t.Errorf("clientImpl.SAdd() = %v , want = %v", result, 2)
		}
	})

	t.Run("Add_already_existing_member_to_a_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//adding member
		_, err := c.SAdd("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		//adding member to set again
		result, err := c.SAdd("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 0 {
			t.Errorf("clientImpl.SAdd() = %v , want = %v", result, 0)
		}
	})

	t.Run("Passing_empty_slice_to_SAdd", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		_, err := c.SAdd("TestKey", []string{})
		if err == nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "ERR wrong number of arguments for 'sadd' command")
			return
		}
	})

	t.Run("Remove_member_from_a_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//adding member to remove later
		_, err := c.SAdd("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		//removing member from set
		result, err := c.SRem("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SRem() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 1 {
			t.Errorf("clientImpl.SRem() = %v , want = %v", result, 1)
		}
	})

	t.Run("Remove_multiple_members_from_a_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//adding members to remove later
		_, err := c.SAdd("TestKey", members2)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		//removing members from set
		result, err := c.SRem("TestKey", members2)
		if err != nil {
			t.Errorf("clientImpl.SRem() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 2 {
			t.Errorf("clientImpl.SRem() = %v , want = %v", result, 2)
		}
	})

	t.Run("Remove_members_which_are_not_in_a_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//removing members from set
		result, err := c.SRem("TestKey", members2)
		if err != nil {
			t.Errorf("clientImpl.SRem() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 0 {
			t.Errorf("clientImpl.SRem() = %v , want = %v", result, 0)
		}
	})

	t.Run("Passing_empty_slice_to_SRem", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		_, err := c.SRem("TestKey", []string{})
		if err == nil {
			t.Errorf("clientImpl.SRem() error = %v, wantErr = %s", err, "ERR wrong number of arguments for 'sadd' command")
			return
		}
	})

	t.Run("Member_present_in_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//adding members to check existence later
		_, err := c.SAdd("TestKey", members)
		if err != nil {
			t.Errorf("clientImpl.SAdd() error = %v, wantErr = %s", err, "nil")
			return
		}
		//checking member exixtence in set
		result, err := c.SIsMember("TestKey", members[0])
		if err != nil {
			t.Errorf("clientImpl.SIsMember() error = %v, wantErr = %s", err, "nil")
			return
		}
		if !result {
			t.Errorf("clientImpl.SIsMember() = %v , want = %v", result, true)
		}
	})

	t.Run("Member_not_present_in_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//checking member exixtence in set
		result, err := c.SIsMember("TestKey", members[0])
		if err != nil {
			t.Errorf("clientImpl.SIsMember() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result {
			t.Errorf("clientImpl.SIsMember() = %v , want = %v", result, false)
		}
	})

	t.Run("union_of_set", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//create sets to perform union
		_, err := c.SAdd("TestKey1", "member1", "member2")
		require.NoError(t, err)
		_, err = c.SAdd("TestKey2", "member3", "member4")
		require.NoError(t, err)
		//perform union operation on sets
		result, err := c.SUnionStore("TestKey3", "TestKey1", "TestKey2")
		if err != nil {
			t.Errorf("clientImpl.SUnionStore() error = %v, wantErr = %s", err, "nil")
			return
		}
		if result != 4 {
			t.Errorf("clientImpl.SUnionStore() = %v , want = %v", result, 4)
		}
	})
}

//test set commands in pipeline mode
func Test_redisPipelinersSetFunctionality(t *testing.T) {
	members := []string{"member1", "member2"}
	members2 := []string{"member3", "member4"}

	t.Run("Add/remove_member_to/from_a_set_using_pipeline/Incr/Expire", func(t *testing.T) {
		c := &clientImpl{
			config: &Config{},
			client: newTestRedis(t),
		}
		//creating a pipeline
		pipeline := c.CreatePipeline()
		//adding PSAdd into pipeline
		pSAddResult1 := pipeline.PSAdd("TestPipeSet", members)
		pSAddResult2 := pipeline.PSAdd("TestPipeSet", members2)
		//adding PSRem into pipeline
		pSRemResult1 := pipeline.PSRem("TestPipeSet", members)
		pSRemResult2 := pipeline.PSRem("TestPipeSet", members2)
		//adding Incr into pipeline
		pIncrResult := pipeline.Incr("count")
		pExpResult := pipeline.Expire("count", time.Second)
		result, err := pipeline.Exec()
		//Pipeline exec error nil check
		require.NoError(t, err)
		//Assert PSAdd Errors
		assert.NoError(t, result[0].Err)
		assert.NoError(t, result[1].Err)
		//Assert PSRem Errors
		assert.NoError(t, result[2].Err)
		assert.NoError(t, result[3].Err)
		//Assert pIncr Errors
		assert.NoError(t, result[4].Err)
		//Assert Expire Errors
		assert.NoError(t, result[5].Err)
		//Assert PSAdd Args
		assert.EqualValues(t, []interface{}{"sadd", "TestPipeSet", "member1", "member2"}, result[0].Args)
		assert.EqualValues(t, []interface{}{"sadd", "TestPipeSet", "member3", "member4"}, result[1].Args)
		//Assert PSRem Args
		assert.EqualValues(t, []interface{}{"srem", "TestPipeSet", "member1", "member2"}, result[2].Args)
		assert.EqualValues(t, []interface{}{"srem", "TestPipeSet", "member3", "member4"}, result[3].Args)
		//Assert pIncr Args
		assert.EqualValues(t, []interface{}{"incr", "count"}, result[4].Args)
		//Assert Expire Args
		assert.EqualValues(t, []interface{}{"expire", "count", int64(1)}, result[5].Args)
		//Assert PSAdd Return Value
		assert.EqualValues(t, 2, pSAddResult1.Val())
		assert.EqualValues(t, 2, pSAddResult2.Val())
		//Assert PSRem Return Value
		assert.EqualValues(t, 2, pSRemResult1.Val())
		assert.EqualValues(t, 2, pSRemResult2.Val())
		//Assert pIncr Return Value
		assert.EqualValues(t, 1, pIncrResult.Val())
		//Assert Expire Return Value
		assert.True(t, pExpResult.Val())
	})
}

func TestIncr(t *testing.T) {
	redis := newTestRedis(t)
	client := clientImpl{
		client: redis,
		config: &Config{},
	}

	client.Incr("counter")

	counter, _ := redis.Get("counter").Result()
	if counter != "1" {
		t.Errorf("expected: 1, got: %v", counter)
	}
}

func TestIncrBy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		redis := newTestRedis(t)
		client := clientImpl{
			client: redis,
			config: &Config{},
		}

		got, err := client.IncrBy("counter", 2)
		assert.NoError(t, err)
		assert.EqualValues(t, 2, got)

		counter, err := redis.Get("counter").Result()
		assert.NoError(t, err)
		assert.EqualValues(t, "2", counter)
	})
	t.Run("Failed", func(t *testing.T) {
		rBrokenClient := redis.NewClient(&redis.Options{Addr: "1.2.3.4"})

		client := clientImpl{
			client: rBrokenClient,
			config: &Config{},
		}

		got, err := client.IncrBy("counter", 2)
		assert.Error(t, err)
		assert.EqualValues(t, 0, got)
	})
}

func TestDecr(t *testing.T) {
	redis := newTestRedis(t)
	client := clientImpl{
		client: redis,
		config: &Config{},
	}

	_, err := client.Decr("counter")
	require.NoError(t, err)

	counter, err := redis.Get("counter").Result()
	require.NoError(t, err)
	if counter != "-1" {
		t.Errorf("expected: -1, got: %v", counter)
	}
}

func TestDecrBy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		redis := newTestRedis(t)
		client := clientImpl{
			client: redis,
			config: &Config{},
		}

		got, err := client.DecrBy("counter", 2)
		assert.NoError(t, err)
		assert.EqualValues(t, -2, got)

		counter, err := redis.Get("counter").Result()
		assert.NoError(t, err)
		assert.EqualValues(t, "-2", counter)
	})
	t.Run("Failed", func(t *testing.T) {
		rBrokenClient := redis.NewClient(&redis.Options{Addr: "1.2.3.4"})

		client := clientImpl{
			client: rBrokenClient,
			config: &Config{},
		}

		got, err := client.DecrBy("counter", 2)
		assert.Error(t, err)
		assert.EqualValues(t, 0, got)
	})
}

func TestKeys(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		redis := newTestRedis(t)
		client := clientImpl{
			client: redis,
			config: &Config{},
		}

		err := client.Set("counter-1", 1)
		assert.NoError(t, err)

		err = client.Set("counter-2", 2)
		assert.NoError(t, err)

		counters, err := redis.Keys("counter*").Result()
		assert.NoError(t, err)
		assert.EqualValues(t, []string{"counter-1", "counter-2"}, counters)
	})
	t.Run("Failed", func(t *testing.T) {
		rBrokenClient := redis.NewClient(&redis.Options{Addr: "1.2.3.4"})

		client := clientImpl{
			client: rBrokenClient,
			config: &Config{},
		}

		got, err := client.Keys("counter*")
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestExpire(t *testing.T) {
	redis := newTestRedis(t)
	client := clientImpl{
		client: redis,
		config: &Config{},
	}

	redis.Set("cache", "data", 0)

	client.Expire("cache", 10*time.Second)

	ttl, _ := redis.TTL("cache").Result()
	if ttl.Seconds() == -1 {
		t.Errorf("key doesn't have associated expire")
	}
	if ttl.Seconds() == -2 {
		t.Errorf("key not found")
	}
}

func Test_clientImpl_MGet(t *testing.T) {
	type args struct {
		keys []string
	}
	type storedData struct {
		key   string
		value string
	}

	tests := []struct {
		name       string
		args       args
		storedData []storedData
		want       interface{}
		wantErr    bool
	}{
		{
			name: "TestCase1:Successfully_got_all_records",
			args: args{keys: []string{"1", "2", "3"}},
			storedData: []storedData{
				{
					key:   "1",
					value: "value1",
				},
				{
					key:   "2",
					value: "value2",
				},
				{
					key:   "3",
					value: "value3",
				},
			},
			want:    []interface{}{"value1", "value2", "value3"},
			wantErr: false,
		},
		{
			name: "TestCase2:Successfully_got_only_2_records",
			args: args{keys: []string{"1", "2", "3"}},
			storedData: []storedData{
				{
					key:   "1",
					value: "value1",
				},
				{
					key:   "3",
					value: "value3",
				},
			},
			want:    []interface{}{"value1", nil, "value3"},
			wantErr: false,
		},
		{
			name: "TestCase3:Successfully_got_only_1_records",
			args: args{keys: []string{"1", "2", "3"}},
			storedData: []storedData{
				{
					key:   "3",
					value: "value3",
				},
			},
			want:    []interface{}{nil, nil, "value3"},
			wantErr: false,
		},
		{
			name:       "TestCase4:not_found_any_records",
			args:       args{keys: []string{"1", "2", "3"}},
			storedData: []storedData{},
			want:       []interface{}{nil, nil, nil},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &clientImpl{
				client: newTestRedis(t),
				config: &Config{},
			}

			for _, item := range tt.storedData {
				assert.NoError(t, c.Set(item.key, item.value))
			}

			got, err := c.MGet(tt.args.keys...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTTL(t *testing.T) {
	rd := newTestRedis(t)
	cl := clientImpl{
		client: rd,
		config: &Config{},
	}

	rd.Set("cache", "data", 10*time.Second)

	d, err := cl.TTL("cache")
	assert.NoError(t, err)

	if d.Seconds() < 0 || d.Seconds() > 10 {
		t.Error("invalid ttl")
	}
}
