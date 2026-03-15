// AppNotification.swift
// PocketAide

import Foundation

/// 알림 모음(Notifications) 도메인 모델.
///
/// App Group UserDefaults (group.com.dlddu.PocketAide)에 저장된 알림 데이터를
/// 표현합니다. 읽기 전용으로, 생성/편집 없이 조회만 지원합니다.
struct AppNotification: Identifiable, Codable, Equatable {
    let id: UUID
    let appName: String
    let sender: String
    let body: String
    let date: Date

    enum CodingKeys: String, CodingKey {
        case id
        case appName = "app_name"
        case sender
        case body
        case date
    }
}
