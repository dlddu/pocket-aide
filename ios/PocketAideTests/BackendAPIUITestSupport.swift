// mock-exception: EXT — 실 OIDC IdP 대신 oidcmock 과 PKCE 왕복을 해 앱과 같은 사용자의 토큰을 받고, 그 토큰으로 실 백엔드 API 에 데이터를 시드한다 (docs/e2e-mocking-policy.md 허용목록)
import CryptoKit
import XCTest

struct BackendReply {
    let status: Int
    let body: Data
    let location: String?
    let failingURL: String?

    var text: String { String(bytes: body, encoding: .utf8) ?? "" }

    var json: [String: Any]? {
        (try? JSONSerialization.jsonObject(with: body)) as? [String: Any]
    }
}

private final class BackendReplyBox {
    var reply = BackendReply(status: 0, body: Data(), location: nil, failingURL: nil)
}

private final class BackendRedirectStopper: NSObject, URLSessionTaskDelegate {
    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        willPerformHTTPRedirection response: HTTPURLResponse,
        newRequest request: URLRequest,
        completionHandler: @escaping (URLRequest?) -> Void
    ) {
        completionHandler(nil)
    }
}

struct BackendAPI {
    static let issuer = "http://127.0.0.1:5556"
    static let server = "http://127.0.0.1:8080"
    static let clientID = "pocket-aide-ios-dev"
    static let redirectURI = "pocketaide-dev://callback"

    private static let session = URLSession(
        configuration: .ephemeral,
        delegate: BackendRedirectStopper(),
        delegateQueue: nil
    )

    let token: String

    static func signIn(subject: String? = nil, file: StaticString = #filePath, line: UInt = #line) -> BackendAPI? {
        let verifier = (UUID().uuidString + UUID().uuidString).replacingOccurrences(of: "-", with: "")
        let challenge = Data(SHA256.hash(data: Data(verifier.utf8))).base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
        var authorize = URLComponents(string: issuer + "/authorize")
        authorize?.queryItems = [
            URLQueryItem(name: "response_type", value: "code"),
            URLQueryItem(name: "client_id", value: clientID),
            URLQueryItem(name: "redirect_uri", value: redirectURI),
            URLQueryItem(name: "scope", value: "openid profile email"),
            URLQueryItem(name: "state", value: UUID().uuidString),
            URLQueryItem(name: "code_challenge", value: challenge),
            URLQueryItem(name: "code_challenge_method", value: "S256")
        ]
        if let subject {
            authorize?.queryItems?.append(URLQueryItem(name: "login_hint", value: subject))
        }
        guard let authorizeURL = authorize?.url, let tokenURL = URL(string: issuer + "/token") else {
            XCTFail("The oidcmock endpoints should form URLs", file: file, line: line)
            return nil
        }
        let redirect = send(URLRequest(url: authorizeURL))
        guard let code = authorizationCode(in: redirect) else {
            XCTFail("oidcmock /authorize should redirect with a code: \(redirect.status) \(redirect.text)", file: file, line: line)
            return nil
        }

        var form = URLComponents()
        form.queryItems = [
            URLQueryItem(name: "grant_type", value: "authorization_code"),
            URLQueryItem(name: "code", value: code),
            URLQueryItem(name: "code_verifier", value: verifier),
            URLQueryItem(name: "client_id", value: clientID),
            URLQueryItem(name: "redirect_uri", value: redirectURI)
        ]
        var request = URLRequest(url: tokenURL)
        request.httpMethod = "POST"
        request.httpBody = Data((form.percentEncodedQuery ?? "").utf8)
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
        let reply = send(request)
        guard reply.status == 200, let token = reply.json?["access_token"] as? String, !token.isEmpty else {
            XCTFail("oidcmock /token should issue an access token: \(reply.status) \(reply.text)", file: file, line: line)
            return nil
        }
        return BackendAPI(token: token)
    }

    func call(_ method: String, _ path: String, json: [String: Any]? = nil) -> BackendReply {
        guard let url = URL(string: Self.server + path) else {
            return BackendReply(status: 0, body: Data(), location: nil, failingURL: nil)
        }
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        if let json, let body = try? JSONSerialization.data(withJSONObject: json) {
            request.httpBody = body
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }
        return Self.send(request)
    }

    private static func authorizationCode(in reply: BackendReply) -> String? {
        for candidate in [reply.location, reply.failingURL].compactMap({ $0 }) {
            let value = URLComponents(string: candidate)?.queryItems?.first { $0.name == "code" }?.value
            if let value, !value.isEmpty {
                return value
            }
        }
        return nil
    }

    private static func send(_ request: URLRequest) -> BackendReply {
        let done = DispatchSemaphore(value: 0)
        let box = BackendReplyBox()
        session.dataTask(with: request) { data, response, error in
            let http = response as? HTTPURLResponse
            let failing = (error as NSError?)?.userInfo[NSURLErrorFailingURLStringErrorKey] as? String
            let detail = data ?? Data(error.map { String(describing: $0) }?.utf8 ?? "".utf8)
            box.reply = BackendReply(
                status: http?.statusCode ?? 0,
                body: detail,
                location: http?.value(forHTTPHeaderField: "Location"),
                failingURL: failing
            )
            done.signal()
        }.resume()
        XCTAssertTrue(done.wait(timeout: .now() + 20) == .success, "\(request.url?.absoluteString ?? "") should answer")
        return box.reply
    }
}
