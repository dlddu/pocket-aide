// ChatMessage.swift
// PocketAide

import Foundation

/// 채팅 메시지를 나타내는 도메인 모델.
struct ChatMessage: Identifiable, Equatable {
    let id: UUID
    let role: Role
    let content: String

    init(id: UUID = UUID(), role: Role, content: String) {
        self.id = id
        self.role = role
        self.content = content
    }

    /// 메시지 역할 (사용자 또는 AI).
    enum Role: String, Equatable {
        case user
        case assistant
    }
}
