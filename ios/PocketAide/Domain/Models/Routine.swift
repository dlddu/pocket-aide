// Routine.swift
// PocketAide

import Foundation

/// 루틴 도메인 모델.
struct Routine: Identifiable, Codable {
    let id: Int
    let name: String
    let intervalDays: Int
    let lastDoneAt: String
    let nextDueDate: String
    let dDay: Int
    let note: String?
    let notifyEnabled: Bool?

    enum CodingKeys: String, CodingKey {
        case id, name
        case intervalDays = "interval_days"
        case lastDoneAt = "last_done_at"
        case nextDueDate = "next_due_date"
        case dDay = "d_day"
        case note
        case notifyEnabled = "notify_enabled"
    }
}
