// 검증 시나리오: test-github-monitor.md#시나리오 16
import XCTest

final class WebhookHistoryGroupingUITests: XCTestCase {
    private let workflows = ["Build", "Lint", "Deploy", "Docs"]

    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    private func events(repo: String, token: String) -> [WebhookEvent] {
        let first = WebhookHistoryUI.commit(1, token: token)
        let direct = WebhookHistoryUI.commit(2, token: token)
        let second = WebhookHistoryUI.commit(3, token: token)
        return [
            WebhookEvent(repo: repo, pullRequest: 1, commit: first, branch: "feature/one", workflow: "Build", conclusion: "success"),
            WebhookEvent(repo: repo, pullRequest: nil, commit: direct, branch: "main", workflow: "Deploy", conclusion: "success"),
            WebhookEvent(repo: repo, pullRequest: 2, commit: second, branch: "feature/two", workflow: "Build", conclusion: "success"),
            WebhookEvent(repo: repo, pullRequest: 1, commit: first, branch: "feature/one", workflow: "Lint", conclusion: "failure"),
            WebhookEvent(repo: repo, pullRequest: nil, commit: direct, branch: "main", workflow: "Docs", conclusion: "success")
        ]
    }

    private func groupOrder(of sent: [WebhookEvent]) -> [String] {
        var keys: [String] = []
        for event in sent.reversed() where !keys.contains(event.groupKey) {
            keys.append(event.groupKey)
        }
        return keys
    }

    private func title(of lead: WebhookEvent) -> String {
        if let number = lead.pullRequest {
            return "\(lead.repo) · #\(number)"
        }
        return "\(lead.repo) · \(lead.branch) @\(lead.commit.prefix(7))"
    }

    private func assertHeader(_ labels: [String], of members: [WebhookEvent], group: String) {
        guard let lead = members.first else {
            XCTFail("Group \(group) should have members")
            return
        }
        let passed = members.filter { $0.conclusion == "success" }.count
        let failed = members.count - passed
        XCTAssertTrue(labels.contains(title(of: lead)), "Group \(group) should be titled \(title(of: lead)): \(labels)")
        XCTAssertTrue(labels.contains("CI \(members.count)건"), "Group \(group) should count \(members.count) items: \(labels)")
        XCTAssertTrue(labels.contains("미확인 \(members.count)"), "Group \(group) should count \(members.count) unacknowledged: \(labels)")
        XCTAssertEqual(labels.contains("통과 \(passed)"), passed > 0, "Group \(group) should summarise \(passed) passed: \(labels)")
        XCTAssertEqual(labels.contains("실패 \(failed)"), failed > 0, "Group \(group) should summarise \(failed) failed: \(labels)")
        XCTAssertTrue(labels.contains { $0.hasPrefix("최근 ") }, "Group \(group) should show its latest event time: \(labels)")
    }

    private func normalized(_ labels: [String], repo: String, token: String) -> [String] {
        labels
            .filter { !$0.hasPrefix("최근 ") && $0.range(of: "^[0-9]{2}:[0-9]{2}$", options: .regularExpression) == nil }
            .map { $0.replacingOccurrences(of: repo, with: "R").replacingOccurrences(of: String(token.prefix(6)), with: "T") }
            .sorted()
    }

    private func composition(of sent: [WebhookEvent], repo: String, token: String) -> [[String]] {
        let order = groupOrder(of: sent)
        guard let top = order.first else { return [] }
        let count = sent.filter { $0.groupKey == top }.count
        let app = WebhookHistoryUI.openMonitor(until: top, reads: "CI \(count)건")
        var seen = Set<String>()
        var result: [String: [String]] = [:]
        for group in order {
            XCTAssertTrue(WebhookHistoryUI.reveal(group, in: app), "Group \(group) should be listed")
            let labels = WebhookHistoryUI.labels(of: group, in: app)
            let members = sent.filter { $0.groupKey == group }
            assertHeader(labels, of: members, group: group)
            XCTAssertEqual(
                labels.filter { workflows.contains($0) },
                members.reversed().map { $0.workflow },
                "Group \(group) should list its items newest first: \(labels)"
            )
            seen.formUnion(WebhookHistoryUI.groups(of: repo, in: app))
            result[group.replacingOccurrences(of: repo, with: "R").replacingOccurrences(of: token, with: "T")] =
                normalized(labels, repo: repo, token: token)
        }
        XCTAssertEqual(seen, Set(order), "The five events of \(repo) should form exactly these group cards")
        XCTAssertEqual(order.count, 3, "Two PRs and one PR-less commit should make three groups")
        app.terminate()
        return result.keys.sorted().compactMap { result[$0] }
    }

    func testEventsOfTheSamePullRequestOrCommitShareOneGroupCardWhateverTheArrivalOrder() {
        let token = WebhookHistoryUI.runToken()
        let repo = "dlddu/e2e-grouping-\(token)"
        let sent = events(repo: repo, token: token)
        WebhookHistoryUI.send(sent)
        let first = composition(of: sent, repo: repo, token: token)

        let otherToken = WebhookHistoryUI.runToken()
        let otherRepo = "dlddu/e2e-regrouped-\(otherToken)"
        let reordered = Array(events(repo: otherRepo, token: otherToken).reversed())
        WebhookHistoryUI.send(reordered)
        let second = composition(of: reordered, repo: otherRepo, token: otherToken)

        XCTAssertEqual(first, second, "A different arrival order should give the same groups")
    }
}
