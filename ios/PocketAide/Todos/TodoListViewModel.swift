import Foundation
import PocketAideAPI

@MainActor
final class TodoListViewModel: ObservableObject {
    let area: TodoArea
    @Published private(set) var items: [TodoItem] = []
    @Published private(set) var isLoading = false
    @Published var query = ""
    @Published var errorMessage: String?

    private(set) var api: APIClient?

    init(area: TodoArea, api: APIClient?) {
        self.area = area
        self.api = api
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    var visibleItems: [TodoItem] { TodoSearch.filter(items, query: query) }
    var openItems: [TodoItem] { visibleItems.filter { !$0.isDone } }
    var doneItems: [TodoItem] { visibleItems.filter(\.isDone) }
    var openCount: Int { items.filter { !$0.isDone }.count }
    var doneCount: Int { items.count - openCount }

    func load() async {
        guard let api else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            items = try await api.listTodos(area: area)
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func add(_ draft: TodoDraft) async {
        guard let api else { return }
        do {
            _ = try await api.createTodo(area: area, draft: draft)
            await load()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func update(id: Int64, draft: TodoDraft) async {
        guard let api else { return }
        do {
            _ = try await api.updateTodo(area: area, id: id, draft: draft)
            await load()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func toggleDone(_ item: TodoItem) async {
        var draft = TodoDraft(item)
        draft.done.toggle()
        await update(id: item.id, draft: draft)
    }

    func delete(id: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteTodo(area: area, id: id)
            items.removeAll { $0.id == id }
        } catch {
            errorMessage = String(describing: error)
        }
    }
}
