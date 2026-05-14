import Foundation
import PocketAideAPI

@MainActor
final class AffirmationsViewModel: ObservableObject {
    @Published private(set) var items: [Affirmation] = []
    @Published private(set) var heroID: Int64?
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?

    private(set) var api: APIClient?
    let selector = RotationSelector()
    let rotationSeed: UInt64?

    init(api: APIClient?, rotationSeed: UInt64? = nil) {
        self.api = api
        self.rotationSeed = rotationSeed
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    func load() async {
        guard let api else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            let loaded = try await api.listAffirmations()
            items = loaded
            rotateHero()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func add(text: String, priority: AffirmationPriority) async {
        guard let api else { return }
        do {
            let created = try await api.createAffirmation(text: text, priority: priority)
            items.insert(created, at: 0)
            heroID = created.id // new sentences become the hero immediately (PRD-5 AC: 자동 노출)
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func update(id: Int64, text: String, priority: AffirmationPriority) async {
        guard let api else { return }
        do {
            let updated = try await api.updateAffirmation(id: id, text: text, priority: priority)
            if let idx = items.firstIndex(where: { $0.id == id }) {
                items[idx] = updated
            }
            heroID = updated.id
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func delete(id: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteAffirmation(id: id)
            items.removeAll { $0.id == id }
            if heroID == id {
                rotateHero()
            }
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func rotateHero() {
        if let seed = rotationSeed {
            var rng = SeededRNG(seed: seed)
            heroID = selector.pick(from: items, using: &rng)?.id
        } else {
            var rng = SystemRandomNumberGenerator()
            heroID = selector.pick(from: items, using: &rng)?.id
        }
    }

    var heroItem: Affirmation? {
        guard let heroID else { return items.first }
        return items.first(where: { $0.id == heroID }) ?? items.first
    }
}

struct SeededRNG: RandomNumberGenerator {
    private var state: UInt64

    init(seed: UInt64) {
        self.state = seed == 0 ? 0xDEAD_BEEF_CAFE_F00D : seed
    }

    mutating func next() -> UInt64 {
        state &+= 0x9E37_79B9_7F4A_7C15
        var z = state
        z = (z ^ (z &>> 30)) &* 0xBF58_476D_1CE4_E5B9
        z = (z ^ (z &>> 27)) &* 0x94D0_49BB_1331_11EB
        return z ^ (z &>> 31)
    }
}
