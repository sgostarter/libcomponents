package wallet

import "github.com/go-redis/redis/v8"

var (
	lockSetScript = redis.NewScript(`
		local account =  KEYS[1]

		local idKey = ARGV[1]
		local totalKey = ARGV[2]
		local val = tonumber(ARGV[3])
		local flag = tonumber(ARGV[4])

		local ret = redis.call("HGET", account, idKey)
		if not ret == false and flag <= 0 then
			return redis.error_reply("exists") 
		end

		local incr = val
		if ret == false then
			redis.call("HSET", account, idKey, val)
		elseif flag == 1 then
			redis.call("HINCRBY", account, idKey, val)
		else
			redis.call("HSET", account, idKey, val)
			incr = incr - tonumber(ret)
		end

		redis.call("HINCRBY", account, totalKey, incr)

		return 1
	`)

	lockRemoveScript = redis.NewScript(`
		local account =  KEYS[1]

		local idKey = ARGV[1]
		local totalKey = ARGV[2]

		local coins = redis.call("HGET", account, idKey)
		if coins == false then
			return 0
		end

		redis.call("HINCRBY", account, totalKey, -tonumber(coins))

		return redis.call("HDEL", account, idKey)
	`)

	lockTransferScript = redis.NewScript(`
		local fromAccount =  KEYS[1]
		local toAccount = KEYS[2]

		local fromIDKey = ARGV[1]
		local fromTotalKey = ARGV[2]
		local toIDKey = ARGV[3]
		local toTotalKey = ARGV[4]
		local flag = tonumber(ARGV[5])

		local fromCoins = redis.call("HGET", fromAccount, fromIDKey)
		if fromCoins == false then
			return redis.error_reply("notExists") 
		end

		local toCoins = redis.call("HGET", toAccount, toIDKey)
		if not toCoins == false and flag ~= 1 then
			return redis.error_reply("exists") 
		end

		if toCoins == false then
			redis.call("HSET", toAccount, toIDKey, fromCoins)
		else
			redis.call("HINCRBY", toAccount, toIDKey, fromCoins)
		end

		redis.call("HINCRBY", toAccount, toTotalKey, fromCoins)

		redis.call("HDEL", fromAccount, fromIDKey)
		redis.call("HINCRBY", fromAccount, fromTotalKey, -tonumber(fromCoins))

		return true
	`)

	walletTrans2LockerScript = redis.NewScript(`
		local wallet =  KEYS[1]
		local toAccount = KEYS[2]
		local history = KEYS[3]

		local fromAccount = ARGV[1]
		local fromCoins = tonumber(ARGV[2])
		local toIDKey = ARGV[3]
		local toTotalKey = ARGV[4]
		local flag = tonumber(ARGV[5])
		local historyScore = ARGV[6]
		local historyRemark = ARGV[7]

		local coins = redis.call("HGET", wallet, fromAccount)
		if coins == false or tonumber(coins) < fromCoins then
			return 1
		end

		local toCoins = redis.call("HGET", toAccount, toIDKey)
		if not toCoins == false and flag ~= 1 then
			return 2
		end

		if toCoins == false then
			redis.call("HSET", toAccount, toIDKey, fromCoins)
		else
			redis.call("HINCRBY", toAccount, toIDKey, fromCoins)
		end

		redis.call("HINCRBY", toAccount, toTotalKey, fromCoins)

		redis.call("HINCRBY", wallet, fromAccount, -fromCoins)

		redis.call("ZADD",  history, historyScore,  -fromCoins.."\n"..historyRemark)

		return 0
	`)

	lockerTrans2WalletScript = redis.NewScript(`
		local fromAccount =  KEYS[1]
		local wallet = KEYS[2]
		local history = KEYS[3]

		local fromIDKey = ARGV[1]
		local fromTotalKey = ARGV[2]
		local walletAccount = ARGV[3]
		local historyScore = ARGV[4]
		local historyMember = ARGV[5]

		local fromCoins = redis.call("HGET", fromAccount, fromIDKey)
		if fromCoins == false then
			return 1
		end

		redis.call("HINCRBY", wallet, walletAccount, fromCoins)

		redis.call("HDEL", fromAccount, fromIDKey)
		redis.call("HINCRBY", fromAccount, fromTotalKey, -tonumber(fromCoins))

		redis.call("ZADD", history, historyScore, fromCoins.."\n"..historyMember)

		return 0
	`)
)
