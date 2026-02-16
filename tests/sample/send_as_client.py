#!/usr/bin/env python

"""
Module Docstring
"""

import socket
import time


def send_to(
    source_ip: str, source_port: int, destin_ip: str, destin_port: int, message: str
):
    """
    Send UDP Data to.

    Args:
        source_ip: The source ip address.

    Returns:
        None
    """

    sock = None
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.bind((source_ip, source_port))
        data = message.encode("utf-8")
        sock.sendto(data, (destin_ip, destin_port))

        # pylint: disable=broad-exception-caught
    except Exception as e:
        print(f"Error {e} occurred")

    finally:
        if not sock is None:
            sock.close()


if __name__ == "__main__":
    #    for x in range(10):
    send_to("localhost", 50001, "localhost", 55555, "Hello, from Python")
    time.sleep(1)
