package redisimpls

import "github.com/go-redis/redis/v8"

var (
	addAccountScript = redis.NewScript(`
		local idKey =  KEYS[1]
		local nameKey = KEYS[2]
		local usersCreateAtKey = KEYS[3]

		local vId = ARGV[1]
		local vName = ARGV[2]
		local vHPass = ARGV[3]
		local vData = ARGV[4]
		local vCreateAt = ARGV[5]

		local ret = redis.call("GET", nameKey)
		if ret ~= false then
			return redis.error_reply("name exists") 
		end

		local exists = redis.call('EXISTS', idKey)
		
		if exists == 1 then
			return redis.error_reply("id exists") 
		end

		redis.call("HSET", idKey, "name", vName, "pass", vHPass, "data", vData, "create_at", vCreateAt)
		redis.call("SET", nameKey, vId)
		redis.call("ZADD", usersCreateAtKey, vCreateAt, vId)

		return 0
	`)

	updateAccountPasswordScript = redis.NewScript(`
		local idKey =  KEYS[1]

		local vHPass = ARGV[1]

		local exists = redis.call('EXISTS', idKey)
		
		if exists == 0 then
			return redis.error_reply("id not exists") 
		end

		redis.call("HSET", idKey, "pass", vHPass)

		return 0
	`)

	updateAccountNameScript = redis.NewScript(`
		local idKey =  KEYS[1]

		local vName = ARGV[1]

		local exists = redis.call('EXISTS', idKey)
		
		if exists == 0 then
			return redis.error_reply("id not exists") 
		end

		redis.call("HSET", idKey, "name", vName)

		return 0
	`)

	updateAccountAdvanceConfigScript = redis.NewScript(`
		local idKey =  KEYS[1]

		local vAdvCfg = ARGV[1]

		local exists = redis.call('EXISTS', idKey)
		
		if exists == 0 then
			return redis.error_reply("id not exists") 
		end

		redis.call("HSET", idKey, "adv_cfg", vAdvCfg)

		return 0
	`)

	updateAccountPropertyDataScript = redis.NewScript(`
		local idKey =  KEYS[1]

		local vData = ARGV[1]

		local exists = redis.call('EXISTS', idKey)
		
		if exists == 0 then
			return redis.error_reply("id not exists") 
		end

		redis.call("HSET", idKey, "property_data", vData)

		return 0
	`)

	tokenAddScript = redis.NewScript(`
		local tokenKey =  KEYS[1]
		local idTokensKey = KEYS[2]

		local vToken = ARGV[1]
		local vUID = ARGV[2]
		local vTTLSeconds = ARGV[3]

		redis.call("SET", tokenKey, vUID, "EX", vTTLSeconds)
		redis.call("SADD", idTokensKey, vToken)

		return 0
	`)

	tokenCheckScript = redis.NewScript(`
		local tokenKey =  KEYS[1]

		local vRenewSeconds = tonumber(ARGV[1])

		local ttl = redis.call("TTL", tokenKey)
		if ttl == -2 then
			return 0
		end

		if ttl == -1 then
			return 1
		end

		if vRenewSeconds <= 0 then
			return 1
		end
		
		redis.call("EXPIRE", tokenKey, ttl+vRenewSeconds)

		return 1
	`)

	tokenDelScript = redis.NewScript(`
		local tokenKey =  KEYS[1]
		local idTokensKey = KEYS[2]

		local vToken = ARGV[1]

		redis.call("DEL", tokenKey)
		redis.call("SREM", idTokensKey, vToken)

		return 0
	`)
)
