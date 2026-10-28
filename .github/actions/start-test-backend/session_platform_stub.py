# mock-exception: EXT — 실 세션 플랫폼 컨트롤 플레인은 k8s 클러스터·데이터 플레인 이미지·공급자 자격증명을 요구해 macOS 러너에서 띄울 수 없다; 백엔드(sessions.Client)가 부르는 앱 경로만 실 응답 모양으로 서빙한다 (docs/e2e-mocking-policy.md)
import json
import sys
import threading
import uuid
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

API = "/api/v1"

CONFIG = {
    "claudeCode": {"defaultModel": "sonnet", "models": ["sonnet", "opus"]},
    "workloadTypes": [
        {"type": "shell", "label": "셸", "description": "터미널에서 직접 명령을 실행해요.",
         "takesModel": False, "input": "terminal", "defaultModel": None, "models": []},
        {"type": "claude-code", "label": "Claude Code", "description": "Claude 에이전트에게 프롬프트로 일을 맡겨요.",
         "takesModel": True, "input": "prompt", "defaultModel": "sonnet", "models": ["sonnet", "opus"]},
        {"type": "minimax-code", "label": "MiniMax Code", "description": "MiniMax 에이전트에게 프롬프트로 일을 맡겨요.",
         "takesModel": True, "input": "prompt", "defaultModel": "minimax-m2", "models": ["minimax-m2"]},
    ],
}

LOCK = threading.Lock()
SESSIONS = {}
OUTPUTS = {}
UNREACHABLE = False


def now():
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%fZ")


def public(s):
    return dict(s)


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, fmt, *args):
        sys.stderr.write("%s %s\n" % (self.address_string(), fmt % args))

    def send_json(self, status, body=None):
        raw = b"" if body is None else json.dumps(body).encode()
        self.send_response(status)
        if body is not None:
            self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def read_body(self):
        n = int(self.headers.get("Content-Length") or 0)
        raw = self.rfile.read(n) if n else b""
        if not raw.strip():
            return {}
        return json.loads(raw)

    def drop(self):
        self.close_connection = True
        self.connection.close()

    def route(self, method):
        global UNREACHABLE
        path = self.path.split("?", 1)[0]
        if path == "/healthz":
            return self.send_json(200, {"ok": True})
        if path.startswith("/e2e/") and method == "POST":
            with LOCK:
                if path == "/e2e/reset":
                    SESSIONS.clear()
                    OUTPUTS.clear()
                    UNREACHABLE = False
                elif path == "/e2e/unreachable":
                    UNREACHABLE = True
                elif path == "/e2e/reachable":
                    UNREACHABLE = False
                else:
                    return self.send_json(404, {"error": "not found"})
            return self.send_json(200, {"ok": True})
        if UNREACHABLE:
            return self.drop()
        if not path.startswith(API + "/"):
            return self.send_json(404, {"error": "not found"})
        parts = path[len(API) + 1:].split("/")
        try:
            body = self.read_body() if method == "POST" else {}
        except ValueError:
            return self.send_json(400, {"error": "invalid json"})
        with LOCK:
            return self.dispatch(method, parts, body)

    def dispatch(self, method, parts, body):
        if parts == ["config"] and method == "GET":
            return self.send_json(200, CONFIG)
        if parts == ["sessions"] and method == "GET":
            return self.send_json(200, {"sessions": [public(s) for s in SESSIONS.values()]})
        if parts == ["sessions"] and method == "POST":
            name = str(body.get("name", "")).strip()
            wl = body.get("workloadType", "shell")
            info = next((w for w in CONFIG["workloadTypes"] if w["type"] == wl), None)
            if not name or info is None:
                return self.send_json(400, {"error": "invalid request"})
            sid = uuid.uuid4().hex[:12]
            stamp = now()
            s = {"id": sid, "name": name, "workloadType": wl, "state": "active",
                 "createdAt": stamp, "lastAccess": stamp}
            if info["takesModel"]:
                s["model"] = body.get("model") or info["defaultModel"]
            SESSIONS[sid] = s
            OUTPUTS[sid] = ""
            return self.send_json(201, public(s))
        if len(parts) < 2 or parts[0] != "sessions":
            return self.send_json(404, {"error": "not found"})
        s = SESSIONS.get(parts[1])
        if s is None:
            return self.send_json(404, {"error": "session not found"})
        sid = parts[1]
        if len(parts) == 2 and method == "GET":
            return self.send_json(200, public(s))
        if len(parts) == 2 and method == "DELETE":
            del SESSIONS[sid]
            OUTPUTS.pop(sid, None)
            return self.send_json(204)
        if len(parts) == 3 and method == "POST":
            return self.action(s, parts[2], body)
        return self.send_json(404, {"error": "not found"})

    def action(self, s, action, body):
        sid = s["id"]
        if action == "read":
            if s["state"] == "snapshot":
                s["state"] = "active"
            out = OUTPUTS.get(sid, "")
            offset = int(body.get("offset") or 0)
            if offset > len(out):
                offset = 0
            s["lastAccess"] = now()
            return self.send_json(200, {"session": public(s), "payload": out[offset:], "nextOffset": len(out)})
        if action == "write":
            if s["state"] == "snapshot":
                return self.send_json(422, {"error": "session is snapshotted"})
            data = str(body.get("payload", ""))
            OUTPUTS[sid] += data + "\n"
            s["lastAccess"] = now()
            return self.send_json(200, {"session": public(s), "path": "/sessions/%s/input" % sid})
        if action == "switch":
            s["state"] = "active"
            s["lastAccess"] = now()
            return self.send_json(200, public(s))
        if action == "snapshot":
            s["state"] = "snapshot"
            s["checkpoint"] = {"sizeBytes": 2048, "createdAt": now()}
            return self.send_json(200, public(s))
        return self.send_json(404, {"error": "not found"})

    def do_GET(self):
        self.route("GET")

    def do_POST(self):
        self.route("POST")

    def do_DELETE(self):
        self.route("DELETE")


if __name__ == "__main__":
    ThreadingHTTPServer(("127.0.0.1", int(sys.argv[1])), Handler).serve_forever()
