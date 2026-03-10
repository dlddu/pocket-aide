// Tests for the TabBar / AppTab model layer.
//
// SwiftUI views cannot be instantiated on Linux CI, so these tests target the
// pure-Swift `AppTab` enum that drives the tab bar.  UI-layer tests that need
// SwiftUI are compiled only on Apple platforms via the canImport guard.
//
// Covered:
//   - All 7 tabs exist in `AppTab.allCases`.
//   - Each tab has a non-empty title and SF Symbol name.
//   - Each tab maps to the expected title string.
//   - CaseIterable order is stable (Chat first, Quotes last).

import XCTest
@testable import PocketAide

final class AppTabTests: XCTestCase {

    // MARK: - Completeness

    func test_allCases_containsSevenTabs() {
        XCTAssertEqual(AppTab.allCases.count, 7)
    }

    func test_allCases_containsChatTab() {
        XCTAssertTrue(AppTab.allCases.contains(.chat))
    }

    func test_allCases_containsRoutineTab() {
        XCTAssertTrue(AppTab.allCases.contains(.routine))
    }

    func test_allCases_containsPersonalTodoTab() {
        XCTAssertTrue(AppTab.allCases.contains(.personalTodo))
    }

    func test_allCases_containsWorkTodoTab() {
        XCTAssertTrue(AppTab.allCases.contains(.workTodo))
    }

    func test_allCases_containsScratchPadTab() {
        XCTAssertTrue(AppTab.allCases.contains(.scratchPad))
    }

    func test_allCases_containsNotificationsTab() {
        XCTAssertTrue(AppTab.allCases.contains(.notifications))
    }

    func test_allCases_containsQuotesTab() {
        XCTAssertTrue(AppTab.allCases.contains(.quotes))
    }

    // MARK: - Titles

    func test_chat_title_isChat() {
        XCTAssertEqual(AppTab.chat.title, "Chat")
    }

    func test_routine_title_isRoutine() {
        XCTAssertEqual(AppTab.routine.title, "Routine")
    }

    func test_personalTodo_title_isPersonal() {
        XCTAssertEqual(AppTab.personalTodo.title, "Personal")
    }

    func test_workTodo_title_isWork() {
        XCTAssertEqual(AppTab.workTodo.title, "Work")
    }

    func test_scratchPad_title_isScratch() {
        XCTAssertEqual(AppTab.scratchPad.title, "Scratch")
    }

    func test_notifications_title_isAlerts() {
        XCTAssertEqual(AppTab.notifications.title, "Alerts")
    }

    func test_quotes_title_isQuotes() {
        XCTAssertEqual(AppTab.quotes.title, "Quotes")
    }

    func test_allTabs_haveNonEmptyTitle() {
        for tab in AppTab.allCases {
            XCTAssertFalse(
                tab.title.isEmpty,
                "Tab \(tab) has an empty title"
            )
        }
    }

    // MARK: - Symbol names

    func test_allTabs_haveNonEmptySymbolName() {
        for tab in AppTab.allCases {
            XCTAssertFalse(
                tab.symbolName.isEmpty,
                "Tab \(tab) has an empty symbolName"
            )
        }
    }

    func test_chat_symbolName() {
        XCTAssertEqual(AppTab.chat.symbolName, "bubble.left.and.bubble.right")
    }

    func test_routine_symbolName() {
        XCTAssertEqual(AppTab.routine.symbolName, "repeat")
    }

    func test_personalTodo_symbolName() {
        XCTAssertEqual(AppTab.personalTodo.symbolName, "person.crop.circle")
    }

    func test_workTodo_symbolName() {
        XCTAssertEqual(AppTab.workTodo.symbolName, "briefcase")
    }

    func test_scratchPad_symbolName() {
        XCTAssertEqual(AppTab.scratchPad.symbolName, "pencil.and.scribble")
    }

    func test_notifications_symbolName() {
        XCTAssertEqual(AppTab.notifications.symbolName, "bell")
    }

    func test_quotes_symbolName() {
        XCTAssertEqual(AppTab.quotes.symbolName, "quote.bubble")
    }

    // MARK: - Ordering

    func test_firstTab_isChat() {
        XCTAssertEqual(AppTab.allCases.first, .chat)
    }

    func test_lastTab_isQuotes() {
        XCTAssertEqual(AppTab.allCases.last, .quotes)
    }

    func test_tabOrder_isStable() {
        // Arrange
        let expected: [AppTab] = [
            .chat, .routine, .personalTodo, .workTodo, .scratchPad, .notifications, .quotes,
        ]

        // Assert
        XCTAssertEqual(AppTab.allCases, expected)
    }

    // MARK: - Equatability

    func test_sameTab_isEqual() {
        XCTAssertEqual(AppTab.chat, AppTab.chat)
    }

    func test_differentTabs_areNotEqual() {
        XCTAssertNotEqual(AppTab.chat, AppTab.quotes)
    }

    // MARK: - Unique titles and symbols

    func test_allTitles_areUnique() {
        let titles = AppTab.allCases.map(\.title)
        let uniqueTitles = Set(titles)
        XCTAssertEqual(
            titles.count,
            uniqueTitles.count,
            "Duplicate titles found: \(titles)"
        )
    }

    func test_allSymbolNames_areUnique() {
        let symbols = AppTab.allCases.map(\.symbolName)
        let uniqueSymbols = Set(symbols)
        XCTAssertEqual(
            symbols.count,
            uniqueSymbols.count,
            "Duplicate symbol names found: \(symbols)"
        )
    }
}
