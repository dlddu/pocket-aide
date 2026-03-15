// NotificationRepository.swift
// PocketAide

import Foundation

/// App Group UserDefaults에서 알림 데이터를 읽어 앱별로 그룹핑하는 레포지토리.
///
/// - App Group 식별자: `group.com.dlddu.PocketAide`
/// - UserDefaults 키: `"notifications"`
/// - 저장 형식: `[AppNotification]`을 JSON 인코딩한 Data
///
/// 읽기 전용입니다. 알림 생성/편집은 지원하지 않습니다.
final class NotificationRepository {

    // MARK: - Constants

    static let appGroupIdentifier = "group.com.dlddu.PocketAide"
    static let userDefaultsKey = "notifications"

    // MARK: - Properties

    private let userDefaults: UserDefaults
    private let decoder: JSONDecoder

    // MARK: - Initialisation

    /// - Parameter userDefaults: 주입 가능한 UserDefaults 인스턴스.
    ///   기본값은 App Group UserDefaults입니다.
    ///   테스트 시 임시 suite로 대체할 수 있습니다.
    init(userDefaults: UserDefaults? = nil) {
        self.userDefaults = userDefaults
            ?? UserDefaults(suiteName: NotificationRepository.appGroupIdentifier)
            ?? .standard
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        self.decoder = decoder
    }

    // MARK: - Public Interface

    /// App Group UserDefaults에서 모든 알림을 읽어 반환합니다.
    ///
    /// - Returns: 저장된 `AppNotification` 배열. 데이터가 없거나 디코딩에
    ///   실패하면 빈 배열을 반환합니다.
    func fetchAll() -> [AppNotification] {
        guard let data = userDefaults.data(forKey: NotificationRepository.userDefaultsKey) else {
            return []
        }
        return (try? decoder.decode([AppNotification].self, from: data)) ?? []
    }

    /// 알림을 `appName`을 기준으로 그룹핑하여 딕셔너리로 반환합니다.
    ///
    /// - Returns: `[appName: [AppNotification]]` 형태의 딕셔너리.
    ///   각 배열은 `date` 기준 내림차순으로 정렬됩니다.
    func fetchGroupedByApp() -> [String: [AppNotification]] {
        let all = fetchAll()
        var grouped: [String: [AppNotification]] = [:]
        for notification in all {
            grouped[notification.appName, default: []].append(notification)
        }
        for key in grouped.keys {
            grouped[key] = grouped[key]?.sorted { $0.date > $1.date }
        }
        return grouped
    }
}

// MARK: - Emoji Mapping

extension NotificationRepository {

    /// 앱 이름에 대응하는 이모지를 반환합니다.
    ///
    /// 알려지지 않은 앱은 기본 이모지(`"🔔"`)를 반환합니다.
    static func emoji(for appName: String) -> String {
        switch appName {
        case "카카오톡":  return "💬"
        case "Slack":    return "💼"
        case "문자":      return "✉️"
        case "메일":      return "📧"
        case "Instagram": return "📸"
        case "YouTube":  return "▶️"
        default:         return "🔔"
        }
    }
}

// MARK: - URL Scheme Mapping

extension NotificationRepository {

    /// 앱 이름에 대응하는 URL Scheme 문자열을 반환합니다.
    ///
    /// 지원하지 않는 앱은 `nil`을 반환합니다.
    static func urlScheme(for appName: String) -> String? {
        switch appName {
        case "카카오톡":  return "kakaotalk://"
        case "Slack":    return "slack://"
        case "문자":      return "sms://"
        case "메일":      return "message://"
        case "Instagram": return "instagram://"
        case "YouTube":  return "youtube://"
        default:         return nil
        }
    }
}
