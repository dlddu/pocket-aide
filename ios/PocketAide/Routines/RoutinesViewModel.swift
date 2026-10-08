import Foundation
import PocketAideAPI

@MainActor
final class RoutinesViewModel: ObservableObject {
    @Published private(set) var today: [RoutineDay] = []
    @Published private(set) var routines: [Routine] = []
    @Published private(set) var dayKey: String = RoutineDayFormat.string(from: RoutinesViewModel.launchToday())
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?

    private(set) var api: APIClient?
    private var retryAction: (() async -> Void)?

    var canRetry: Bool { retryAction != nil }

    func retry() async {
        guard let retryAction else { return }
        await retryAction()
    }

    private func clearFailure() {
        retryAction = nil
        errorMessage = nil
    }

    private func fail(_ action: RoutineFailureCopy.Action, _ error: Error, retry: (() async -> Void)?) {
        retryAction = retry
        errorMessage = RoutineFailureCopy.message(for: action, error: error)
    }

    init(api: APIClient?) {
        self.api = api
    }

    // mock-exception: DET — 「오늘」은 기기 달력이 정하고 XCUITest 는 시뮬레이터 시계를 바꿀 수 없다; UI 테스트만 「오늘」로 쓸 날짜 값을 주입한다 (docs/e2e-mocking-policy.md)
    nonisolated static func launchToday() -> Date {
        let parts = (ProcessInfo.processInfo.environment["ROUTINES_TODAY"] ?? "").split(separator: "-").compactMap { Int($0) }
        guard parts.count == 3,
              let day = Calendar.current.date(from: DateComponents(year: parts[0], month: parts[1], day: parts[2], hour: 12)) else {
            return Date()
        }
        return day
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
        clearFailure()
        defer { isLoading = false }
        let key = RoutineDayFormat.string(from: Self.launchToday())
        do {
            let all = try await api.listRoutines()
            let scheduled = try await api.listRoutines(on: key)
            dayKey = key
            routines = all
            today = scheduled
        } catch {
            fail(.load, error) { [weak self] in await self?.load() }
        }
    }

    func create(_ draft: RoutineDraft) async {
        guard let api else { return }
        do {
            _ = try await api.createRoutine(draft)
            await load()
        } catch {
            fail(.create, error) { [weak self] in await self?.create(draft) }
        }
    }

    func delete(id: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteRoutine(id: id)
            routines.removeAll { $0.id == id }
            today.removeAll { $0.id == id }
            clearFailure()
        } catch {
            fail(.delete, error) { [weak self] in await self?.delete(id: id) }
        }
    }

    func addStep(to routineID: Int64, title: String) async {
        guard let api else { return }
        do {
            _ = try await api.addRoutineStep(routineID: routineID, title: title)
            await load()
        } catch {
            fail(.addStep, error) { [weak self] in await self?.addStep(to: routineID, title: title) }
        }
    }

    func deleteStep(_ stepID: Int64, from routineID: Int64) async {
        guard let api else { return }
        do {
            try await api.deleteRoutineStep(routineID: routineID, stepID: stepID)
            await load()
        } catch {
            fail(.deleteStep, error) { [weak self] in await self?.deleteStep(stepID, from: routineID) }
        }
    }

    func toggle(_ step: RoutineDayStep, in routine: RoutineDay) async {
        await setStep(step, in: routine, checked: !step.checked)
    }

    private func setStep(_ step: RoutineDayStep, in routine: RoutineDay, checked: Bool) async {
        guard let api else { return }
        do {
            let updated = try await api.setRoutineStep(
                routineID: routine.id,
                stepID: step.id,
                day: routine.day,
                checked: checked
            )
            if let index = today.firstIndex(where: { $0.id == updated.id }) {
                today[index] = updated
            }
            clearFailure()
        } catch {
            fail(.check, error) { [weak self] in await self?.setStep(step, in: routine, checked: checked) }
        }
    }

    func history(for routineID: Int64) async -> [RoutineHistoryDay] {
        guard let api else { return [] }
        do {
            return try await api.routineHistory(routineID: routineID, endingOn: dayKey)
        } catch {
            fail(.history, error, retry: nil)
            return []
        }
    }
}
