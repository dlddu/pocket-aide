// APIError.swift
// PocketAide

import Foundation

/// API 호출 중 발생할 수 있는 에러 타입.
public enum APIError: Error {
    case unauthorized
    case notFound
    case serverError(Int)
    case networkError(Error)
    case decodingError(Error)
}

extension APIError: Equatable {
    public static func == (lhs: APIError, rhs: APIError) -> Bool {
        switch (lhs, rhs) {
        case (.unauthorized, .unauthorized):
            return true
        case (.notFound, .notFound):
            return true
        case (.serverError(let l), .serverError(let r)):
            return l == r
        case (.networkError(let l), .networkError(let r)):
            return l.localizedDescription == r.localizedDescription
        case (.decodingError(let l), .decodingError(let r)):
            return l.localizedDescription == r.localizedDescription
        default:
            return false
        }
    }
}
