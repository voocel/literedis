import socket
from typing import Any, List, Optional
import struct

class Client:
    def __init__(self, host: str, port: int):
        self.conn = socket.create_connection((host, port))
        self.reader = self.conn.makefile('rb')

    def do(self, cmd: str, *args: Any) -> Any:
        self._write_command(cmd, *args)
        return self._read_reply()

    def _write_command(self, cmd: str, *args: Any) -> None:
        data = f"*{len(args) + 1}\r\n${len(cmd)}\r\n{cmd}\r\n".encode()
        for arg in args:
            arg_str = str(arg)
            data += f"${len(arg_str)}\r\n{arg_str}\r\n".encode()
        self.conn.sendall(data)

    def _read_reply(self) -> Any:
        byte = self.reader.read(1)
        if byte == b'+':
            return self._read_line().decode()
        elif byte == b'-':
            raise Exception(self._read_line().decode())
        elif byte == b':':
            return int(self._read_line())
        elif byte == b'$':
            length = int(self._read_line())
            if length == -1:
                return None
            data = self.reader.read(length)
            self.reader.read(2)  # discard CRLF
            return data.decode()
        elif byte == b'*':
            length = int(self._read_line())
            if length == -1:
                return None
            return [self._read_reply() for _ in range(length)]
        else:
            raise Exception(f"Unknown reply: {byte}")

    def _read_line(self) -> bytes:
        data = self.reader.readline()
        return data.rstrip(b'\r\n')

    def close(self) -> None:
        self.conn.close()
