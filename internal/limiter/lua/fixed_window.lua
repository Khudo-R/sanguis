local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])

local current = redis.call("INCR", key)

if current == 1 then
    redis.call("PEXPIRE", key, window_ms)
end

if current > limit then
    return { 0, current - 1 }
end

return { 1, limit - current }
