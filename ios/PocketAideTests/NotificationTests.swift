// NotificationTests.swift
// PocketAideTests
//
// Unit tests for the Notifications (알림 모음) feature:
//   - AppNotification model Codable encoding / decoding
//   - NotificationRepository.fetchAll() reads from injected UserDefaults
//   - NotificationRepository.fetchGroupedByApp() grouping and sort order
//   - Empty-data handling (missing key, empty array, malformed JSON)
//   - NotificationRepository.emoji(for:) app-name → emoji mapping
//   - NotificationRepository.urlScheme(for:) app-name → URL scheme mapping
//
// DLD-736: 11-2: 알림 모음 — 단위 테스트 작성 (TDD Red phase)
//
// NOTE: These tests are in the TDD Red phase — they will fail until
// code-writer implements AppNotification and NotificationRepository.
//
// Targets production types:
//   - PocketAide.AppNotification          (to be created)
//   - PocketAide.NotificationRepository   (to be created)

import XCTest
@testable import PocketAide

// MARK: - Helpers

/// ISO 8601 문자열을 Date로 변환합니다. 테스트 픽스처 생성 전용입니다.
private func iso8601Date(_ string: String) -> Date {
    let formatter = ISO8601DateFormatter()
    return formatter.date(from: string)!
}

/// 테스트용 AppNotification JSON 배열 Data를 만듭니다.
private func makeNotificationsData(_ notifications: [AppNotification]) -> Data {
    let encoder = JSONEncoder()
    encoder.dateEncodingStrategy = .iso8601
    return try! encoder.encode(notifications)
}

/// 테스트 격리를 위한 인메모리 UserDefaults suite를 만듭니다.
private func makeIsolatedDefaults(suiteName: String = UUID().uuidString) -> UserDefaults {
    let defaults = UserDefaults(suiteName: suiteName)!
    defaults.removePersistentDomain(forName: suiteName)
    return defaults
}

// MARK: - AppNotificationCodableTests

/// `AppNotification`의 `Codable` 인코딩/디코딩을 검증합니다.
final class AppNotificationCodableTests: XCTestCase {

    // MARK: - Happy Path: Encode → Decode round-trip

    /// 모든 필드가 정상 값인 알림을 인코딩 후 디코딩하면 원본과 동일해야 합니다.
    func test_codable_roundTrip_preservesAllFields() throws {
        // Arrange
        let originalDate = iso8601Date("2026-03-15T09:00:00Z")
        let original = AppNotification(
            id: UUID(uuidString: "12345678-1234-1234-1234-123456789012")!,
            appName: "카카오톡",
            sender: "홍길동",
            body: "안녕하세요!",
            date: originalDate
        )
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        // Act
        let data = try encoder.encode(original)
        let decoded = try decoder.decode(AppNotification.self, from: data)

        // Assert
        XCTAssertEqual(decoded.id, original.id)
        XCTAssertEqual(decoded.appName, original.appName)
        XCTAssertEqual(decoded.sender, original.sender)
        XCTAssertEqual(decoded.body, original.body)
        XCTAssertEqual(decoded.date.timeIntervalSinceReferenceDate,
                       original.date.timeIntervalSinceReferenceDate,
                       accuracy: 1.0,
                       "date must survive ISO 8601 round-trip within 1 second")
    }

    /// JSON에서 snake_case 키(`app_name`)를 올바르게 디코딩해야 합니다.
    func test_decode_snakeCaseKey_appName() throws {
        // Arrange
        let json = """
        {
            "id": "AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA",
            "app_name": "Slack",
            "sender": "채널",
            "body": "새 메시지",
            "date": "2026-03-15T08:30:00Z"
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        // Act
        let notification = try decoder.decode(AppNotification.self, from: json)

        // Assert
        XCTAssertEqual(notification.appName, "Slack")
    }

    /// `id` 필드가 유효한 UUID 문자열일 때 올바르게 디코딩해야 합니다.
    func test_decode_uuidId_isPreserved() throws {
        // Arrange
        let expectedID = UUID(uuidString: "BBBBBBBB-BBBB-BBBB-BBBB-BBBBBBBBBBBB")!
        let json = """
        {
            "id": "BBBBBBBB-BBBB-BBBB-BBBB-BBBBBBBBBBBB",
            "app_name": "문자",
            "sender": "010-1234-5678",
            "body": "인증번호: 123456",
            "date": "2026-03-15T07:00:00Z"
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        // Act
        let notification = try decoder.decode(AppNotification.self, from: json)

        // Assert
        XCTAssertEqual(notification.id, expectedID)
    }

    /// `body`가 빈 문자열인 알림도 정상적으로 디코딩해야 합니다.
    func test_decode_emptyBody_isAllowed() throws {
        // Arrange
        let json = """
        {
            "id": "CCCCCCCC-CCCC-CCCC-CCCC-CCCCCCCCCCCC",
            "app_name": "메일",
            "sender": "no-reply@example.com",
            "body": "",
            "date": "2026-03-15T06:00:00Z"
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        // Act
        let notification = try decoder.decode(AppNotification.self, from: json)

        // Assert
        XCTAssertEqual(notification.body, "")
    }

    // MARK: - Error Case: Missing Required Field

    /// 필수 필드(`app_name`)가 누락된 JSON을 디코딩하면 `DecodingError`를 던져야 합니다.
    func test_decode_missingAppName_throwsDecodingError() {
        // Arrange
        let json = """
        {
            "id": "DDDDDDDD-DDDD-DDDD-DDDD-DDDDDDDDDDDD",
            "sender": "누군가",
            "body": "내용",
            "date": "2026-03-15T05:00:00Z"
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        // Act & Assert
        XCTAssertThrowsError(
            try decoder.decode(AppNotification.self, from: json),
            "Missing app_name must throw DecodingError"
        )
    }

    /// 배열 형태의 알림 JSON도 올바르게 디코딩해야 합니다.
    func test_decode_array_decodesMultipleNotifications() throws {
        // Arrange
        let json = """
        [
            {
                "id": "EEEEEEEE-EEEE-EEEE-EEEE-EEEEEEEEEEEE",
                "app_name": "카카오톡",
                "sender": "홍길동",
                "body": "안녕",
                "date": "2026-03-15T10:00:00Z"
            },
            {
                "id": "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF",
                "app_name": "Slack",
                "sender": "#general",
                "body": "안녕하세요",
                "date": "2026-03-15T10:01:00Z"
            }
        ]
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        // Act
        let notifications = try decoder.decode([AppNotification].self, from: json)

        // Assert
        XCTAssertEqual(notifications.count, 2)
        XCTAssertEqual(notifications[0].appName, "카카오톡")
        XCTAssertEqual(notifications[1].appName, "Slack")
    }
}

// MARK: - NotificationRepositoryFetchTests

/// `NotificationRepository.fetchAll()` 과 `fetchGroupedByApp()`의
/// 데이터 읽기 및 그룹핑 로직을 검증합니다.
final class NotificationRepositoryFetchTests: XCTestCase {

    // MARK: - Properties

    private var sut: NotificationRepository!
    private var defaults: UserDefaults!

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        defaults = makeIsolatedDefaults()
        sut = NotificationRepository(userDefaults: defaults)
    }

    override func tearDown() {
        sut = nil
        defaults = nil
        super.tearDown()
    }

    // MARK: - Happy Path: fetchAll

    /// UserDefaults에 저장된 알림 배열을 모두 반환해야 합니다.
    func test_fetchAll_returnsStoredNotifications() {
        // Arrange
        let notifications = [
            AppNotification(
                id: UUID(),
                appName: "카카오톡",
                sender: "홍길동",
                body: "점심 먹었어?",
                date: iso8601Date("2026-03-15T12:00:00Z")
            ),
            AppNotification(
                id: UUID(),
                appName: "Slack",
                sender: "#dev",
                body: "PR 리뷰 부탁드려요",
                date: iso8601Date("2026-03-15T11:00:00Z")
            )
        ]
        defaults.set(makeNotificationsData(notifications),
                     forKey: NotificationRepository.userDefaultsKey)

        // Act
        let result = sut.fetchAll()

        // Assert
        XCTAssertEqual(result.count, 2)
        XCTAssertEqual(result[0].appName, "카카오톡")
        XCTAssertEqual(result[1].appName, "Slack")
    }

    /// 저장된 알림의 모든 필드가 정확히 복원되어야 합니다.
    func test_fetchAll_preservesAllFieldValues() {
        // Arrange
        let expectedDate = iso8601Date("2026-03-15T09:30:00Z")
        let expectedID = UUID()
        let notification = AppNotification(
            id: expectedID,
            appName: "메일",
            sender: "boss@example.com",
            body: "보고서 제출 부탁드립니다",
            date: expectedDate
        )
        defaults.set(makeNotificationsData([notification]),
                     forKey: NotificationRepository.userDefaultsKey)

        // Act
        let result = sut.fetchAll()

        // Assert
        XCTAssertEqual(result.count, 1)
        let fetched = result[0]
        XCTAssertEqual(fetched.id, expectedID)
        XCTAssertEqual(fetched.appName, "메일")
        XCTAssertEqual(fetched.sender, "boss@example.com")
        XCTAssertEqual(fetched.body, "보고서 제출 부탁드립니다")
        XCTAssertEqual(fetched.date.timeIntervalSinceReferenceDate,
                       expectedDate.timeIntervalSinceReferenceDate,
                       accuracy: 1.0)
    }

    // MARK: - Edge Case: Empty Data

    /// UserDefaults에 해당 키가 없으면 빈 배열을 반환해야 합니다.
    func test_fetchAll_noKeyPresent_returnsEmptyArray() {
        // Arrange — defaults에 아무 데이터도 저장하지 않음

        // Act
        let result = sut.fetchAll()

        // Assert
        XCTAssertEqual(result, [])
    }

    /// 빈 배열 `[]`이 저장된 경우 빈 배열을 반환해야 합니다.
    func test_fetchAll_emptyArrayStored_returnsEmptyArray() {
        // Arrange
        let emptyData = makeNotificationsData([])
        defaults.set(emptyData, forKey: NotificationRepository.userDefaultsKey)

        // Act
        let result = sut.fetchAll()

        // Assert
        XCTAssertEqual(result, [])
    }

    /// 손상된(malformed) JSON이 저장된 경우 빈 배열을 반환해야 하며 크래시가 발생하면 안 됩니다.
    func test_fetchAll_malformedJSON_returnsEmptyArray() {
        // Arrange
        let corrupt = "this is not valid json".data(using: .utf8)!
        defaults.set(corrupt, forKey: NotificationRepository.userDefaultsKey)

        // Act
        let result = sut.fetchAll()

        // Assert
        XCTAssertEqual(result, [])
    }

    // MARK: - Happy Path: fetchGroupedByApp

    /// `fetchGroupedByApp()`은 알림을 appName 키로 그룹핑해야 합니다.
    func test_fetchGroupedByApp_groupsByAppName() {
        // Arrange
        let notifications = [
            AppNotification(id: UUID(), appName: "카카오톡", sender: "A",
                            body: "msg1", date: iso8601Date("2026-03-15T10:00:00Z")),
            AppNotification(id: UUID(), appName: "카카오톡", sender: "B",
                            body: "msg2", date: iso8601Date("2026-03-15T09:00:00Z")),
            AppNotification(id: UUID(), appName: "Slack",   sender: "#ch",
                            body: "msg3", date: iso8601Date("2026-03-15T08:00:00Z")),
        ]
        defaults.set(makeNotificationsData(notifications),
                     forKey: NotificationRepository.userDefaultsKey)

        // Act
        let grouped = sut.fetchGroupedByApp()

        // Assert
        XCTAssertEqual(grouped["카카오톡"]?.count, 2,
                       "카카오톡 그룹은 2개 알림을 포함해야 합니다")
        XCTAssertEqual(grouped["Slack"]?.count, 1,
                       "Slack 그룹은 1개 알림을 포함해야 합니다")
        XCTAssertNil(grouped["없는앱"],
                     "존재하지 않는 앱 키는 nil이어야 합니다")
    }

    /// 같은 앱 그룹 내 알림은 date 내림차순으로 정렬되어야 합니다.
    func test_fetchGroupedByApp_sortsByDateDescendingWithinGroup() {
        // Arrange
        let older = AppNotification(id: UUID(), appName: "카카오톡", sender: "A",
                                    body: "오래된 메시지",
                                    date: iso8601Date("2026-03-15T08:00:00Z"))
        let newer = AppNotification(id: UUID(), appName: "카카오톡", sender: "B",
                                    body: "최신 메시지",
                                    date: iso8601Date("2026-03-15T10:00:00Z"))
        // 의도적으로 오래된 것을 먼저 삽입
        defaults.set(makeNotificationsData([older, newer]),
                     forKey: NotificationRepository.userDefaultsKey)

        // Act
        let grouped = sut.fetchGroupedByApp()

        // Assert
        let kakaoGroup = try! XCTUnwrap(grouped["카카오톡"])
        XCTAssertEqual(kakaoGroup.count, 2)
        XCTAssertEqual(kakaoGroup[0].body, "최신 메시지",
                       "첫 번째 항목은 가장 최신 알림이어야 합니다")
        XCTAssertEqual(kakaoGroup[1].body, "오래된 메시지",
                       "두 번째 항목은 오래된 알림이어야 합니다")
    }

    /// 데이터가 없으면 빈 딕셔너리를 반환해야 합니다.
    func test_fetchGroupedByApp_noData_returnsEmptyDictionary() {
        // Arrange — 아무 데이터도 없음

        // Act
        let grouped = sut.fetchGroupedByApp()

        // Assert
        XCTAssertTrue(grouped.isEmpty,
                      "데이터가 없으면 빈 딕셔너리를 반환해야 합니다")
    }

    /// 한 앱의 알림만 있어도 딕셔너리에 해당 키 하나만 생성되어야 합니다.
    func test_fetchGroupedByApp_singleApp_createsSingleKey() {
        // Arrange
        let notifications = [
            AppNotification(id: UUID(), appName: "문자", sender: "010-9999-0000",
                            body: "배송 완료", date: iso8601Date("2026-03-15T14:00:00Z")),
            AppNotification(id: UUID(), appName: "문자", sender: "010-1111-2222",
                            body: "인증번호 456789", date: iso8601Date("2026-03-15T13:00:00Z")),
        ]
        defaults.set(makeNotificationsData(notifications),
                     forKey: NotificationRepository.userDefaultsKey)

        // Act
        let grouped = sut.fetchGroupedByApp()

        // Assert
        XCTAssertEqual(grouped.keys.count, 1, "키는 하나여야 합니다")
        XCTAssertEqual(grouped["문자"]?.count, 2)
    }

    /// 네 개의 서로 다른 앱 알림이 있으면 딕셔너리에 네 개의 키가 생성되어야 합니다.
    func test_fetchGroupedByApp_fourApps_createsFourKeys() {
        // Arrange
        let notifications = [
            AppNotification(id: UUID(), appName: "카카오톡", sender: "홍길동",
                            body: "ㅎㅇ", date: iso8601Date("2026-03-15T10:00:00Z")),
            AppNotification(id: UUID(), appName: "Slack", sender: "#ch",
                            body: "배포 완료", date: iso8601Date("2026-03-15T10:01:00Z")),
            AppNotification(id: UUID(), appName: "문자", sender: "01012345678",
                            body: "인증번호", date: iso8601Date("2026-03-15T10:02:00Z")),
            AppNotification(id: UUID(), appName: "메일", sender: "hr@company.com",
                            body: "급여명세서", date: iso8601Date("2026-03-15T10:03:00Z")),
        ]
        defaults.set(makeNotificationsData(notifications),
                     forKey: NotificationRepository.userDefaultsKey)

        // Act
        let grouped = sut.fetchGroupedByApp()

        // Assert
        XCTAssertEqual(grouped.keys.count, 4)
        XCTAssertNotNil(grouped["카카오톡"])
        XCTAssertNotNil(grouped["Slack"])
        XCTAssertNotNil(grouped["문자"])
        XCTAssertNotNil(grouped["메일"])
    }
}

// MARK: - NotificationRepositoryEmojiTests

/// `NotificationRepository.emoji(for:)` 앱 이름 → 이모지 매핑을 검증합니다.
final class NotificationRepositoryEmojiTests: XCTestCase {

    // MARK: - Known Apps

    func test_emoji_kakaoTalk_returnsChatBubble() {
        XCTAssertEqual(NotificationRepository.emoji(for: "카카오톡"), "💬")
    }

    func test_emoji_slack_returnsBriefcase() {
        XCTAssertEqual(NotificationRepository.emoji(for: "Slack"), "💼")
    }

    func test_emoji_sms_returnsEnvelope() {
        XCTAssertEqual(NotificationRepository.emoji(for: "문자"), "✉️")
    }

    func test_emoji_mail_returnsEmailIcon() {
        XCTAssertEqual(NotificationRepository.emoji(for: "메일"), "📧")
    }

    func test_emoji_instagram_returnsCamera() {
        XCTAssertEqual(NotificationRepository.emoji(for: "Instagram"), "📸")
    }

    func test_emoji_youtube_returnsPlayButton() {
        XCTAssertEqual(NotificationRepository.emoji(for: "YouTube"), "▶️")
    }

    // MARK: - Unknown App

    /// 알려지지 않은 앱 이름은 기본 벨 이모지를 반환해야 합니다.
    func test_emoji_unknownApp_returnsDefaultBell() {
        XCTAssertEqual(NotificationRepository.emoji(for: "UnknownApp"), "🔔")
    }

    /// 빈 문자열도 기본 이모지를 반환해야 합니다.
    func test_emoji_emptyString_returnsDefaultBell() {
        XCTAssertEqual(NotificationRepository.emoji(for: ""), "🔔")
    }
}

// MARK: - NotificationRepositoryURLSchemeTests

/// `NotificationRepository.urlScheme(for:)` 앱 이름 → URL Scheme 매핑을 검증합니다.
final class NotificationRepositoryURLSchemeTests: XCTestCase {

    // MARK: - Known Apps

    func test_urlScheme_kakaoTalk_returnsKakaoScheme() {
        XCTAssertEqual(NotificationRepository.urlScheme(for: "카카오톡"), "kakaotalk://")
    }

    func test_urlScheme_slack_returnsSlackScheme() {
        XCTAssertEqual(NotificationRepository.urlScheme(for: "Slack"), "slack://")
    }

    func test_urlScheme_sms_returnsSmsScheme() {
        XCTAssertEqual(NotificationRepository.urlScheme(for: "문자"), "sms://")
    }

    func test_urlScheme_mail_returnsMessageScheme() {
        XCTAssertEqual(NotificationRepository.urlScheme(for: "메일"), "message://")
    }

    func test_urlScheme_instagram_returnsInstagramScheme() {
        XCTAssertEqual(NotificationRepository.urlScheme(for: "Instagram"), "instagram://")
    }

    func test_urlScheme_youtube_returnsYouTubeScheme() {
        XCTAssertEqual(NotificationRepository.urlScheme(for: "YouTube"), "youtube://")
    }

    // MARK: - Unknown App

    /// 지원하지 않는 앱은 `nil`을 반환해야 합니다.
    func test_urlScheme_unknownApp_returnsNil() {
        XCTAssertNil(NotificationRepository.urlScheme(for: "UnknownApp"))
    }

    /// 빈 문자열도 `nil`을 반환해야 합니다.
    func test_urlScheme_emptyString_returnsNil() {
        XCTAssertNil(NotificationRepository.urlScheme(for: ""))
    }
}
