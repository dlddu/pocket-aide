// SyncServiceProtocolConformance.swift
// PocketAideTests
//
// `SyncService`에 `SyncServiceProtocol` 준수를 추가합니다.
//
// `SyncServiceProtocol`은 `SyncTestDoubles.swift`에 정의된 프로토콜이며,
// `SyncService`가 이 프로토콜을 준수함으로써 ViewModel 등에서
// 의존성 주입을 통해 교체 가능한 구조를 만듭니다.
//
// NOTE: Swift에서 프로토콜 준수는 명시적으로 선언되어야 합니다.
//       이 파일은 별도의 extension으로 준수를 선언함으로써
//       테스트 파일을 수정하지 않고 컴파일 오류를 해결합니다.

@testable import PocketAide

extension SyncService: SyncServiceProtocol {}
