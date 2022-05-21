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
		elseif (bit.band(flag,1)) ~= 0 then
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
		local historyRemark = ARGV[6]

		local coins = redis.call("HGET", wallet, fromAccount)
		if coins == false or tonumber(coins) < fromCoins then
			if (bit.band(flag,4)) == 0 then
				return 1
			end
		end

		local toCoins = redis.call("HGET", toAccount, toIDKey)
		if not toCoins == false and (bit.band(flag,1)) == 0 then
			return 2
		end

		if toCoins == false then
			redis.call("HSET", toAccount, toIDKey, fromCoins)
		else
			redis.call("HINCRBY", toAccount, toIDKey, fromCoins)
		end

		redis.call("HINCRBY", toAccount, toTotalKey, fromCoins)

		redis.call("HINCRBY", wallet, fromAccount, -fromCoins)

		redis.call("LPUSH",  history, -fromCoins.."\n"..historyRemark)

		return 0
	`)

	walletTrans2WalletScript = redis.NewScript(`
		local walletFrom =  KEYS[1]
		local walletTo = KEYS[2]
		local historyFrom = KEYS[3]
		local historyTo = KEYS[4]

		local fromAccount = ARGV[1]
		local fromCoins = tonumber(ARGV[2])
		local toAccount = ARGV[3]
		local flag = tonumber(ARGV[4])
		local historyFromRemark = ARGV[5]
		local historyToRemark = ARGV[6]

		if tonumber(fromCoins) <= 0 then
			return redis.error_reply("invalid coins amount") 
		end

		local coins = redis.call("HGET", walletFrom, fromAccount)
		if coins == false or tonumber(coins) < fromCoins then
			if (bit.band(flag,4)) == 0 then
				return 1
			end
		end

		redis.call("HINCRBY", walletFrom, fromAccount, -fromCoins)
		redis.call("HINCRBY", walletTo, toAccount, fromCoins)

		redis.call("LPUSH",  historyFrom, -fromCoins.."\n"..historyFromRemark)
		redis.call("LPUSH",  historyTo, fromCoins.."\n"..historyToRemark)

		return 0
	`)

	lockerTrans2WalletScript = redis.NewScript(`
		local fromAccount =  KEYS[1]
		local wallet = KEYS[2]
		local history = KEYS[3]

		local fromIDKey = ARGV[1]
		local fromTotalKey = ARGV[2]
		local walletAccount = ARGV[3]
		local historyMember = ARGV[4]

		local fromCoins = redis.call("HGET", fromAccount, fromIDKey)
		if fromCoins == false then
			return 1
		end

		redis.call("HINCRBY", wallet, walletAccount, fromCoins)

		redis.call("HDEL", fromAccount, fromIDKey)
		redis.call("HINCRBY", fromAccount, fromTotalKey, -tonumber(fromCoins))

		redis.call("LPUSH", history, fromCoins.."\n"..historyMember)

		return 0
	`)
)
