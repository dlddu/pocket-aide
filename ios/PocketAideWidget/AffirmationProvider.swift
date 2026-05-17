import Foundation
import os
import PocketAideAPI
import PocketAideStorage
import WidgetKit

private let logger = Logger(subsystem: "com.dlddu.PocketAide.Widget", category: "AffirmationProvider")

/// Drives the Large widget timeline. Builds 24 entries × 30-minute strides
/// using `SeededRNG(seed: entry.date.timeIntervalSince1970)` so each entry's
/// pick is deterministic and survives across snapshot/timeline calls.
struct AffirmationProvider: TimelineProvider {
    typealias Entry = PocketAideWidgetEntry

    private static let refreshInterval: TimeInterval = 30 * 60
    private static let entryCount = 24

    private let selector = RotationSelector()

    func placeholder(in _: Context) -> PocketAideWidgetEntry {
        PocketAideWidgetEntry(
            date: Date(),
            state: .loaded(Self.previewAffirmation)
        )
    }

    func getSnapshot(in context: Context, completion: @escaping (PocketAideWidgetEntry) -> Void) {
        if context.isPreview {
            completion(placeholder(in: context))
            return
        }
        Task {
            let state = await fetchState(at: Date())
            completion(PocketAideWidgetEntry(date: Date(), state: state))
        }
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<PocketAideWidgetEntry>) -> Void) {
        Task {
            let now = Date()
            let items = await fetchAffirmations()
            switch items {
            case .success(let pool):
                if pool.isEmpty {
                    let entry = PocketAideWidgetEntry(date: now, state: .empty)
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
                    return PocketAideWidgetEntry(date: date, state: .loaded(pick))
                }
                let last = entries.last?.date ?? now
                completion(Timeline(entries: entries, policy: .after(last)))
            case .needsLogin:
                let entry = PocketAideWidgetEntry(date: now, state: .needsLogin)
                completion(Timeline(
                    entries: [entry],
                    policy: .after(now.addingTimeInterval(Self.refreshInterval))
                ))
            case .error:
                let entry = PocketAideWidgetEntry(date: now, state: .error)
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
}
