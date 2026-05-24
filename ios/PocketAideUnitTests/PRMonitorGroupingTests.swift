import XCTest
@testable import PocketAideAPI

final class PRMonitorGroupingTests: XCTestCase {
    private func item(
        id: Int64,
        repo: String = "dlddu/pocket-aide",
        prNumber: Int? = nil,
        headSHA: String = "",
        conclusion: String = "success",
        acked: Int64? = nil,
        createdAt: Int64 = 0
    ) -> NotificationHistoryItem {
        NotificationHistoryItem(
            id: id,
            repoFullName: repo,
            prNumber: prNumber,
            prTitle: prNumber.map { "PR #\($0)" },
            prURL: nil,
            commitURL: nil,
            runURL: nil,
            workflowName: "CI",
            headBranch: "main",
            headSHA: headSHA,
            conclusion: conclusion,
            acknowledgedAt: acked,
            createdAt: createdAt
        )
    }

    // MARK: - 그룹 키 산출

    func testGroupKeyUsesPRWhenAvailable() {
        let it = item(id: 1, prNumber: 42, headSHA: "abc")
        XCTAssertEqual(HistoryGrouping.groupKey(for: it), "pr:dlddu/pocket-aide:42")
    }

    func testGroupKeyFallsBackToHeadSHA() {
        let it = item(id: 1, prNumber: nil, headSHA: "abc123")
        XCTAssertEqual(HistoryGrouping.groupKey(for: it), "sha:dlddu/pocket-aide:abc123")
    }

    func testGroupKeyFallsBackToRowWhenNoPRAndNoSHA() {
        let it = item(id: 7, prNumber: nil, headSHA: "")
        XCTAssertEqual(HistoryGrouping.groupKey(for: it), "row:7")
    }

    // MARK: - 그룹핑 결과

    func testItemsWithSamePRAreGroupedTogether() {
        let items = [
            item(id: 1, prNumber: 42, createdAt: 100),
            item(id: 2, prNumber: 42, createdAt: 200),
            item(id: 3, prNumber: 43, createdAt: 150),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.count, 2)
        let pr42 = groups.first(where: { $0.prNumber == 42 })
        XCTAssertNotNil(pr42)
        XCTAssertEqual(pr42?.items.count, 2)
        // 그룹 내 시각 역순
        XCTAssertEqual(pr42?.items.map(\.id), [2, 1])
    }

    func testPRLessItemsFallBackToHeadSHAGroup() {
        let items = [
            item(id: 1, prNumber: nil, headSHA: "abc", createdAt: 100),
            item(id: 2, prNumber: nil, headSHA: "abc", createdAt: 200),
            item(id: 3, prNumber: nil, headSHA: "def", createdAt: 150),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.count, 2)
        let abc = groups.first(where: { $0.headSHA == "abc" })
        XCTAssertEqual(abc?.items.count, 2)
        XCTAssertEqual(abc?.id, "sha:dlddu/pocket-aide:abc")
    }

    func testNeitherPRNorHeadSHAResultsInSingletonGroups() {
        let items = [
            item(id: 1, prNumber: nil, headSHA: "", createdAt: 100),
            item(id: 2, prNumber: nil, headSHA: "", createdAt: 200),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.count, 2)
        XCTAssertTrue(groups.allSatisfy { $0.items.count == 1 })
    }

    // MARK: - 미확인 우선 정렬

    func testUnacknowledgedGroupsComeFirst() {
        // 더 최근에 도착한 그룹이 전부 확인된 상태, 더 오래된 그룹에 미확인 항목이 있어도
        // 미확인 그룹이 먼저 와야 한다 (AC11).
        let items = [
            // 그룹 A: PR #1, 모두 확인됨, 최근
            item(id: 10, prNumber: 1, acked: 999, createdAt: 1000),
            // 그룹 B: PR #2, 미확인, 더 오래됨
            item(id: 20, prNumber: 2, acked: nil, createdAt: 500),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.count, 2)
        // 첫 번째는 미확인 (PR #2), 두 번째는 확인 완료 (PR #1)
        XCTAssertEqual(groups[0].prNumber, 2)
        XCTAssertEqual(groups[1].prNumber, 1)
    }

    func testWithinUnreadSectionLatestGroupComesFirst() {
        let items = [
            item(id: 10, prNumber: 1, acked: nil, createdAt: 500),
            item(id: 20, prNumber: 2, acked: nil, createdAt: 1000),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.map(\.prNumber), [2, 1])
    }

    // MARK: - 미확인 카운트

    func testUnacknowledgedCountPerGroup() {
        let items = [
            item(id: 1, prNumber: 42, acked: nil, createdAt: 100),
            item(id: 2, prNumber: 42, acked: 500, createdAt: 200),
            item(id: 3, prNumber: 42, acked: nil, createdAt: 300),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.count, 1)
        XCTAssertEqual(groups[0].unacknowledgedCount, 2)
        XCTAssertFalse(groups[0].allAcknowledged)
    }

    func testGroupAllAcknowledgedWhenEveryItemAcked() {
        let items = [
            item(id: 1, prNumber: 42, acked: 100, createdAt: 100),
            item(id: 2, prNumber: 42, acked: 200, createdAt: 200),
        ]
        let groups = HistoryGrouping.group(items)
        XCTAssertEqual(groups.count, 1)
        XCTAssertTrue(groups[0].allAcknowledged)
        XCTAssertEqual(groups[0].unacknowledgedCount, 0)
    }

    // MARK: - 종합 상태 카운트

    func testGroupCountsSuccessAndFailureSeparately() {
        let items = [
            item(id: 1, prNumber: 42, conclusion: "success"),
            item(id: 2, prNumber: 42, conclusion: "failure"),
            item(id: 3, prNumber: 42, conclusion: "cancelled"),
            item(id: 4, prNumber: 42, conclusion: "success"),
        ]
        let groups = HistoryGrouping.group(items)
        let pr = groups[0]
        XCTAssertEqual(pr.successCount, 2)
        // failure + cancelled 모두 "실패 계열"로 카운트
        XCTAssertEqual(pr.failureCount, 2)
        XCTAssertEqual(pr.inProgressCount, 0)
    }

    func testGroupCountsInProgressSeparately() {
        // CI 시작(requested) 이벤트는 백엔드가 conclusion에 run status를 정규화해
        // 저장한다(queued / in_progress). 이들은 성공/실패가 아닌 진행 중으로 집계.
        let items = [
            item(id: 1, prNumber: 42, conclusion: "queued"),
            item(id: 2, prNumber: 42, conclusion: "in_progress"),
            item(id: 3, prNumber: 42, conclusion: "success"),
        ]
        let groups = HistoryGrouping.group(items)
        let pr = groups[0]
        XCTAssertEqual(pr.inProgressCount, 2)
        XCTAssertEqual(pr.successCount, 1)
        XCTAssertEqual(pr.failureCount, 0)
    }

    // MARK: - 입력이 정렬되지 않은 상태에서도 안정

    func testInputOrderDoesNotAffectGrouping() {
        let unsorted = [
            item(id: 3, prNumber: 42, createdAt: 300),
            item(id: 1, prNumber: 42, createdAt: 100),
            item(id: 2, prNumber: 42, createdAt: 200),
        ]
        let groups = HistoryGrouping.group(unsorted)
        XCTAssertEqual(groups.count, 1)
        XCTAssertEqual(groups[0].items.map(\.id), [3, 2, 1])
    }

    // MARK: - 빈 입력

    func testEmptyInputReturnsNoGroups() {
        XCTAssertEqual(HistoryGrouping.group([]), [])
    }
}
