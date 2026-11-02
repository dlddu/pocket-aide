# mock-exception: EXT — FCM 발송은 Firebase 서비스 계정 키·실 기기 등록 토큰 경계라 CI 안에서 못 한다; 백엔드가 일회용 서비스 계정의 token_uri 와 FCM_HOST 로 보내는 OAuth 토큰 교환과 POST /v1/projects/<id>/messages:send 를 받아 기록하고 테스트 러너가 GET /e2e/received 로 조회한다 (docs/e2e-mocking-policy.md)
import base64
import json
import secrets
import socket
import sys
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

SCOPE = "https://www.googleapis.com/auth/firebase.messaging"
JWT_GRANT = "urn:ietf:params:oauth:grant-type:jwt-bearer"
MAX_DATA_BYTES = 4096
PROJECTS_PREFIX = "/v1/projects/"
SEND_SUFFIX = "/messages:send"

RECEIVED = []
ISSUED = set()
LOCK = threading.Lock()


def decode_segment(segment):
    padded = segment + "=" * (-len(segment) % 4)
    return json.loads(base64.urlsafe_b64decode(padded.encode("ascii")))


def google_error(code, status, message):
    return {"error": {"code": code, "message": message, "status": status}}


class Handler(BaseHTTPRequestHandler):
    def send_json(self, status, body):
        raw = json.dumps(body, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=UTF-8")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def read_body(self):
        length = int(self.headers.get("Content-Length") or 0)
        return self.rfile.read(length) if length else b""

    def token_url(self):
        return "http://%s/token" % self.headers.get("Host", "")

    def do_POST(self):
        path = urlparse(self.path).path
        if path == "/token":
            self.exchange_token()
        elif path.startswith(PROJECTS_PREFIX) and path.endswith(SEND_SUFFIX):
            self.send_message(path[len(PROJECTS_PREFIX):-len(SEND_SUFFIX)])
        else:
            self.send_json(404, google_error(404, "NOT_FOUND", "unknown path %s" % path))

    def exchange_token(self):
        form = parse_qs(self.read_body().decode("utf-8"))
        if (form.get("grant_type") or [""])[0] != JWT_GRANT:
            self.send_json(400, {"error": "unsupported_grant_type"})
            return
        parts = (form.get("assertion") or [""])[0].split(".")
        try:
            header, claims = decode_segment(parts[0]), decode_segment(parts[1])
        except (IndexError, ValueError):
            self.send_json(400, {"error": "invalid_grant", "error_description": "malformed assertion"})
            return
        problems = []
        if len(parts) != 3 or not parts[2]:
            problems.append("unsigned assertion")
        if header.get("alg") != "RS256":
            problems.append("alg must be RS256")
        if not claims.get("iss"):
            problems.append("missing iss")
        if SCOPE not in str(claims.get("scope", "")).split():
            problems.append("scope must include firebase.messaging")
        if claims.get("aud") != self.token_url():
            problems.append("aud must be %s" % self.token_url())
        if int(claims.get("exp", 0)) <= time.time():
            problems.append("expired assertion")
        if problems:
            self.send_json(400, {"error": "invalid_grant", "error_description": "; ".join(problems)})
            return
        token = "ya29.e2e-" + secrets.token_urlsafe(24)
        with LOCK:
            ISSUED.add(token)
        self.send_json(200, {"access_token": token, "token_type": "Bearer", "expires_in": 3599})

    def send_message(self, project):
        auth = self.headers.get("Authorization") or ""
        with LOCK:
            authorised = auth.startswith("Bearer ") and auth[len("Bearer "):] in ISSUED
        if not authorised:
            self.send_json(401, google_error(401, "UNAUTHENTICATED", "Request had invalid authentication credentials."))
            return
        try:
            message = json.loads(self.read_body() or b"{}").get("message")
        except ValueError:
            message = None
        if not isinstance(message, dict) or not message.get("token"):
            self.send_json(400, google_error(400, "INVALID_ARGUMENT", "message.token is required"))
            return
        data = message.get("data") or {}
        if not isinstance(data, dict) or any(not isinstance(v, str) for v in data.values()):
            self.send_json(400, google_error(400, "INVALID_ARGUMENT", "message.data values must be strings"))
            return
        notification = message.get("notification") or {}
        size = len(json.dumps(data, ensure_ascii=False).encode("utf-8")) + len(
            json.dumps(notification, ensure_ascii=False).encode("utf-8"))
        if size > MAX_DATA_BYTES:
            self.send_json(400, google_error(400, "INVALID_ARGUMENT", "Message is too big (%d bytes)" % size))
            return
        name = "projects/%s/messages/%s" % (project, secrets.token_hex(8))
        entry = {"project": project, "token": message["token"], "message": message, "name": name, "received_at": time.time()}
        with LOCK:
            RECEIVED.append(entry)
        sys.stderr.write("received %s\n" % json.dumps(entry, ensure_ascii=False))
        sys.stderr.flush()
        self.send_json(200, {"name": name})

    def do_GET(self):
        url = urlparse(self.path)
        if url.path == "/healthz":
            self.send_json(200, {"ok": True})
            return
        if url.path != "/e2e/received":
            self.send_json(404, google_error(404, "NOT_FOUND", "unknown path %s" % url.path))
            return
        token = (parse_qs(url.query).get("token") or [""])[0]
        with LOCK:
            rows = [r for r in RECEIVED if not token or r["token"] == token]
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
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 5559
    try:
        server = DualStackServer(("::", port), Handler)
    except OSError:
        server = ThreadingHTTPServer(("0.0.0.0", port), Handler)
    server.serve_forever()
