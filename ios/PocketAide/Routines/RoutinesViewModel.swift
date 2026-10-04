import Foundation
import PocketAideAPI

@MainActor
final class RoutinesViewModel: ObservableObject {
    @Published private(set) var today: [RoutineDay] = []
    @Published private(set) var routines: [Routine] = []
    @Published private(set) var dayKey: String = RoutineDayFormat.string(from: Date())
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?

    private(set) var api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    func replaceAPI(_ client: APIClient) {
        self.api = client
    }

    var restingToday: [Routine] {
        let scheduled = Set(today.map(\.id))
        return routines.filter { !scheduled.contains($0.id) }
    }

    func routine(id: Int64) -> Routine? {
        routines.first { $0.id == id }
    }

    func load() async {
        guard let api else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        let key = RoutineDayFormat.string(from: Date())
        do {
            let all = try await api.listRoutines()
            let scheduled = try await api.listRoutines(on: key)
            dayKey = key
            routines = all
            today = scheduled
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func create(_ draft: RoutineDraft) async {
        guard let api else { return }
        do {
            _ = try await api.createRoutine(draft)
            await load()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func delete(id: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteRoutine(id: id)
            routines.removeAll { $0.id == id }
            today.removeAll { $0.id == id }
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func addStep(to routineID: Int64, title: String) async {
        guard let api else { return }
        do {
            _ = try await api.addRoutineStep(routineID: routineID, title: title)
            await load()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func deleteStep(_ stepID: Int64, from routineID: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteRoutineStep(routineID: routineID, stepID: stepID)
            await load()
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func toggle(_ step: RoutineDayStep, in routine: RoutineDay) async {
        guard let api else { return }
        do {
            let updated = try await api.setRoutineStep(
                routineID: routine.id,
                stepID: step.id,
                day: routine.day,
                checked: !step.checked
            )
            if let index = today.firstIndex(where: { $0.id == updated.id }) {
                today[index] = updated
            }
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func history(for routineID: Int64) async -> [RoutineHistoryDay] {
        guard let api else { return [] }
        do {
            return try await api.routineHistory(routineID: routineID, endingOn: dayKey)
        } catch {
            errorMessage = String(describing: error)
            return []
        }
    }
}
