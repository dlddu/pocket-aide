// ChatService.swift
// PocketAide

import Foundation

/// 서버와 통신하여 채팅 메시지를 전송하고 히스토리를 조회하는 데이터 레이어 서비스.
final class ChatService {

    // MARK: - Response Types

    struct HistoryItem: Decodable {
        let role: String
        let content: String
    }

    // MARK: - Request Types

    private struct SendRequest: Encodable {
        let message: String
        let model: String
    }

    // MARK: - SSE Streaming

    /// 서버로 메시지를 전송하고 SSE 스트림을 통해 AI 응답을 토큰 단위로 수신합니다.
    ///
    /// - Parameters:
    ///   - message: 사용자 메시지
    ///   - model: 사용할 LLM 모델 이름 (빈 문자열이면 서버 기본값 사용)
    ///   - serverURL: 서버 주소
    ///   - token: JWT 인증 토큰
    ///   - onToken: 각 토큰이 수신될 때 호출되는 클로저
    func sendMessage(
        message: String,
        model: String,
        serverURL: String,
        token: String,
        onToken: @escaping (String) -> Void
    ) async throws -> String {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }

        let url = baseURL.appendingPathComponent("/chat/send")
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let body = SendRequest(message: message, model: model)
        urlRequest.httpBody = try JSONEncoder().encode(body)

        let (asyncBytes, response) = try await URLSession.shared.bytes(for: urlRequest)

        if let httpResponse = response as? HTTPURLResponse {
            switch httpResponse.statusCode {
            case 200..<300:
                break
            case 401:
                throw APIError.unauthorized
            default:
                throw APIError.serverError(httpResponse.statusCode)
            }
        }

        var fullResponse = ""

        for try await line in asyncBytes.lines {
            guard line.hasPrefix("data: ") else { continue }
            let data = String(line.dropFirst("data: ".count))
            guard data != "[DONE]" else { break }
            onToken(data)
            if fullResponse.isEmpty {
                fullResponse = data
            } else {
                fullResponse += " " + data
            }
        }

        return fullResponse
    }

    // MARK: - History

    /// 서버에서 채팅 히스토리를 조회합니다.
    ///
    /// - Parameters:
    ///   - serverURL: 서버 주소
    ///   - token: JWT 인증 토큰
    /// - Returns: 메시지 히스토리 배열
    func getHistory(serverURL: String, token: String) async throws -> [HistoryItem] {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }

        let client = APIClient(baseURL: baseURL)
        let items: [HistoryItem] = try await client.request(
            path: "/chat/history",
            method: .get,
            token: token
        )
        return items
    }
}
