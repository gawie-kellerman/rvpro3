import socket
import struct

PT_UNKNOWN = 0
PT_UDP_FORWARD = 1
PT_RADAR_MULTICAST = 2
PT_UDP_INSTRUCTION = 3
PT_STATS = 4
PT_SERVER_CLOSES_CONNECTION = 5 

HEADER_SIZE = 23

class Packet:
    def __init__(self):
        self.delimiter = 1044266619
        self.version = 1
        self.date = 0
        self.size = 23
        self.type = PT_UNKNOWN
        self.target_ip = 0
        self.target_port = 0
        self.data = None
        self.source_addr = None

    def read(self, addr, buffer):
        print(type(buffer))
        print(len(buffer))
        print(buffer.hex())
        self.delimiter = struct.unpack_from('<l', buffer, 0)[0]
        self.version = struct.unpack_from('<H', buffer, 4)[0]
        self.date = struct.unpack_from('<q', buffer, 6)[0]
        self.size = struct.unpack_from('<H', buffer, 14)[0]
        self.type = int(buffer[16])
        self.target_ip = struct.unpack_from('<I', buffer, 17)[0]
        self.target_port = struct.unpack_from('<H', buffer, 21)[0]
        copy_size = self.size - HEADER_SIZE
        self.data = buffer[HEADER_SIZE:copy_size]
        self.source_addr = addr

    def dump(self):
        print(f"version {self.version}, type: {self.type}, size: {self.size}, port: {self.target_port}")


def print_summary(summary):
    msg_total = 0
    for address in summary:
        detail = summary[address]
        print(address, " messages: ", detail[1], " bytes:", detail[2])
        msg_total += detail[1]
    print("total messages:", msg_total)


def main():
    summary = {}
    empty = 0
    count = 0

    sck = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sck.bind(("192.168.11.1", 55555))

    while True:
        message, address = sck.recvfrom(8192)

        print("received ", len(message), " from", address)

        # We echo the message as an instruction back to the virtual radar
        # sending it via the VirtualHost to the HubHost to the dispatcher
        sck.sendto(message, ("192.168.11.12", 55555))


        if len(message) == 0:
            empty += 1

            if empty > 10:
                break

        else:
            count += 1

            detail = summary.get(address, [address, 0, 0])
            detail[1] += 1
            detail[2] += len(message)
            summary[address] = detail


            if count % 100 == 0:
                print_summary(summary)

if __name__ == "__main__":
    print("Starting up TCPHub aggregate as desktop")
    main()
