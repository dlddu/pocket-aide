import Foundation
import os
import PocketAideAPI
import PocketAideStorage
import WidgetKit

private let logger = Logger(subsystem: "com.dlddu.PocketAide.Widget", category: "AffirmationProvider")

/// Each entry's pick is seeded by its date (`SeededRNG`) so it is deterministic
/// and survives across snapshot/timeline calls.
struct AffirmationProvider: TimelineProvider {
    typealias Entry = PocketAideWidgetEntry

    private static let refreshInterval: TimeInterval = 30 * 60
    private static let entryCount = 24

    private let selector = RotationSelector()

    func placeholder(in _: Context) -> PocketAideWidgetEntry {
        PocketAideWidgetEntry(
            date: Date(),
            state: .loaded(Self.previewAffirmation),
            calendar: .loaded(Self.previewEvents)
        )
    }

    func getSnapshot(in context: Context, completion: @escaping (PocketAideWidgetEntry) -> Void) {
        if context.isPreview {
            completion(placeholder(in: context))
            return
        }
        Task {
            let now = Date()
            let state = await fetchState(at: now)
            let calendar = CalendarSnapshot.load(from: now, through: now).state(at: now)
            completion(PocketAideWidgetEntry(date: now, state: state, calendar: calendar))
        }
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<PocketAideWidgetEntry>) -> Void) {
        Task {
            let now = Date()
            let items = await fetchAffirmations()
            let lastEntryDate = now.addingTimeInterval(Double(Self.entryCount - 1) * Self.refreshInterval)
            let snapshot = CalendarSnapshot.load(from: now, through: lastEntryDate)
            switch items {
            case .success(let pool):
                if pool.isEmpty {
                    let entry = PocketAideWidgetEntry(date: now, state: .empty, calendar: snapshot.state(at: now))
                    completion(Timeline(
                        entries: [entry],
                        policy: .after(now.addingTimeInterval(Self.refreshInterval))
                    ))
                    return
                }
                let entries = (0..<Self.entryCount).map { offset -> PocketAideWidgetEntry in
                    let date = now.addingTimeInterval(Double(offset) * Self.refreshInterval)
                    var rng = SeededRNG(seed: UInt64(date.timeIntervalSince1970))
                    let pick = selector.pick(from: pool, using: &rng) ?? pool[0]
                    return PocketAideWidgetEntry(date: date, state: .loaded(pick), calendar: snapshot.state(at: date))
                }
                let last = entries.last?.date ?? now
                let reload = snapshot.authorized ? now.addingTimeInterval(Self.refreshInterval) : last
                completion(Timeline(entries: entries, policy: .after(reload)))
            case .needsLogin:
                let entry = PocketAideWidgetEntry(date: now, state: .needsLogin, calendar: snapshot.state(at: now))
                completion(Timeline(
                    entries: [entry],
                    policy: .after(now.addingTimeInterval(Self.refreshInterval))
                ))
            case .error:
                let entry = PocketAideWidgetEntry(date: now, state: .error, calendar: snapshot.state(at: now))
                // Back off on errors so a flapping backend doesn't burn the
                // system's per-widget refresh budget.
                completion(Timeline(
                    entries: [entry],
                    policy: .after(now.addingTimeInterval(60 * 60))
                ))
            }
        }
    }

    private func fetchState(at date: Date) async -> WidgetAffirmationState {
        let result = await fetchAffirmations()
        switch result {
        case .success(let pool):
            guard !pool.isEmpty else { return .empty }
            var rng = SeededRNG(seed: UInt64(date.timeIntervalSince1970))
            let pick = selector.pick(from: pool, using: &rng) ?? pool[0]
            return .loaded(pick)
        case .needsLogin:
            return .needsLogin
        case .error:
            return .error
        }
    }

    private enum FetchResult {
        case success([Affirmation])
        case needsLogin
        case error
    }

    private func fetchAffirmations() async -> FetchResult {
        let accessGroup = Bundle.main.object(forInfoDictionaryKey: "KeychainAccessGroup") as? String
        let resolvedGroup = accessGroup.flatMap { $0.isEmpty ? nil : $0 }
        let baseURL = Bundle.main.object(forInfoDictionaryKey: "BackendBaseURL") as? String ?? "<nil>"
        logger.debug("fetch start: accessGroup=\(resolvedGroup ?? "<nil>", privacy: .public) baseURL=\(baseURL, privacy: .public)")

        let store = KeychainTokenStore(accessGroup: resolvedGroup)

        do {
            if (try store.load()) == nil {
                logger.info("no token in keychain → needsLogin")
                return .needsLogin
            }
        } catch {
            logger.error("keychain load failed: \(String(describing: error), privacy: .public)")
            return .error
        }

        let api: APIClient
        do {
            api = try APIClient.fromBundle(.main, tokenStore: store)
        } catch {
            logger.error("APIClient.fromBundle failed: \(String(describing: error), privacy: .public)")
            return .error
        }

        do {
            let items = try await api.listAffirmations()
            logger.info("fetched \(items.count, privacy: .public) affirmations")
            return .success(items)
        } catch APIError.badStatus(401, _) {
            logger.info("listAffirmations → 401, needsLogin")
            return .needsLogin
        } catch {
            logger.error("listAffirmations failed: \(String(describing: error), privacy: .public)")
            return .error
        }
    }

    private static let previewAffirmation = Affirmation(
        id: 0,
        text: "작게 시작해서 매일 1%씩. 1년에 37배.",
        priority: .high,
        createdAt: 0,
        updatedAt: 0
    )

    private static let previewEvents = UpcomingEventsSummary(
        events: [
            UpcomingEvent(
                title: "분기 리뷰 · 회의실 4",
                start: Date().addingTimeInterval(60 * 60),
                end: Date().addingTimeInterval(150 * 60),
                isAllDay: false
            ),
        ],
        moreCount: 2
    )
}
