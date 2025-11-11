import socket


def main():
    st = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    try:
        st.connect(("0.0.0.0", 45000))

        while True:
            data = st.recv(1600)
            if not data:
                print("No data received.  Connection closed")
                break

            print(f"Data received {len(data)}")
            pass
    except Exception as e:
        print(f"error {e}")


if __name__ == "__main__":
    main()
