// MemoServiceProtocol.swift
// PocketAide

/// 메모 CRUD 및 이동 서비스의 공개 인터페이스.
///
/// 프로덕션 구현체(`MemoService`)와 테스트용 대역(`MockMemoService`)
/// 모두 이 프로토콜을 준수합니다.
protocol MemoServiceProtocol {
    func list(serverURL: String, token: String) async throws -> [Memo]
    func create(content: String, source: String, serverURL: String, token: String) async throws -> Memo
    func delete(id: Int, serverURL: String, token: String) async throws
    func move(id: Int, target: String, serverURL: String, token: String) async throws
}
