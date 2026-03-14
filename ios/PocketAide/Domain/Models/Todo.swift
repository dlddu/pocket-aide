// Todo.swift
// PocketAide

import Foundation

/// 개인 투두 도메인 모델.
struct Todo: Identifiable, Codable {
    let id: Int
    let title: String
    let type: String
    let note: String?
    let priority: String?
    let completedAt: String?

    enum CodingKeys: String, CodingKey {
        case id, title, type, note, priority
        case completedAt = "completed_at"
    }

    /// 완료 여부를 반환합니다.
    var isCompleted: Bool { completedAt != nil }

    /// priority 값을 기준으로 정렬 가중치를 반환합니다 (낮을수록 높은 우선순위).
    var prioritySortOrder: Int {
        switch priority {
        case "high":   return 1
        case "medium": return 2
        case "low":    return 3
        default:       return 4
        }
    }
}
