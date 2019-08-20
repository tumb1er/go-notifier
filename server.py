import json
import socketserver
import time
from datetime import datetime
from random import random


class CurrentTimeHandler(socketserver.BaseRequestHandler):
    """
    Handler that sends current time with random delays.
    """

    def handle(self):
        tooltip = "Timer"
        title = "Current time"
        while True:
            time.sleep(random() * 3)
            info = datetime.now().isoformat()
            data = {
                'tooltip': tooltip,
                'title': title,
                'info': info
            }
            packet = f'{json.dumps(data)}\n'
            try:
                self.request.sendall(packet.encode('utf-8'))
            except (BrokenPipeError, ConnectionAbortedError):
                break


if __name__ == '__main__':
    server = socketserver.ThreadingTCPServer(("localhost", 9998), CurrentTimeHandler)
    server.serve_forever()
