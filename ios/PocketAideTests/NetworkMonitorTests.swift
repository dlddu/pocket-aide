// NetworkMonitorTests.swift
// PocketAideTests
//
// Unit tests for NetworkMonitor.
//
// DLD-740: 데이터 동기화 — NetworkMonitor unit 테스트
//
// 테스트 대상:
//   - NetworkMonitor 초기화 (isConnected 초기값)
//   - NetworkMonitor.isConnected 상태 (Bool Published)
//   - NetworkMonitor.onReconnect 콜백 호출 시점 (오프라인 → 온라인 전환 시)
//   - NWPathMonitor 의존성 주입 가능 구조 검증
//
// NOTE: TDD Red Phase — NetworkMonitor가 아직 구현되지 않았으므로
//       이 테스트는 실패 상태입니다. 구현 후 통과해야 합니다.
//
// 구현 필요 파일: ios/PocketAide/Data/Sync/NetworkMonitor.swift
//
// NetworkMonitor 요구사항:
//   - @Observable 또는 ObservableObject 준수
//   - var isConnected: Bool { get }
//   - func start() — 모니터링 시작
//   - func stop() — 모니터링 중지
//   - var onReconnect: (() -> Void)? — 오프라인→온라인 전환 시 호출되는 콜백
//   - 테스트 가능하도록 NWPathMonitor 의존성을 주입받거나
//     isConnected를 외부에서 설정 가능한 구조여야 함
//
// 네이티브 플랫폼: macOS (Mac Native 전용)
// 테스트 프레임워크: XCTest

import XCTest
import Combine
@testable import PocketAide

// MARK: - NetworkMonitorTests

final class NetworkMonitorTests: XCTestCase {

    // MARK: - Properties

    private var cancellables: Set<AnyCancellable> = []

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        cancellables = []
    }

    override func tearDown() {
        cancellables = []
        super.tearDown()
    }

    // MARK: - Initialisation

    /// NetworkMonitor는 초기화 시 isConnected 프로퍼티를 가져야 한다.
    ///
    /// Scenario:
    ///   NetworkMonitor() 생성
    ///   → isConnected 프로퍼티 접근 가능 (Bool)
    func test_init_isConnectedPropertyIsAccessible() {
        // Arrange & Act
        let sut = NetworkMonitor()

        // Assert
        // isConnected가 Bool 타입으로 접근 가능한지 확인
        let isConnected: Bool = sut.isConnected
        // 초기값은 구현에 따라 다를 수 있으므로 값 자체보다는 접근 가능성만 검증
        _ = isConnected
    }

    /// NetworkMonitor는 start()와 stop() 메서드를 가져야 한다.
    ///
    /// Scenario:
    ///   start()/stop() 호출이 가능해야 함 (크래시 없이)
    func test_init_startAndStopAreCallable() {
        // Arrange
        let sut = NetworkMonitor()

        // Act & Assert (must not crash)
        sut.start()
        sut.stop()
    }

    // MARK: - isConnected 상태 변경

    /// isConnected를 false로 설정하면 해당 값이 반영되어야 한다.
    ///
    /// Scenario:
    ///   sut.simulateDisconnect() 또는 sut.isConnected = false
    ///   → isConnected == false
    func test_isConnected_whenOffline_returnsFalse() {
        // Arrange
        let sut = NetworkMonitor()

        // Act: 테스트 목적의 연결 끊김 시뮬레이션
        sut.simulateConnectionChange(isConnected: false)

        // Assert
        XCTAssertFalse(sut.isConnected, "isConnected should be false after simulating disconnect")
    }

    /// isConnected를 true로 설정하면 해당 값이 반영되어야 한다.
    ///
    /// Scenario:
    ///   sut.simulateConnect() 또는 sut.isConnected = true
    ///   → isConnected == true
    func test_isConnected_whenOnline_returnsTrue() {
        // Arrange
        let sut = NetworkMonitor()

        // Act
        sut.simulateConnectionChange(isConnected: true)

        // Assert
        XCTAssertTrue(sut.isConnected, "isConnected should be true after simulating connection")
    }

    // MARK: - onReconnect 콜백

    /// 오프라인에서 온라인으로 전환될 때 onReconnect 콜백이 호출되어야 한다.
    ///
    /// Scenario:
    ///   오프라인 상태 → 온라인 전환
    ///   → onReconnect 콜백이 정확히 1회 호출됨
    func test_onReconnect_calledWhenTransitioningOfflineToOnline() {
        // Arrange
        let sut = NetworkMonitor()
        var reconnectCallCount = 0
        sut.onReconnect = { reconnectCallCount += 1 }

        // Act: 오프라인 상태에서 시작
        sut.simulateConnectionChange(isConnected: false)
        // 온라인으로 복귀
        sut.simulateConnectionChange(isConnected: true)

        // Assert
        XCTAssertEqual(reconnectCallCount, 1,
                       "onReconnect should be called exactly once when transitioning offline → online")
    }

    /// 이미 온라인인 상태에서 온라인으로 변경되면 onReconnect가 호출되지 않아야 한다.
    ///
    /// Scenario:
    ///   온라인 → 온라인 (상태 변화 없음)
    ///   → onReconnect 호출되지 않음
    func test_onReconnect_notCalledWhenAlreadyOnline() {
        // Arrange
        let sut = NetworkMonitor()
        var reconnectCallCount = 0
        sut.onReconnect = { reconnectCallCount += 1 }

        // Act: 이미 온라인 상태에서 다시 온라인으로 설정
        sut.simulateConnectionChange(isConnected: true)
        sut.simulateConnectionChange(isConnected: true)

        // Assert
        XCTAssertEqual(reconnectCallCount, 0,
                       "onReconnect should NOT be called when already online (no transition)")
    }

    /// 온라인에서 오프라인으로 전환될 때 onReconnect가 호출되지 않아야 한다.
    ///
    /// Scenario:
    ///   온라인 → 오프라인 전환
    ///   → onReconnect 호출되지 않음
    func test_onReconnect_notCalledWhenGoingOffline() {
        // Arrange
        let sut = NetworkMonitor()
        var reconnectCallCount = 0
        sut.onReconnect = { reconnectCallCount += 1 }
        sut.simulateConnectionChange(isConnected: true)

        // Act
        sut.simulateConnectionChange(isConnected: false)

        // Assert
        XCTAssertEqual(reconnectCallCount, 0,
                       "onReconnect should NOT be called when going offline")
    }

    /// 여러 번 오프라인/온라인 전환 시 온라인 복귀 횟수만큼 onReconnect가 호출되어야 한다.
    ///
    /// Scenario:
    ///   offline → online → offline → online (2번 온라인 복귀)
    ///   → onReconnect 2회 호출
    func test_onReconnect_calledEachTimeReconnecting() {
        // Arrange
        let sut = NetworkMonitor()
        var reconnectCallCount = 0
        sut.onReconnect = { reconnectCallCount += 1 }

        // Act: 2 reconnect cycles
        sut.simulateConnectionChange(isConnected: false)
        sut.simulateConnectionChange(isConnected: true)   // +1
        sut.simulateConnectionChange(isConnected: false)
        sut.simulateConnectionChange(isConnected: true)   // +2

        // Assert
        XCTAssertEqual(reconnectCallCount, 2,
                       "onReconnect should be called once per offline→online transition")
    }

    // MARK: - Combine Published 검증

    /// NetworkMonitor의 isConnected 변경사항을 Combine으로 구독할 수 있어야 한다.
    ///
    /// Scenario:
    ///   isConnected 변경 → Combine subscriber가 새 값을 수신
    func test_isConnected_combinePublisher_receivesUpdates() {
        // Arrange
        let sut = NetworkMonitor()
        let expectation = expectation(description: "Combine subscriber receives isConnected update")
        var receivedValues: [Bool] = []

        sut.isConnectedPublisher
            .dropFirst() // 초기값 무시
            .sink { value in
                receivedValues.append(value)
                if receivedValues.count >= 1 {
                    expectation.fulfill()
                }
            }
            .store(in: &cancellables)

        // Act
        sut.simulateConnectionChange(isConnected: false)

        // Assert
        waitForExpectations(timeout: 1.0)
        XCTAssertFalse(receivedValues.isEmpty, "Combine subscriber should receive at least one update")
        XCTAssertEqual(receivedValues.last, false, "Last received value should be false (offline)")
    }

    // MARK: - onReconnect 콜백 nil 처리

    /// onReconnect 콜백이 nil일 때 온라인 전환 시 크래시가 발생하지 않아야 한다.
    ///
    /// Scenario:
    ///   onReconnect = nil
    ///   오프라인 → 온라인 전환
    ///   → 크래시 없음
    func test_onReconnect_whenNil_doesNotCrash() {
        // Arrange
        let sut = NetworkMonitor()
        sut.onReconnect = nil
        sut.simulateConnectionChange(isConnected: false)

        // Act & Assert (must not crash)
        sut.simulateConnectionChange(isConnected: true)
    }
}
