// OfflineQueue.swift
// PocketAide
//
// DLD-740: 데이터 동기화 — UserDefaults 기반 오프라인 변경사항 큐

import Foundation

// MARK: - OfflineQueue

/// UserDefaults에 동기화 변경사항을 영속화하는 FIFO 큐.
///
/// 네트워크가 없을 때 발생한 변경사항을 보관하고,
/// 온라인 복귀 시 SyncService가 이를 서버로 전송합니다.
final class OfflineQueue {

    // MARK: - Constants

    private static let defaultsKey = "offline_sync_queue"

    // MARK: - Properties

    private let userDefaults: UserDefaults
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    /// 현재 큐에 저장된 변경사항 수.
    var count: Int {
        return dequeueAll().count
    }

    // MARK: - Init

    /// 지정된 UserDefaults를 사용해 큐를 초기화합니다.
    ///
    /// - Parameter userDefaults: 변경사항을 저장할 UserDefaults 인스턴스.
    init(userDefaults: UserDefaults = .standard) {
        self.userDefaults = userDefaults

        let enc = JSONEncoder()
        enc.dateEncodingStrategy = .iso8601
        self.encoder = enc

        let dec = JSONDecoder()
        dec.dateDecodingStrategy = .iso8601
        self.decoder = dec
    }

    // MARK: - Public Interface

    /// 변경사항을 큐의 끝에 추가합니다.
    ///
    /// - Parameter change: 큐에 추가할 SyncChange.
    func enqueue(_ change: SyncChange) {
        var current = loadChanges()
        current.append(change)
        saveChanges(current)
    }

    /// 큐에 저장된 모든 변경사항을 반환합니다.
    ///
    /// 이 메서드는 non-destructive입니다: 큐를 비우지 않습니다.
    /// - Returns: FIFO 순서로 정렬된 SyncChange 배열.
    func dequeueAll() -> [SyncChange] {
        return loadChanges()
    }

    /// 큐의 모든 변경사항을 제거합니다.
    func clearAll() {
        saveChanges([])
    }

    // MARK: - Private

    private func loadChanges() -> [SyncChange] {
        guard let data = userDefaults.data(forKey: Self.defaultsKey) else {
            return []
        }
        do {
            return try decoder.decode([SyncChange].self, from: data)
        } catch {
            return []
        }
    }

    private func saveChanges(_ changes: [SyncChange]) {
        do {
            let data = try encoder.encode(changes)
            userDefaults.set(data, forKey: Self.defaultsKey)
        } catch {
            // 직렬화 실패 시 기존 상태 유지.
        }
    }
}
