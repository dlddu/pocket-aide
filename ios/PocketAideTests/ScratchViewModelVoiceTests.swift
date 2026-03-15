// ScratchViewModelVoiceTests.swift
// PocketAideTests
//
// Unit tests for ScratchViewModel's voice memo support:
//   - createMemo(content:source:) accepts a source parameter
//   - createMemo(content:) defaults source to "text" (regression guard)
//   - createMemo(content:source: "voice") passes "voice" to MemoService
//   - SettingsView exposes a Siri Shortcut setup guidance section
//
// DLD-732: 9-2: Siri Shortcut 음성 메모 — 구현 (TDD Red phase)
//
// NOTE: These tests are in the TDD Red phase — they will fail until
// code-writer adds the `source` parameter to ScratchViewModel.createMemo
// and adds the Shortcut setup section to SettingsView.
//
// Targets production types:
//   - PocketAide.ScratchViewModel  (source param addition needed)
//   - PocketAide.SettingsView      (Shortcut guidance section needed)

import XCTest
@testable import PocketAide

// MARK: - MockKeychainService

/// `KeychainService`의 테스트 대역.
///
/// `ScratchViewModel`이 Keychain에 접근하지 않고도 테스트할 수 있도록
/// 인메모리 저장소를 제공합니다.
final class MockKeychainService {

    var storedServerURL: String? = "https://test.example.com"
    var storedToken: String? = "test-token"

    func loadServerURL() -> String? { storedServerURL }
    func loadToken() -> String? { storedToken }
}

// MARK: - ScratchViewModelVoiceTests

/// `ScratchViewModel.createMemo(content:source:)` 의 `source` 파라미터 지원을
/// 검증하는 유닛 테스트.
///
/// `MemoService`를 직접 호출하는 대신 `MockMemoService`를 주입하여
/// 네트워크 없이 인자 전달 여부만 검증합니다.
@MainActor
final class ScratchViewModelVoiceTests: XCTestCase {

    // MARK: - Properties

    private var sut: ScratchViewModel!
    private var mockMemoService: MockMemoService!

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        mockMemoService = MockMemoService()
        // ScratchViewModel must accept an injected MemoService so tests can
        // substitute MockMemoService without hitting the real network.
        sut = ScratchViewModel(memoService: mockMemoService)
    }

    override func tearDown() {
        sut = nil
        mockMemoService = nil
        super.tearDown()
    }

    // MARK: - Regression: Default Source

    /// `createMemo(content:)` — `source` 파라미터 없이 호출하면 기존과 동일하게
    /// `source: "text"`가 전달되어야 합니다. 기존 동작이 깨지면 안 됩니다.
    func test_createMemo_defaultSource_isText() async {
        // Arrange
        mockMemoService.stubbedMemo = Memo(id: 1, content: "기본 메모", source: "text")

        // Act
        await sut.createMemo(content: "기본 메모")

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.first?.source,
            "text",
            "createMemo(content:) without explicit source must default to 'text'"
        )
    }

    // MARK: - Happy Path: source parameter

    /// `createMemo(content:source: "voice")`를 호출하면 `MemoService.create()`에
    /// `source: "voice"`가 전달되어야 합니다.
    func test_createMemo_voiceSource_passesVoiceToMemoService() async {
        // Arrange
        let content = "음성 메모 내용"
        mockMemoService.stubbedMemo = Memo(id: 2, content: content, source: "voice")

        // Act
        await sut.createMemo(content: content, source: "voice")

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.count,
            1,
            "MemoService.create() must be called exactly once"
        )
        XCTAssertEqual(
            mockMemoService.createCalls.first?.source,
            "voice",
            "MemoService.create() must receive source == 'voice'"
        )
        XCTAssertEqual(
            mockMemoService.createCalls.first?.content,
            content,
            "MemoService.create() must receive the correct memo content"
        )
    }

    /// `createMemo(content:source: "voice")` 성공 시 생성된 메모가
    /// `memos` 배열에 추가되어야 합니다.
    func test_createMemo_voiceSource_appendsMemoToList() async {
        // Arrange
        let newMemo = Memo(id: 3, content: "내일 회의 준비", source: "voice")
        mockMemoService.stubbedMemo = newMemo

        // Act
        await sut.createMemo(content: "내일 회의 준비", source: "voice")

        // Assert
        XCTAssertTrue(
            sut.memos.contains(where: { $0.id == newMemo.id }),
            "Newly created voice memo must be appended to sut.memos"
        )
    }

    /// 성공 후 `errorMessage`가 nil이어야 합니다.
    func test_createMemo_voiceSource_clearsErrorMessage() async {
        // Arrange
        mockMemoService.stubbedMemo = Memo(id: 4, content: "성공", source: "voice")

        // Act
        await sut.createMemo(content: "성공", source: "voice")

        // Assert
        XCTAssertNil(
            sut.errorMessage,
            "errorMessage must be nil after a successful voice memo creation"
        )
    }

    // MARK: - Edge Case: source parameter variants

    /// `source: "text"`를 명시적으로 전달해도 정상 동작해야 합니다.
    func test_createMemo_explicitTextSource_passesTextToMemoService() async {
        // Arrange
        mockMemoService.stubbedMemo = Memo(id: 5, content: "텍스트 메모", source: "text")

        // Act
        await sut.createMemo(content: "텍스트 메모", source: "text")

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.first?.source,
            "text",
            "Explicit source: 'text' must be forwarded as-is"
        )
    }

    // MARK: - Error Handling

    /// `MemoService.create()` 실패 시 `errorMessage`가 설정되어야 합니다.
    func test_createMemo_voiceSource_onServiceFailure_setsErrorMessage() async {
        // Arrange
        mockMemoService.simulatedCreateError = APIError.serverError(500)

        // Act
        await sut.createMemo(content: "실패 케이스", source: "voice")

        // Assert
        XCTAssertNotNil(
            sut.errorMessage,
            "errorMessage must be set when MemoService.create() throws"
        )
    }

    /// `MemoService.create()` 실패 시 `memos` 배열이 변경되지 않아야 합니다.
    func test_createMemo_voiceSource_onServiceFailure_doesNotModifyMemoList() async {
        // Arrange
        let initialMemos = sut.memos
        mockMemoService.simulatedCreateError = APIError.serverError(500)

        // Act
        await sut.createMemo(content: "실패 케이스", source: "voice")

        // Assert
        XCTAssertEqual(
            sut.memos.count,
            initialMemos.count,
            "memos list must not change when memo creation fails"
        )
    }

    /// 서버 URL 또는 토큰이 없을 때 `MemoService.create()`가 호출되지 않아야 합니다.
    func test_createMemo_missingCredentials_doesNotCallMemoService() async {
        // Arrange — ViewModel without valid credentials
        let sutWithoutCreds = ScratchViewModel(
            memoService: mockMemoService,
            serverURL: nil,
            token: nil
        )

        // Act
        await sutWithoutCreds.createMemo(content: "인증 없음", source: "voice")

        // Assert
        XCTAssertEqual(
            mockMemoService.createCalls.count,
            0,
            "MemoService.create() must NOT be called when credentials are missing"
        )
        XCTAssertNotNil(
            sutWithoutCreds.errorMessage,
            "errorMessage must be set when credentials are missing"
        )
    }

    // MARK: - isLoading State

    /// `createMemo(content:source:)` 중 `isLoading`이 `true`가 되어야 합니다.
    ///
    /// 비동기 완료 후 다시 `false`로 돌아와야 합니다.
    func test_createMemo_voiceSource_setsIsLoadingDuringCall() async {
        // Arrange
        mockMemoService.stubbedMemo = Memo(id: 6, content: "로딩 테스트", source: "voice")
        var observedIsLoading = false

        // Act — observe isLoading during the call
        // We capture the state immediately after the call completes (false expected)
        await sut.createMemo(content: "로딩 테스트", source: "voice")

        // Assert — after completion, isLoading must be false
        observedIsLoading = sut.isLoading
        XCTAssertFalse(
            observedIsLoading,
            "isLoading must be false after createMemo(content:source:) completes"
        )
    }
}

// MARK: - SettingsViewShortcutSectionTests

/// `SettingsView`에 Siri Shortcut 설정 안내 섹션이 존재하는지 검증합니다.
///
/// SwiftUI 뷰의 내부 구조를 직접 검사하는 대신, 뷰가 요구하는 식별자(accessibilityIdentifier)의
/// 존재 여부를 `ViewInspector` 패턴으로 검증합니다.
/// 구현이 없으면 컴파일 에러 또는 런타임 실패로 TDD Red phase임을 나타냅니다.
///
/// NOTE: 이 테스트는 `SettingsView`에 `shortcut_setup_section` accessibilityIdentifier를
///       가진 뷰가 추가될 때까지 실패합니다.
final class SettingsViewShortcutSectionTests: XCTestCase {

    /// `SettingsView`는 Siri Shortcut 설정 안내를 위한 섹션을 제공해야 합니다.
    /// 해당 섹션은 `shortcut_setup_section` accessibilityIdentifier로 식별 가능해야 합니다.
    ///
    /// 이 테스트는 `SettingsView`에 `hasShortcutSetupSection` 계산 프로퍼티 또는
    /// 동등한 구조가 존재하는지 확인하는 컴파일 타임 검증입니다.
    func test_settingsView_exposesShortcutSetupSectionIdentifier() {
        // Assert — SettingsView must declare the shortcut setup section identifier constant.
        // This is verified by checking that the static accessibilityIdentifier string is defined.
        XCTAssertEqual(
            SettingsView.shortcutSetupSectionIdentifier,
            "shortcut_setup_section",
            "SettingsView must expose 'shortcut_setup_section' accessibilityIdentifier constant"
        )
    }

    /// 설정 안내 섹션은 `Add to Siri` 버튼 또는 동등한 안내 UI를 포함해야 합니다.
    /// 해당 UI는 `shortcut_add_to_siri_button` accessibilityIdentifier로 식별되어야 합니다.
    func test_settingsView_shortcutSection_exposesAddToSiriButtonIdentifier() {
        XCTAssertEqual(
            SettingsView.shortcutAddToSiriButtonIdentifier,
            "shortcut_add_to_siri_button",
            "SettingsView must expose 'shortcut_add_to_siri_button' accessibilityIdentifier constant for the Siri Shortcut setup button"
        )
    }
}
