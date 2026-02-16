local socket = require("socket")

local host = "127.0.0.1"
local port = 55556

print("Creating udp object")
local udp = assert(socket.udp())

print("Binding to localhost and port")
udp:setsockname(host, port)

while true do
    print("sending data")
    udp:sendto("hello from lua", "127.0.0.1", 55555)
    udp:sleep(1)
end
