local socket = require 'socket'

local conn = socket.connect('localhost', 8000)

---@param c table
---@param str string
local function write_string_with_len(c, str)
	return c:send(string.char(#str) .. str)
end

-- identify
write_string_with_len(conn, 'gregory')

-- subscribe
conn:send '\x00'
write_string_with_len(conn, 'games')

-- listen
while true do
	local data_len = conn:receive(1)
	if not data_len then
		break
	end

	data_len = string.byte(data_len)

	local data = conn:receive(data_len)
	print(data)
end
