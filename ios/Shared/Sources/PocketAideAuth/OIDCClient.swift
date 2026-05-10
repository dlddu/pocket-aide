import AuthenticationServices
import CryptoKit
import Foundation
import PocketAideAPI
import PocketAideStorage
#if canImport(UIKit)
import UIKit
#endif

public enum OIDCError: Error, CustomStringConvertible {
    case discovery(Error)
    case noCode
    case tokenExchange(Int, String)
    case decoding(Error)
    case cancelled
    case missingPresentationAnchor

    public var description: String {
        switch self {
        case .discovery(let e): return "discovery: \(e)"
        case .noCode: return "no code in callback"
        case .tokenExchange(let s, let body): return "token \(s): \(body)"
        case .decoding(let e): return "decode: \(e)"
        case .cancelled: return "cancelled"
        case .missingPresentationAnchor: return "no presentation anchor"
        }
    }
}

@MainActor
public final class OIDCClient: NSObject, ASWebAuthenticationPresentationContextProviding {
    private let api: APIClient
    private let tokenStore: TokenStoring
    private let session: URLSession
    private weak var presentationAnchor: ASPresentationAnchor?

    public init(api: APIClient, tokenStore: TokenStoring, session: URLSession = .shared) {
        self.api = api
        self.tokenStore = tokenStore
        self.session = session
    }

    public func setPresentationAnchor(_ anchor: ASPresentationAnchor?) {
        self.presentationAnchor = anchor
    }

    public func signIn() async throws -> TokenBundle {
        let cfg = try await api.authConfig()
        let discovery = try await fetchDiscovery(issuer: cfg.issuer)

        let verifier = makeCodeVerifier()
        let challenge = codeChallenge(for: verifier)
        let state = randomState()

        var components = URLComponents(string: discovery.authorizationEndpoint)!
        var query = components.queryItems ?? []
        query.append(contentsOf: [
            URLQueryItem(name: "response_type", value: "code"),
            URLQueryItem(name: "client_id", value: cfg.clientId),
            URLQueryItem(name: "redirect_uri", value: cfg.redirectUri),
            URLQueryItem(name: "scope", value: "openid profile email offline_access"),
            URLQueryItem(name: "state", value: state),
            URLQueryItem(name: "code_challenge", value: challenge),
            URLQueryItem(name: "code_challenge_method", value: "S256"),
        ])
        components.queryItems = query
        let authURL = components.url!

        let scheme = URL(string: cfg.redirectUri)?.scheme
        let callback = try await runWebAuth(url: authURL, scheme: scheme)

        guard let callbackComponents = URLComponents(url: callback, resolvingAgainstBaseURL: false),
              let code = callbackComponents.queryItems?.first(where: { $0.name == "code" })?.value else {
            throw OIDCError.noCode
        }

        let bundle = try await exchangeCode(
            code: code,
            verifier: verifier,
            cfg: cfg,
            tokenEndpoint: discovery.tokenEndpoint
        )
        try tokenStore.save(bundle)
        return bundle
    }

    public func signOut() throws {
        try tokenStore.clear()
    }

    private func runWebAuth(url: URL, scheme: String?) async throws -> URL {
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<URL, Error>) in
            let session = ASWebAuthenticationSession(url: url, callbackURLScheme: scheme) { callbackURL, error in
                if let error {
                    if let asError = error as? ASWebAuthenticationSessionError, asError.code == .canceledLogin {
                        cont.resume(throwing: OIDCError.cancelled)
                    } else {
                        cont.resume(throwing: error)
                    }
                    return
                }
                guard let callbackURL else {
                    cont.resume(throwing: OIDCError.noCode)
                    return
                }
                cont.resume(returning: callbackURL)
            }
            session.presentationContextProvider = self
            session.prefersEphemeralWebBrowserSession = false
            if !session.start() {
                cont.resume(throwing: OIDCError.cancelled)
            }
        }
    }

    private func exchangeCode(code: String, verifier: String, cfg: AuthConfig, tokenEndpoint: String) async throws -> TokenBundle {
        var req = URLRequest(url: URL(string: tokenEndpoint)!)
        req.httpMethod = "POST"
        req.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
        let body = [
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": cfg.redirectUri,
            "client_id": cfg.clientId,
            "code_verifier": verifier,
        ]
        req.httpBody = formEncode(body).data(using: .utf8)

        let (data, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else {
            throw OIDCError.tokenExchange(-1, "no http response")
        }
        guard (200..<300).contains(http.statusCode) else {
            throw OIDCError.tokenExchange(http.statusCode, String(data: data, encoding: .utf8) ?? "")
        }
        do {
            let parsed = try JSONDecoder().decode(TokenResponse.self, from: data)
            let expiresAt = Date().addingTimeInterval(TimeInterval(parsed.expiresIn ?? 3600))
            return TokenBundle(accessToken: parsed.accessToken, refreshToken: parsed.refreshToken, expiresAt: expiresAt)
        } catch {
            throw OIDCError.decoding(error)
        }
    }

    private func fetchDiscovery(issuer: String) async throws -> Discovery {
        let url = URL(string: issuer.trimmingCharacters(in: CharacterSet(charactersIn: "/")) + "/.well-known/openid-configuration")!
        do {
            let (data, _) = try await session.data(from: url)
            return try JSONDecoder().decode(Discovery.self, from: data)
        } catch {
            throw OIDCError.discovery(error)
        }
    }

    public func presentationAnchor(for _: ASWebAuthenticationSession) -> ASPresentationAnchor {
        if let presentationAnchor {
            return presentationAnchor
        }
        #if canImport(UIKit)
        let scene = UIApplication.shared.connectedScenes
            .compactMap { $0 as? UIWindowScene }
            .first(where: { $0.activationState == .foregroundActive })
        if let window = scene?.windows.first(where: { $0.isKeyWindow }) ?? scene?.windows.first {
            return window
        }
        #endif
        return ASPresentationAnchor()
    }

    private struct Discovery: Codable {
        let issuer: String
        let authorizationEndpoint: String
        let tokenEndpoint: String

        enum CodingKeys: String, CodingKey {
            case issuer
            case authorizationEndpoint = "authorization_endpoint"
            case tokenEndpoint = "token_endpoint"
        }
    }

    private struct TokenResponse: Codable {
        let accessToken: String
        let refreshToken: String?
        let expiresIn: Int?
        let tokenType: String?
        let idToken: String?

        enum CodingKeys: String, CodingKey {
            case accessToken = "access_token"
            case refreshToken = "refresh_token"
            case expiresIn = "expires_in"
            case tokenType = "token_type"
            case idToken = "id_token"
        }
    }
}

private func makeCodeVerifier() -> String {
    var bytes = [UInt8](repeating: 0, count: 64)
    _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
    return Data(bytes).base64URLEncodedString()
}

private func codeChallenge(for verifier: String) -> String {
    let hash = SHA256.hash(data: Data(verifier.utf8))
    return Data(hash).base64URLEncodedString()
}

private func randomState() -> String {
    var bytes = [UInt8](repeating: 0, count: 16)
    _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
    return Data(bytes).base64URLEncodedString()
}

private func formEncode(_ pairs: [String: String]) -> String {
    pairs.map { k, v in
        let kk = k.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? k
        let vv = v.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? v
        return "\(kk)=\(vv)"
    }.joined(separator: "&")
}

private extension Data {
    func base64URLEncodedString() -> String {
        base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}
