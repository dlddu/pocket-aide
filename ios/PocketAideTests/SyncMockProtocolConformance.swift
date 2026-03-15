// SyncMockProtocolConformance.swift
// PocketAideTests
//
// `MockSyncAPIClient`와 `MockOfflineQueueForService`에
// SyncService가 요구하는 프로토콜 준수를 추가합니다.
//
// `MockSyncAPIClient`는 `SyncServiceTests.swift`에 정의된 테스트 대역이며,
// `SyncService` 생성자에 주입할 수 있도록 `SyncAPIClientProtocol`을 준수해야 합니다.
//
// `MockOfflineQueueForService`는 `SyncServiceTests.swift`에 정의된 테스트 대역이며,
// `SyncService` 생성자에 주입할 수 있도록 `SyncOfflineQueueProtocol`을 준수해야 합니다.
//
// NOTE: Swift에서 프로토콜 준수는 명시적으로 선언되어야 합니다.
//       이 파일은 별도의 extension으로 준수를 선언함으로써
//       테스트 파일을 수정하지 않고 컴파일 오류를 해결합니다.

@testable import PocketAide

extension MockSyncAPIClient: SyncAPIClientProtocol {}
extension MockOfflineQueueForService: SyncOfflineQueueProtocol {}
