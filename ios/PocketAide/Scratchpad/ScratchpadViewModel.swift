import Foundation
import PocketAideAPI

@MainActor
final class ScratchpadViewModel: ObservableObject {
    @Published private(set) var items: [ScratchpadItem] = []
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?

    private(set) var api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    var unclassifiedCount: Int { items.count }
    var sections: [ScratchpadSection] { ScratchpadSections.group(items) }

    func load() async {
        guard let api else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            items = try await api.listScratchpad()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func add(text: String) async {
        guard let api else { return }
        do {
            let created = try await api.createScratchpadItem(text: text)
            items.insert(created, at: 0)
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func delete(id: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteScratchpadItem(id: id)
            items.removeAll { $0.id == id }
        } catch {
            errorMessage = String(describing: error)
        }
    }

    /// Returns the created affirmation when the target is 다짐, so the caller
    /// can open the priority sheet on it.
    func move(_ item: ScratchpadItem, to target: ScratchpadMoveTarget) async -> Affirmation? {
        guard let api else { return nil }
        do {
            let result = try await api.moveScratchpadItem(id: item.id, to: target)
            items.removeAll { $0.id == item.id }
            return result.affirmation
        } catch {
            errorMessage = String(describing: error)
            return nil
        }
    }

    func setPriority(of affirmation: Affirmation, text: String, priority: AffirmationPriority) async {
        guard let api else { return }
        do {
            _ = try await api.updateAffirmation(id: affirmation.id, text: text, priority: priority)
        } catch {
            errorMessage = String(describing: error)
        }
    }
}
