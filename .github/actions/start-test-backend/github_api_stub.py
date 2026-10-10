# mock-exception: EXT — 실 GitHub 은 전용 테스트 계정 PAT 가 CI 시크릿에 없어 E2E 가 부를 수 없다; 앱의 GitHubClient 가 부르는 GET /user · POST /graphql 만 실 응답 모양으로 서빙한다 (docs/e2e-mocking-policy.md)
import json
import socket
import sys
import threading
import time
from datetime import datetime, timedelta, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

LOGIN = "pocket-aide-e2e"
REPO = "dlddu/pocket-aide-e2e"

TOKENS = {
    "ghp_e2e_valid": "ok",
    "ghp_e2e_revoked": "revoked",
    "ghp_e2e_expires": "expires",
    "ghp_e2e_ratelimited": "ratelimited",
    "ghp_e2e_empty": "empty",
    "ghp_e2e_slow": "slow",
    "ghp_e2e_rollup": "rollup",
}

SCHEDULED_TOKENS = {
    "ghp_e2e_closing_": "closing",
    "ghp_e2e_rerun_": "rerun",
}

SLOW_SECONDS = 20

FAILING_CHECKS = {"FAILURE", "ERROR", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED", "STARTUP_FAILURE"}
RUNNING_CHECKS = {"EXPECTED", "PENDING", "QUEUED", "IN_PROGRESS", "WAITING", "REQUESTED"}

SEARCH_COUNTS = {}
SEARCH_COUNTS_LOCK = threading.Lock()

SEARCH_ALIASES = {
    "authored": "author:{login}",
    "requested": "review-requested:{login}",
    "reviewed": "reviewed-by:{login} -author:{login}",
}


def iso(minutes_ago):
    at = datetime.now(timezone.utc) - timedelta(minutes=minutes_ago)
    return at.strftime("%Y-%m-%dT%H:%M:%SZ")


def pull_request(number, title, author, minutes_ago, rollup):
    status = None if rollup is None else {"state": rollup}
    return {
        "number": number,
        "title": title,
        "url": f"https://github.com/{REPO}/pull/{number}",
        "updatedAt": iso(minutes_ago),
        "repository": {"nameWithOwner": REPO},
        "author": {"login": author},
        "commits": {"nodes": [{"commit": {"statusCheckRollup": status}}]},
    }


def rollup_of(checks):
    if not checks:
        return None
    if any(check in FAILING_CHECKS for check in checks):
        return "FAILURE"
    if any(check in RUNNING_CHECKS for check in checks):
        return "PENDING"
    return "SUCCESS"


def rollup_payload():
    payload = empty_payload()
    payload["data"]["authored"]["nodes"] = [
        pull_request(number, title, LOGIN, minutes_ago, rollup_of(checks))
        for number, title, minutes_ago, checks in (
            (41, "e2e checks all passed", 1, ["SUCCESS", "SUCCESS"]),
            (42, "e2e one check failed", 2, ["SUCCESS", "FAILURE"]),
            (43, "e2e checks running", 3, ["IN_PROGRESS"]),
            (44, "e2e no checks", 4, []),
            (45, "e2e passed and running", 5, ["SUCCESS", "IN_PROGRESS"]),
            (46, "e2e running and one failed", 6, ["SUCCESS", "IN_PROGRESS", "FAILURE"]),
        )
    ]
    return payload


def search_payload():
    return {
        "data": {
            "authored": {"nodes": [
                pull_request(11, "e2e authored passing", LOGIN, 5, "SUCCESS"),
                pull_request(14, "e2e authored without checks", LOGIN, 20, None),
            ]},
            "requested": {"nodes": [
                pull_request(12, "e2e review requested failing", "octocat", 10, "FAILURE"),
                None,
            ]},
            "reviewed": {"nodes": [
                pull_request(13, "e2e reviewed pending", "hubot", 15, "PENDING"),
            ]},
        },
        "errors": [{
            "type": "FORBIDDEN",
            "path": ["requested", "nodes", 1],
            "extensions": {"saml_failure": True},
            "locations": [{"line": 3, "column": 3}],
            "message": "Resource protected by organization SAML enforcement. You must grant your Personal Access token access to this organization.",
        }],
    }


def empty_payload():
    return {"data": {alias: {"nodes": []} for alias in SEARCH_ALIASES}}


def closing_payload(search_number):
    if search_number > 1:
        return empty_payload()
    payload = empty_payload()
    payload["data"]["authored"]["nodes"] = [
        pull_request(21, "e2e authored to be merged", LOGIN, 3, "SUCCESS"),
    ]
    payload["data"]["requested"]["nodes"] = [
        pull_request(22, "e2e review requested to be closed", "octocat", 7, "PENDING"),
    ]
    return payload


def rerun_payload(search_number):
    payload = empty_payload()
    if search_number > 1:
        payload["data"]["authored"]["nodes"] = [pull_request(31, "e2e authored rerun", LOGIN, 0, "SUCCESS")]
    else:
        payload["data"]["authored"]["nodes"] = [pull_request(31, "e2e authored rerun", LOGIN, 30, "PENDING")]
    return payload


def next_search_number(token):
    with SEARCH_COUNTS_LOCK:
        SEARCH_COUNTS[token] = SEARCH_COUNTS.get(token, 0) + 1
        return SEARCH_COUNTS[token]


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def send_json(self, status, body, headers=None):
        data = json.dumps(body).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        for key, value in (headers or {}).items():
            self.send_header(key, value)
        self.end_headers()
        self.wfile.write(data)

    def bad_credentials(self):
        self.send_json(401, {
            "message": "Bad credentials",
            "documentation_url": "https://docs.github.com/rest",
            "status": "401",
        })

    def token(self):
        auth = self.headers.get("Authorization", "")
        if not auth.startswith("Bearer "):
            return None
        return auth[len("Bearer "):]

    def token_behavior(self):
        token = self.token()
        if token is None:
            return None
        if token in TOKENS:
            return TOKENS[token]
        for prefix, behavior in SCHEDULED_TOKENS.items():
            if token.startswith(prefix) and len(token) > len(prefix):
                return behavior
        return None

    def do_GET(self):
        if self.path != "/user":
            self.send_json(404, {"message": "Not Found", "status": "404"})
            return
        behavior = self.token_behavior()
        if behavior in (None, "revoked"):
            self.bad_credentials()
            return
        self.send_json(200, {"login": LOGIN, "id": 9001, "type": "User"})

    def do_POST(self):
        length = int(self.headers.get("Content-Length") or 0)
        raw = self.rfile.read(length)
        if self.path != "/graphql":
            self.send_json(404, {"message": "Not Found", "status": "404"})
            return
        behavior = self.token_behavior()
        if behavior in (None, "revoked", "expires"):
            self.bad_credentials()
            return
        if behavior == "ratelimited":
            self.send_json(403, {
                "message": "API rate limit exceeded for user ID 9001.",
                "documentation_url": "https://docs.github.com/rest/overview/rate-limits-for-the-rest-api",
                "status": "403",
            }, {
                "X-RateLimit-Limit": "5000",
                "X-RateLimit-Remaining": "0",
                "X-RateLimit-Reset": str(int(time.time()) + 600),
                "X-RateLimit-Resource": "graphql",
            })
            return
        problem = self.query_problem(raw)
        if problem:
            self.send_json(200, {"errors": [{"message": problem}]})
            return
        if behavior == "empty":
            self.send_json(200, empty_payload())
            return
        if behavior == "closing":
            self.send_json(200, closing_payload(next_search_number(self.token())))
            return
        if behavior == "rerun":
            self.send_json(200, rerun_payload(next_search_number(self.token())))
            return
        if behavior == "rollup":
            self.send_json(200, rollup_payload())
            return
        if behavior == "slow":
            time.sleep(SLOW_SECONDS)
        self.send_json(200, search_payload())

    def query_problem(self, raw):
        try:
            body = json.loads(raw)
        except ValueError:
            return "Problems parsing JSON"
        query = body.get("query") or ""
        variables = body.get("variables") or {}
        for alias, qualifier in SEARCH_ALIASES.items():
            if f"{alias}: search(query: ${alias}, type: ISSUE" not in query:
                return f"query has no {alias} search"
            expected = "is:pr is:open archived:false " + qualifier.format(login=LOGIN)
            if variables.get(alias) != expected:
                return f"variable {alias} is {variables.get(alias)!r}, expected {expected!r}"
        for field in ("nameWithOwner", "statusCheckRollup", "updatedAt"):
            if field not in query:
                return f"query does not select {field}"
        return None

    def log_message(self, fmt, *args):
        sys.stderr.write("%s %s\n" % (self.log_date_time_string(), fmt % args))
        sys.stderr.flush()


class DualStackServer(ThreadingHTTPServer):
    address_family = socket.AF_INET6

    def server_bind(self):
        self.socket.setsockopt(socket.IPPROTO_IPV6, socket.IPV6_V6ONLY, 0)
        super().server_bind()


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 5557
    DualStackServer(("::", port), Handler).serve_forever()
