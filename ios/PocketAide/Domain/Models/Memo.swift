// Memo.swift
// PocketAide

import Foundation

/// 임시 공간 메모 도메인 모델.
struct Memo: Identifiable, Codable {
    let id: Int
    let content: String
    let source: String

    enum CodingKeys: String, CodingKey {
        case id, content, source
    }
}
