// MemoServiceProtocol.swift
// PocketAide

/// 메모 생성 서비스의 공개 인터페이스.
///
/// 프로덕션 구현체(`MemoService`)와 테스트용 대역(`MockMemoService`)
/// 모두 이 프로토콜을 준수합니다.
protocol MemoServiceProtocol {
    func create(content: String, source: String, serverURL: String, token: String) async throws -> Memo
}
