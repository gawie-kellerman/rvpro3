local socket = require("socket")

local host = "127.0.0.1"
local port = 55555

local udp = assert(socket.udp())

-- Equivalent of bind
udp:setsockname(host, port)

while true do
    local data, sender_ip, sender_port = udp:receivefrom()

    if data then
        print(string.format("Received %s:%d, %s", sender_ip, sender_port, data))
    end
end
