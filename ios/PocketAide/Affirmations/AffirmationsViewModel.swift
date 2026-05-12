import Foundation
import PocketAideAPI
import PocketAideStorage
import SwiftUI

@MainActor
final class AffirmationsViewModel: ObservableObject {
    @Published var items: [Affirmation] = []
    @Published var heroId: Int64?
    @Published var isLoading = false
    @Published var loadError: String?

    private let api: APIClient?
    private let selector = RotationSelector()
    private var seededRng: SplitMix64?

    init(api: APIClient? = nil) {
        if let api {
            self.api = api
        } else {
            let store = KeychainTokenStore(accessGroup: nil)
            self.api = try? APIClient.fromBundle(.main, tokenStore: store)
        }

        if let seedRaw = ProcessInfo.processInfo.environment["ROTATION_SEED"],
           let seed = UInt64(seedRaw) {
            self.seededRng = SplitMix64(seed: seed)
        }
    }

    var hero: Affirmation? {
        guard let id = heroId else { return items.first }
        return items.first(where: { $0.id == id })
    }

    func load() async {
        guard let api else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let fetched = try await api.listAffirmations()
            items = fetched
            rotateHero()
        } catch {
            loadError = String(describing: error)
        }
    }

    func rotateHero() {
        if seededRng != nil {
            var rng = seededRng!
            heroId = selector.pick(from: items, using: &rng)?.id
            seededRng = rng
        } else {
            var rng = SystemRandomNumberGenerator()
            heroId = selector.pick(from: items, using: &rng)?.id
        }
    }

    func add(text: String, priority: Priority) async {
        guard let api else { return }
        do {
            let created = try await api.createAffirmation(text: text, priority: priority)
            items.insert(created, at: 0)
            heroId = created.id
        } catch {
            loadError = String(describing: error)
        }
    }

    func update(id: Int64, text: String, priority: Priority) async {
        guard let api else { return }
        do {
            let updated = try await api.updateAffirmation(id: id, text: text, priority: priority)
            if let idx = items.firstIndex(where: { $0.id == id }) {
                items[idx] = updated
            }
            heroId = updated.id
        } catch {
            loadError = String(describing: error)
        }
    }

    func delete(id: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteAffirmation(id: id)
            items.removeAll(where: { $0.id == id })
            if heroId == id {
                rotateHero()
            }
        } catch {
            loadError = String(describing: error)
        }
    }
}

private struct SplitMix64: RandomNumberGenerator {
    private var state: UInt64

    init(seed: UInt64) {
        self.state = seed == 0 ? 0x9E3779B97F4A7C15 : seed
    }

    mutating func next() -> UInt64 {
        state &+= 0x9E3779B97F4A7C15
        var z = state
        z = (z ^ (z &>> 30)) &* 0xBF58476D1CE4E5B9
        z = (z ^ (z &>> 27)) &* 0x94D049BB133111EB
        return z ^ (z &>> 31)
    }
}
