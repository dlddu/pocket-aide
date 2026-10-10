# mock-exception: EXT — APNs 발송은 Apple 인증키·실 기기 토큰 경계라 CI 안에서 못 한다; 백엔드가 APNS_HOST 로 보내는 POST /3/device/<token> 을 받아 기록하고 테스트 러너가 GET /e2e/received 로 조회한다 (docs/e2e-mocking-policy.md)
import json
import socket
import sys
import threading
import time
import uuid
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

RECEIVED = []
RECEIVED_LOCK = threading.Lock()
DEVICE_PREFIX = "/3/device/"


class Handler(BaseHTTPRequestHandler):
    def send_json(self, status, body, headers=None):
        raw = json.dumps(body, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        for key, value in (headers or {}).items():
            self.send_header(key, value)
        self.end_headers()
        self.wfile.write(raw)

    def do_POST(self):
        path = urlparse(self.path).path
        if not path.startswith(DEVICE_PREFIX) or len(path) == len(DEVICE_PREFIX):
            self.send_json(404, {"reason": "BadPath"})
            return
        length = int(self.headers.get("Content-Length") or 0)
        try:
            payload = json.loads(self.rfile.read(length) or b"{}")
        except ValueError:
            self.send_json(400, {"reason": "PayloadEmpty"})
            return
        if not self.headers.get("apns-topic"):
            self.send_json(400, {"reason": "MissingTopic"})
            return
        if not (self.headers.get("authorization") or "").startswith("bearer "):
            self.send_json(403, {"reason": "MissingProviderToken"})
            return
        apns_id = str(uuid.uuid4()).upper()
        entry = {
            "device": path[len(DEVICE_PREFIX):],
            "topic": self.headers.get("apns-topic"),
            "payload": payload,
            "apns_id": apns_id,
            "received_at": time.time(),
        }
        with RECEIVED_LOCK:
            RECEIVED.append(entry)
        sys.stderr.write("received %s\n" % json.dumps(entry, ensure_ascii=False))
        sys.stderr.flush()
        self.send_response(200)
        self.send_header("apns-id", apns_id)
        self.send_header("Content-Length", "0")
        self.end_headers()

    def do_GET(self):
        url = urlparse(self.path)
        if url.path == "/healthz":
            self.send_json(200, {"ok": True})
            return
        if url.path != "/e2e/received":
            self.send_json(404, {"reason": "BadPath"})
            return
        device = (parse_qs(url.query).get("device") or [""])[0]
        with RECEIVED_LOCK:
            rows = [r for r in RECEIVED if not device or r["device"] == device]
        self.send_json(200, rows)

    def log_message(self, fmt, *args):
        sys.stderr.write("%s %s\n" % (self.log_date_time_string(), fmt % args))
        sys.stderr.flush()


class DualStackServer(ThreadingHTTPServer):
    address_family = socket.AF_INET6

    def server_bind(self):
        self.socket.setsockopt(socket.IPPROTO_IPV6, socket.IPV6_V6ONLY, 0)
        super().server_bind()


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 5558
    DualStackServer(("::", port), Handler).serve_forever()
