// NetworkMonitor.swift
// PocketAide
//
// DLD-740: 데이터 동기화 — NWPathMonitor 래퍼

import Foundation
import Network
import Combine

// MARK: - NetworkMonitor

/// NWPathMonitor를 래핑하여 네트워크 연결 상태를 관찰하는 클래스.
///
/// 오프라인 → 온라인 전환 시 `onReconnect` 콜백이 호출됩니다.
/// `isConnectedPublisher`를 통해 Combine으로도 상태를 구독할 수 있습니다.
final class NetworkMonitor: ObservableObject {

    // MARK: - Properties

    /// 현재 네트워크 연결 상태.
    @Published private(set) var isConnected: Bool = true

    /// 오프라인 → 온라인 전환 시 호출되는 콜백.
    var onReconnect: (() -> Void)?

    /// isConnected의 AnyPublisher.
    var isConnectedPublisher: AnyPublisher<Bool, Never> {
        $isConnected.eraseToAnyPublisher()
    }

    // MARK: - Private Properties

    private let monitor: NWPathMonitor
    private let queue: DispatchQueue

    // MARK: - Init

    init() {
        self.monitor = NWPathMonitor()
        self.queue = DispatchQueue(label: "NetworkMonitor", qos: .utility)
    }

    // MARK: - Public Interface

    /// 네트워크 모니터링을 시작합니다.
    func start() {
        monitor.pathUpdateHandler = { [weak self] path in
            guard let self else { return }
            let newValue = path.status == .satisfied
            DispatchQueue.main.async {
                self.updateConnectionState(newValue)
            }
        }
        monitor.start(queue: queue)
    }

    /// 네트워크 모니터링을 중지합니다.
    func stop() {
        monitor.cancel()
    }

    /// 테스트 목적의 연결 상태 변경 시뮬레이션.
    ///
    /// 오프라인 → 온라인 전환 시 `onReconnect`가 호출됩니다.
    /// - Parameter isConnected: 새 연결 상태.
    func simulateConnectionChange(isConnected: Bool) {
        updateConnectionState(isConnected)
    }

    // MARK: - Private

    private func updateConnectionState(_ newValue: Bool) {
        let wasConnected = isConnected
        isConnected = newValue

        // 오프라인 → 온라인 전환 시에만 onReconnect 호출.
        if !wasConnected && newValue {
            onReconnect?()
        }
    }
}
