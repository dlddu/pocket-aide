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
            calendar: .loaded(Self.previewEvents),
            weather: .loaded(Self.previewWeather, place: "서울"),
            notifications: .loaded(WidgetNotificationsSummary(latest: Self.previewNotification, moreCount: 1)),
            scratchpad: .loaded(unclassified: 12)
        )
    }

    func getSnapshot(in context: Context, completion: @escaping (PocketAideWidgetEntry) -> Void) {
        if context.isPreview {
            completion(placeholder(in: context))
            return
        }
        Task {
            let now = Date()
            let client = makeClient()
            let state = await fetchState(client, at: now)
            let notifications = await fetchNotifications(client)
            let scratchpad = await fetchScratchpad(client)
            let calendar = CalendarSnapshot.load(from: now, through: now).state(at: now)
            let weather = await fetchWeather()
            completion(PocketAideWidgetEntry(
                date: now,
                state: state,
                calendar: calendar,
                weather: weather,
                notifications: notifications,
                scratchpad: scratchpad
            ))
        }
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<PocketAideWidgetEntry>) -> Void) {
        Task {
            let now = Date()
            let client = makeClient()
            let items = await fetchAffirmations(client)
            let notifications = await fetchNotifications(client)
            let scratchpad = await fetchScratchpad(client)
            let weather = await fetchWeather()
            let lastEntryDate = now.addingTimeInterval(Double(Self.entryCount - 1) * Self.refreshInterval)
            let snapshot = CalendarSnapshot.load(from: now, through: lastEntryDate)
            let single = { (state: WidgetAffirmationState) in
                [PocketAideWidgetEntry(
                    date: now,
                    state: state,
                    calendar: snapshot.state(at: now),
                    weather: weather,
                    notifications: notifications,
                    scratchpad: scratchpad
                )]
            }
            let nextReload = now.addingTimeInterval(Self.refreshInterval)
            switch items {
            case .success(let pool):
                if pool.isEmpty {
                    completion(Timeline(entries: single(.empty), policy: .after(nextReload)))
                    return
                }
                let entries = (0..<Self.entryCount).map { offset -> PocketAideWidgetEntry in
                    let date = now.addingTimeInterval(Double(offset) * Self.refreshInterval)
                    var rng = SeededRNG(seed: UInt64(date.timeIntervalSince1970))
                    let pick = selector.pick(from: pool, using: &rng) ?? pool[0]
                    return PocketAideWidgetEntry(
                        date: date,
                        state: .loaded(pick),
                        calendar: snapshot.state(at: date),
                        weather: weather,
                        notifications: notifications,
                        scratchpad: scratchpad
                    )
                }
                completion(Timeline(entries: entries, policy: .after(nextReload)))
            case .needsLogin:
                completion(Timeline(entries: single(.needsLogin), policy: .after(nextReload)))
            case .error:
                completion(Timeline(entries: single(.error), policy: .after(nextReload)))
            }
        }
    }

    private func fetchState(_ client: ClientResult, at date: Date) async -> WidgetAffirmationState {
        let result = await fetchAffirmations(client)
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

    private enum ClientResult {
        case ready(APIClient)
        case needsLogin
        case error
    }

    private func makeClient() -> ClientResult {
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

        do {
            let api = try APIClient.fromBundle(.main, tokenStore: store)
            return .ready(api)
        } catch {
            logger.error("APIClient.fromBundle failed: \(String(describing: error), privacy: .public)")
            return .error
        }
    }

    private func fetchAffirmations(_ client: ClientResult) async -> FetchResult {
        let api: APIClient
        switch client {
        case .ready(let ready): api = ready
        case .needsLogin: return .needsLogin
        case .error: return .error
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

    private func fetchNotifications(_ client: ClientResult) async -> WidgetNotificationState {
        let api: APIClient
        switch client {
        case .ready(let ready): api = ready
        case .needsLogin: return .needsLogin
        case .error: return .error
        }
        do {
            let items = try await api.listNotificationHistory()
            let settings = try await api.notificationSettings()
            return .loaded(WidgetNotifications.summarize(items, settings: settings))
        } catch APIError.badStatus(401, _) {
            logger.info("listNotificationHistory → 401, needsLogin")
            return .needsLogin
        } catch {
            logger.error("listNotificationHistory failed: \(String(describing: error), privacy: .public)")
            return .error
        }
    }

    private func fetchScratchpad(_ client: ClientResult) async -> WidgetScratchpadState {
        let api: APIClient
        switch client {
        case .ready(let ready): api = ready
        case .needsLogin: return .needsLogin
        case .error: return .error
        }
        do {
            return .loaded(unclassified: try await api.listScratchpad().count)
        } catch APIError.badStatus(401, _) {
            logger.info("listScratchpad → 401, needsLogin")
            return .needsLogin
        } catch {
            logger.error("listScratchpad failed: \(String(describing: error), privacy: .public)")
            return .error
        }
    }

    private func fetchWeather() async -> WidgetWeatherState {
        guard let location = WeatherLocationStore().load() else { return .needsLocation }
        do {
            return .loaded(try await WeatherClient.fetchForecast(location).forecast.summary, place: location.placeName)
        } catch {
            logger.error("weather fetch failed: \(String(describing: error), privacy: .public)")
            return .error
        }
    }

    private static let previewWeather = WeatherSummary(
        temperature: 18,
        condition: "비",
        high: 22,
        low: 14,
        precipitationChance: 60
    )

    private static let previewNotification = NotificationHistoryItem(
        id: 0,
        repoFullName: "dlddu/pocket-aide",
        prNumber: 7,
        prTitle: "feat(widget): 알림 모음",
        prURL: nil,
        commitURL: nil,
        runURL: nil,
        workflowName: "CI",
        headBranch: "main",
        headSHA: "",
        conclusion: "success",
        acknowledgedAt: nil,
        createdAt: 0
    )

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
