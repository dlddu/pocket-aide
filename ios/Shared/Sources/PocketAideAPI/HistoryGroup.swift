import Foundation

/// PRD-10 AC13 그룹핑 결과. 같은 PR(있으면) 또는 같은 커밋(head_sha)에 도착한
/// `workflow_run` 이벤트들을 한 그룹으로 묶는다.
///
/// 그룹 키 우선순위:
///   1. PR 번호가 있으면 `pr:<repo>:<num>`
///   2. PR 없고 `head_sha`만 있으면 `sha:<repo>:<sha>`
///   3. 둘 다 없으면 `row:<id>` — 단독 그룹 (PRD-10 이전 row 백필되지 않은 케이스)
public struct HistoryGroup: Identifiable, Equatable, Sendable {
    public let id: String
    public let repoFullName: String
    public let prNumber: Int?
    public let prTitle: String?
    public let prURL: String?
    public let headBranch: String
    public let headSHA: String
    /// 그룹 내 항목 — 시각 역순(createdAt DESC).
    public let items: [NotificationHistoryItem]

    public var unacknowledgedCount: Int {
        items.reduce(0) { $0 + ($1.acknowledgedAt == nil ? 1 : 0) }
    }

    public var allAcknowledged: Bool { unacknowledgedCount == 0 }

    /// 그룹의 "가장 최근" 시각 — 그룹 간 정렬 키.
    public var latestCreatedAt: Int64 {
        items.map(\.createdAt).max() ?? 0
    }

    public var successCount: Int {
        items.reduce(0) { $0 + ($1.conclusion.lowercased() == "success" ? 1 : 0) }
    }

    public var failureCount: Int {
        items.reduce(0) { count, item in
            switch item.conclusion.lowercased() {
            case "failure", "timed_out", "cancelled", "action_required":
                return count + 1
            default:
                return count
            }
        }
    }

    /// 같은 그룹에 속하는 항목 중 "대표"로 노출할 항목 — 가장 최근(첫 번째).
    public var leadItem: NotificationHistoryItem? { items.first }
}

public enum HistoryGrouping {
    /// 평면 항목 리스트를 그룹으로 묶고, 정렬한다.
    ///
    /// - 그룹 내: createdAt DESC.
    /// - 그룹 간: 미확인 그룹 우선 → 각 섹션 내 latestCreatedAt DESC.
    ///
    /// 입력 순서에 의존하지 않으므로 호출자는 정렬 없이 raw 응답을 그대로 넘겨도 된다.
    public static func group(_ items: [NotificationHistoryItem]) -> [HistoryGroup] {
        guard !items.isEmpty else { return [] }

        var keys: [String] = []
        var byKey: [String: [NotificationHistoryItem]] = [:]
        for item in items {
            let key = groupKey(for: item)
            if byKey[key] == nil {
                keys.append(key)
            }
            byKey[key, default: []].append(item)
        }

        let groups: [HistoryGroup] = keys.map { key in
            let sorted = byKey[key]!.sorted { $0.createdAt > $1.createdAt }
            let lead = sorted.first!
            return HistoryGroup(
                id: key,
                repoFullName: lead.repoFullName,
                prNumber: lead.prNumber,
                prTitle: lead.prTitle,
                prURL: lead.prURL,
                headBranch: lead.headBranch,
                headSHA: lead.headSHA,
                items: sorted
            )
        }

        return groups.sorted { lhs, rhs in
            if lhs.allAcknowledged != rhs.allAcknowledged {
                return !lhs.allAcknowledged
            }
            return lhs.latestCreatedAt > rhs.latestCreatedAt
        }
    }

    /// 그룹 키 산출 — PR 우선, 없으면 head_sha, 둘 다 없으면 행 단독.
    public static func groupKey(for item: NotificationHistoryItem) -> String {
        if let n = item.prNumber {
            return "pr:\(item.repoFullName):\(n)"
        }
        if !item.headSHA.isEmpty {
            return "sha:\(item.repoFullName):\(item.headSHA)"
        }
        return "row:\(item.id)"
    }
}
