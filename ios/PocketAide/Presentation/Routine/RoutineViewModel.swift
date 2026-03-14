// RoutineViewModel.swift
// PocketAide

import Foundation

/// 루틴 화면의 상태와 비즈니스 로직을 관리하는 ViewModel.
@MainActor
final class RoutineViewModel: ObservableObject {

    // MARK: - Published State

    @Published var routines: [Routine] = []
    @Published var isLoading = false
    @Published var errorMessage: String? = nil
    @Published var showAddSheet = false

    // MARK: - Computed Sections

    /// D-day가 0 이하인 루틴 (오늘 또는 기한 초과)
    var urgentRoutines: [Routine] { routines.filter { $0.dDay <= 0 } }

    /// D-day가 1 이상인 루틴 (여유 있음)
    var relaxedRoutines: [Routine] { routines.filter { $0.dDay > 0 } }

    // MARK: - Dependencies

    private let routineService = RoutineService()
    private let keychainService = KeychainService()

    // MARK: - Init

    init() {}

    // MARK: - Public Methods

    /// 루틴 목록을 서버에서 불러옵니다.
    func loadRoutines() async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            routines = try await routineService.list(serverURL: serverURL, token: token)
        } catch {
            errorMessage = "루틴 목록을 불러오지 못했습니다."
        }
    }

    /// 새 루틴을 생성합니다.
    func createRoutine(name: String, intervalDays: Int, lastDoneAt: String) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let created = try await routineService.create(
                name: name,
                intervalDays: intervalDays,
                lastDoneAt: lastDoneAt,
                serverURL: serverURL,
                token: token
            )
            routines.append(created)
        } catch {
            errorMessage = "루틴을 생성하지 못했습니다."
        }
    }

    /// 루틴을 수정합니다.
    func updateRoutine(id: Int, name: String? = nil, intervalDays: Int? = nil, note: String? = nil) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let updated = try await routineService.update(
                id: id,
                name: name,
                intervalDays: intervalDays,
                note: note,
                serverURL: serverURL,
                token: token
            )
            if let idx = routines.firstIndex(where: { $0.id == id }) {
                routines[idx] = updated
            }
        } catch {
            errorMessage = "루틴을 수정하지 못했습니다."
        }
    }

    /// 루틴을 삭제합니다.
    func deleteRoutine(id: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await routineService.delete(id: id, serverURL: serverURL, token: token)
            routines.removeAll { $0.id == id }
        } catch {
            errorMessage = "루틴을 삭제하지 못했습니다."
        }
    }

    /// 루틴을 완료 처리합니다.
    func completeRoutine(id: Int) async {
        guard let serverURL = keychainService.loadServerURL(),
              let token = keychainService.loadToken() else {
            errorMessage = "서버 주소 또는 인증 토큰이 없습니다."
            return
        }

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let completed = try await routineService.complete(id: id, serverURL: serverURL, token: token)
            if let idx = routines.firstIndex(where: { $0.id == id }) {
                routines[idx] = completed
            }
        } catch {
            errorMessage = "루틴 완료 처리에 실패했습니다."
        }
    }
}
