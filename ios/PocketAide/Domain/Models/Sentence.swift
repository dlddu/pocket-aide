// Sentence.swift
// PocketAide

import Foundation

struct SentenceCategory: Identifiable, Codable {
    let id: Int
    let name: String
}

struct Sentence: Identifiable, Codable {
    let id: Int
    let content: String
    let categoryId: Int

    enum CodingKeys: String, CodingKey {
        case id, content
        case categoryId = "category_id"
    }
}
