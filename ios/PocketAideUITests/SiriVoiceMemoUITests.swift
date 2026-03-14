// SiriVoiceMemoUITests.swift
// PocketAideUITests
//
// XCUITest suite that covers the end-to-end Siri Shortcut voice memo flow:
//   Siri Shortcut App Intent 트리거 시뮬레이션 → 음성 입력 인터페이스 표시
//   → MockSpeechRecognizer 음성 인식 → 텍스트 변환 확인
//   → 백엔드 API로 메모 저장 (source: "voice")
//   → 앱 진입 → 임시 공간(Scratch Space) 탭에서 해당 메모 존재 확인
//   → 메모 출처 아이콘(mic.fill) 올바르게 표시 확인
//
// DLD-731: 9-1: Siri Shortcut 음성 메모 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (XCTSkip). Activate after DLD-732:
//   - App Intent (VoiceMemoIntent) 구현 및 Siri Shortcut 등록
//   - "siri_shortcut_button" accessibilityIdentifier 노출 (UI 테스트용 트리거)
//   - "voice_memo_recording_indicator" accessibilityIdentifier 노출
//   - "voice_memo_save_confirmation" accessibilityIdentifier 노출
//   - "memo_source_icon_<content>" accessibilityIdentifier 노출 (source == "voice"이면 mic.fill)
//   - "--uitesting" launch argument가 MockSpeechRecognizer 주입 + App Intent 시뮬레이션 활성화

import XCTest

final class SiriVoiceMemoUITests: XCTestCase {

    // MARK: - Properties

    private var app: XCUIApplication!

    // MARK: - Lifecycle

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()
        // "--uitesting" bypasses the auth flow and lands on MainTabView,
        // and additionally injects MockSpeechRecognizer so that voice recognition
        // can be driven deterministically, consistent with VoiceChatUITests.
        app.launchArguments = ["--uitesting"]
        app.launch()
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Happy Path: App Intent Trigger → Voice Input Interface

    /// Siri Shortcut App Intent 트리거 시뮬레이션 시, 음성 입력 인터페이스가
    /// 화면에 표시되어야 한다.
    ///
    /// Expected flow:
    ///   앱 실행 → "siri_shortcut_button" 탭 (App Intent 시뮬레이션)
    ///   → "voice_memo_recording_indicator" 나타남
    func test_siriShortcut_trigger_displaysVoiceInputInterface() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5), "TabBar must be visible before triggering App Intent")

        // Act — simulate Siri Shortcut App Intent trigger via UI test button
        // "--uitesting" exposes "siri_shortcut_button" which fires VoiceMemoIntent
        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button (siri_shortcut_button) must be available in --uitesting mode"
        )
        siriShortcutButton.tap()

        // Assert — voice recording indicator must appear
        let recordingIndicator = app.otherElements["voice_memo_recording_indicator"]
        XCTAssertTrue(
            recordingIndicator.waitForExistence(timeout: 5),
            "Voice memo recording indicator (voice_memo_recording_indicator) should appear after App Intent trigger"
        )
    }

    // MARK: - Happy Path: MockSpeechRecognizer Transcription

    /// MockSpeechRecognizer가 음성 입력을 텍스트로 변환한 결과가
    /// 저장 확인 표시에 반영되어야 한다.
    ///
    /// Expected flow:
    ///   "siri_shortcut_button" 탭 → "voice_memo_recording_indicator" 나타남
    ///   → MockSpeechRecognizer가 simulatedTranscript 방출
    ///   → "voice_memo_save_confirmation" 나타남 (변환된 텍스트 포함)
    func test_siriShortcut_mockSpeechRecognizer_transcribesVoiceToText() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button must be available in --uitesting mode"
        )

        // Act — trigger App Intent; MockSpeechRecognizer will emit simulatedTranscript
        siriShortcutButton.tap()

        let recordingIndicator = app.otherElements["voice_memo_recording_indicator"]
        XCTAssertTrue(
            recordingIndicator.waitForExistence(timeout: 5),
            "Recording indicator must appear to confirm voice input is active"
        )

        // Assert — save confirmation must appear once transcription completes
        // MockSpeechRecognizer emits simulatedTranscript immediately after startRecording()
        let saveConfirmation = app.otherElements["voice_memo_save_confirmation"]
        XCTAssertTrue(
            saveConfirmation.waitForExistence(timeout: 5),
            "Voice memo save confirmation (voice_memo_save_confirmation) should appear after transcription completes"
        )
        // Transcribed text must be non-empty in the confirmation view
        XCTAssertFalse(
            saveConfirmation.label.isEmpty,
            "Save confirmation must contain the transcribed text"
        )
    }

    // MARK: - Happy Path: Voice Memo Saved with source "voice"

    /// 음성 메모가 백엔드 API (POST /memos, source: "voice")로 저장되고,
    /// 저장 완료 후 앱 UI에 반영되어야 한다.
    ///
    /// Expected flow:
    ///   "siri_shortcut_button" 탭 → MockSpeechRecognizer 전사 완료
    ///   → "voice_memo_save_confirmation" 나타남
    ///   → 저장 완료 후 "voice_memo_recording_indicator" 사라짐
    func test_siriShortcut_voiceMemo_savedToBackend() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button must be available"
        )

        // Act — trigger voice memo recording via App Intent
        siriShortcutButton.tap()

        // Wait for transcription and save confirmation
        let saveConfirmation = app.otherElements["voice_memo_save_confirmation"]
        XCTAssertTrue(
            saveConfirmation.waitForExistence(timeout: 5),
            "Save confirmation must appear after transcription"
        )

        // Assert — recording indicator must disappear once memo is saved to backend
        let recordingIndicator = app.otherElements["voice_memo_recording_indicator"]
        XCTAssertFalse(
            recordingIndicator.waitForExistence(timeout: 5),
            "Recording indicator (voice_memo_recording_indicator) should disappear after memo is saved"
        )
    }

    // MARK: - Happy Path: Memo Appears in Scratch Tab

    /// 음성 메모 저장 후 앱의 임시 공간(Scratch Space) 탭에서
    /// 해당 메모가 목록에 존재해야 한다.
    ///
    /// Expected flow:
    ///   "siri_shortcut_button" 탭 → MockSpeechRecognizer "내일 회의 준비하기" 전사
    ///   → 메모 저장 완료 → "tab_scratch" 탭 이동
    ///   → "scratch_list_view" 표시 → "memo_row_내일 회의 준비하기" 존재
    func test_siriShortcut_savedVoiceMemo_appearsInScratchTab() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        // MockSpeechRecognizer는 "--uitesting" 모드에서 simulatedTranscript = "내일 회의 준비하기" 방출
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button must be available in --uitesting mode"
        )

        // Act — trigger App Intent and wait for save confirmation
        siriShortcutButton.tap()

        let saveConfirmation = app.otherElements["voice_memo_save_confirmation"]
        XCTAssertTrue(
            saveConfirmation.waitForExistence(timeout: 5),
            "Save confirmation must appear before navigating to Scratch tab"
        )

        // Navigate to Scratch tab
        tabBar.buttons["tab_scratch"].tap()

        let scratchListView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(
            scratchListView.waitForExistence(timeout: 5),
            "Scratch list view (scratch_list_view) must appear after tapping Scratch tab"
        )

        // Assert — voice memo must be present in the scratch list
        let memoRow = app.otherElements["memo_row_내일 회의 준비하기"]
        XCTAssertTrue(
            memoRow.waitForExistence(timeout: 5),
            "Voice memo '내일 회의 준비하기' (memo_row_내일 회의 준비하기) should appear in the scratch list after Siri Shortcut save"
        )
    }

    // MARK: - Happy Path: Voice Source Icon (mic.fill) Displayed

    /// 음성 메모 행에 mic.fill 출처 아이콘이 올바르게 표시되어야 한다.
    /// (ScratchScreen.swift: source == "voice" → mic.fill 아이콘)
    ///
    /// Expected flow:
    ///   음성 메모 저장 완료 → "tab_scratch" 탭 이동
    ///   → "memo_row_내일 회의 준비하기" 행에
    ///     "memo_source_icon_내일 회의 준비하기" 아이콘 존재
    func test_siriShortcut_voiceMemoRow_displaysMicFillSourceIcon() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button must be available in --uitesting mode"
        )

        // Trigger voice memo and wait for save
        siriShortcutButton.tap()

        let saveConfirmation = app.otherElements["voice_memo_save_confirmation"]
        XCTAssertTrue(
            saveConfirmation.waitForExistence(timeout: 5),
            "Save confirmation must appear before navigating to Scratch tab"
        )

        // Navigate to Scratch tab
        tabBar.buttons["tab_scratch"].tap()

        let scratchListView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(
            scratchListView.waitForExistence(timeout: 5),
            "Scratch list view must appear after tapping Scratch tab"
        )

        let memoRow = app.otherElements["memo_row_내일 회의 준비하기"]
        XCTAssertTrue(
            memoRow.waitForExistence(timeout: 5),
            "Voice memo row must exist before checking source icon"
        )

        // Act & Assert — mic.fill source icon must be present on the voice memo row
        let sourceIcon = memoRow.images["memo_source_icon_내일 회의 준비하기"]
        XCTAssertTrue(
            sourceIcon.waitForExistence(timeout: 5),
            "Voice source icon (memo_source_icon_내일 회의 준비하기) should be displayed on the memo row when source == 'voice'"
        )
    }

    // MARK: - Edge Case: Multiple Partial Transcriptions

    /// MockSpeechRecognizer의 simulatedPhrases를 통한 부분 전사 흐름이
    /// 최종적으로 완전한 메모로 합쳐져야 한다.
    ///
    /// Expected flow:
    ///   "siri_shortcut_button" 탭 → simulatedPhrases = ["장보기", "우유", "계란"]
    ///   → 부분 전사들이 순차적으로 조합
    ///   → "voice_memo_save_confirmation" 에 최종 전사 텍스트 표시
    func test_siriShortcut_partialTranscriptions_combinedIntoFinalMemo() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button must be available"
        )

        // Act — MockSpeechRecognizer will emit simulatedPhrases in sequence
        // The phrases are configured by "--uitesting" launch argument
        siriShortcutButton.tap()

        // Assert — save confirmation must appear with non-empty transcribed text
        let saveConfirmation = app.otherElements["voice_memo_save_confirmation"]
        XCTAssertTrue(
            saveConfirmation.waitForExistence(timeout: 5),
            "Save confirmation should appear after partial transcriptions are combined"
        )
        XCTAssertFalse(
            saveConfirmation.label.isEmpty,
            "Final transcription result must not be empty after partial phrases are combined"
        )
    }

    // MARK: - Edge Case: Empty Transcription Not Saved

    /// 음성 인식 결과가 빈 문자열인 경우, 메모가 저장되지 않아야 한다.
    ///
    /// Expected flow:
    ///   "siri_shortcut_button" 탭 → MockSpeechRecognizer가 빈 transcript 방출
    ///   → "voice_memo_save_confirmation" 나타나지 않음
    ///   → Scratch 탭에 빈 메모 행이 추가되지 않음
    func test_siriShortcut_emptyTranscription_doesNotSaveMemo() throws {
        throw XCTSkip("Skipped: requires Siri Shortcut App Intent implementation (DLD-732)")

        // Arrange
        // This test requires "--uitesting-empty-transcript" or equivalent launch arg
        // to instruct MockSpeechRecognizer to return an empty string.
        // The flag is expected to be supported alongside "--uitesting".
        app.launchArguments = ["--uitesting", "--uitesting-empty-transcript"]
        app.launch()

        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.waitForExistence(timeout: 5))

        let siriShortcutButton = app.buttons["siri_shortcut_button"]
        XCTAssertTrue(
            siriShortcutButton.waitForExistence(timeout: 5),
            "Siri Shortcut trigger button must be available in --uitesting mode"
        )

        // Act — trigger with empty transcript configured
        siriShortcutButton.tap()

        // Assert — save confirmation must NOT appear for empty transcription
        let saveConfirmation = app.otherElements["voice_memo_save_confirmation"]
        XCTAssertFalse(
            saveConfirmation.waitForExistence(timeout: 3),
            "Save confirmation should not appear when transcription result is empty"
        )

        // Navigate to scratch tab and verify no empty memo row was added
        tabBar.buttons["tab_scratch"].tap()

        let scratchListView = app.otherElements["scratch_list_view"]
        XCTAssertTrue(
            scratchListView.waitForExistence(timeout: 5),
            "Scratch list view must appear after tapping Scratch tab"
        )

        // An empty memo row identifier would be "memo_row_" with no trailing content
        let emptyMemoRow = app.otherElements["memo_row_"]
        XCTAssertFalse(
            emptyMemoRow.exists,
            "An empty memo row should not exist in the scratch list after an empty transcription"
        )
    }
}
