// SyncService.swift
// PocketAide
//
// DLD-740: 데이터 동기화 — POST /sync API 호출 서비스

import Foundation

// MARK: - SyncAPIClientProtocol

/// SyncService가 의존하는 API 클라이언트 프로토콜.
/// 테스트에서 MockSyncAPIClient로 대체할 수 있도록 추상화합니다.
protocol SyncAPIClientProtocol {
    func performSync(request: SyncRequest, token: String, serverURL: String) async throws -> SyncResponse
}

// MARK: - SyncOfflineQueueProtocol

/// SyncService가 의존하는 오프라인 큐 프로토콜.
/// 테스트에서 MockOfflineQueueForService로 대체할 수 있도록 추상화합니다.
protocol SyncOfflineQueueProtocol {
    func dequeueAll() -> [SyncChange]
    func clearAll()
}

// MARK: - OfflineQueue + SyncOfflineQueueProtocol

extension OfflineQueue: SyncOfflineQueueProtocol {}

// MARK: - DefaultSyncAPIClient

/// SyncAPIClientProtocol의 기본 구현체.
/// 실제 네트워크 호출을 통해 POST /sync를 수행합니다.
final class DefaultSyncAPIClient: SyncAPIClientProtocol {

    private let encoder: JSONEncoder
    private let decoder: JSONDecoder
    private let session: URLSession

    init(session: URLSession = .shared) {
        self.session = session

        let enc = JSONEncoder()
        enc.dateEncodingStrategy = .iso8601
        self.encoder = enc

        let dec = JSONDecoder()
        dec.dateDecodingStrategy = .iso8601
        self.decoder = dec
    }

    func performSync(request: SyncRequest, token: String, serverURL: String) async throws -> SyncResponse {
        guard let baseURL = URL(string: serverURL) else {
            throw APIError.networkError(URLError(.badURL))
        }
        let url = baseURL.appendingPathComponent("/sync")
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            urlRequest.httpBody = try encoder.encode(request)
        } catch {
            throw APIError.networkError(error)
        }

        let data: Data
        let response: URLResponse

        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch {
            throw APIError.networkError(error)
        }

        if let httpResponse = response as? HTTPURLResponse {
            switch httpResponse.statusCode {
            case 200..<300:
                break
            case 401:
                throw APIError.unauthorized
            case 404:
                throw APIError.notFound
            default:
                throw APIError.serverError(httpResponse.statusCode)
            }
        }

        do {
            return try decoder.decode(SyncResponse.self, from: data)
        } catch {
            throw APIError.decodingError(error)
        }
    }
}

// MARK: - SyncService

/// 오프라인 큐의 변경사항을 서버에 업로드하고 서버 최신 데이터를 수신하는 서비스.
///
/// SyncAPIClientProtocol과 SyncOfflineQueueProtocol 타입 파라미터를 통해
/// 테스트에서 mock 의존성을 주입할 수 있습니다.
///
/// NOTE: SyncServiceProtocol 준수는 SyncTestDoubles.swift에서 정의하는
///       동일한 프로토콜을 통해 테스트 대상으로 동작합니다.
final class SyncService {

    // MARK: - Properties

    private let syncOperation: (SyncRequest, String, String) async throws -> SyncResponse
    private let queueDequeueAll: () -> [SyncChange]
    private let queueClearAll: () -> Void

    // MARK: - Init

    /// APIClient와 OfflineQueue를 주입받아 SyncService를 초기화합니다.
    ///
    /// - Parameters:
    ///   - apiClient: POST /sync를 수행할 API 클라이언트 (SyncAPIClientProtocol 준수 타입).
    ///   - offlineQueue: 오프라인 변경사항을 보관하는 큐 (SyncOfflineQueueProtocol 준수 타입).
    init<APIClient: SyncAPIClientProtocol, Queue: SyncOfflineQueueProtocol>(
        apiClient: APIClient,
        offlineQueue: Queue
    ) {
        self.syncOperation = { request, token, serverURL in
            try await apiClient.performSync(request: request, token: token, serverURL: serverURL)
        }
        self.queueDequeueAll = { offlineQueue.dequeueAll() }
        self.queueClearAll = { offlineQueue.clearAll() }
    }

    /// 기본 구현체로 SyncService를 초기화합니다.
    convenience init() {
        self.init(
            apiClient: DefaultSyncAPIClient(),
            offlineQueue: OfflineQueue()
        )
    }

    // MARK: - SyncServiceProtocol

    /// 오프라인 큐의 변경사항을 서버에 업로드하고 서버 데이터를 반환합니다.
    ///
    /// - Parameters:
    ///   - token: Bearer 인증 토큰.
    ///   - serverURL: 서버의 베이스 URL 문자열.
    /// - Returns: 서버의 최신 SyncServerData.
    /// - Throws: 네트워크 오류 또는 인증 오류 발생 시 throw. 실패 시 큐는 보존됩니다.
    func sync(token: String, serverURL: String) async throws -> SyncServerData {
        let changes = queueDequeueAll()
        let request = SyncRequest(changes: changes)

        let response = try await syncOperation(request, token, serverURL)

        // 성공 시에만 큐를 비웁니다.
        queueClearAll()

        return response.serverData
    }
}
