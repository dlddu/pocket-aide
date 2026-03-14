// Todo.swift
// PocketAide

import Foundation

/// 개인 투두 도메인 모델.
struct Todo: Identifiable, Codable {
    let id: Int
    let title: String
    let type: String
    let completedAt: String?

    enum CodingKeys: String, CodingKey {
        case id, title, type
        case completedAt = "completed_at"
    }

    /// 완료 여부를 반환합니다.
    var isCompleted: Bool { completedAt != nil }
}
