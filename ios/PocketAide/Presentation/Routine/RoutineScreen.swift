// RoutineScreen.swift
// PocketAide

import SwiftUI

/// 루틴 목록을 표시하는 메인 화면.
///
/// AccessibilityIdentifier 목록:
/// - routine_list_view        : root container
/// - routine_section_urgent   : 곧 해야 할 것 섹션
/// - routine_section_relaxed  : 여유 있음 섹션
/// - add_routine_button       : 우측 상단 + 버튼
/// - routine_row_{name}       : 각 루틴 행
/// - routine_dday_label       : 각 행의 D-day 레이블
/// - complete_button          : 스와이프 완료 버튼
struct RoutineScreen: View {

    @StateObject private var viewModel = RoutineViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.routines.isEmpty && !viewModel.isLoading {
                    emptyView
                } else {
                    routineListView
                }
            }
            .navigationTitle("루틴")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        viewModel.showAddSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                    .accessibilityIdentifier("add_routine_button")
                }
            }
            .sheet(isPresented: $viewModel.showAddSheet) {
                RoutineEditScreen(viewModel: viewModel)
            }
            .task {
                await viewModel.loadRoutines()
            }
            .refreshable {
                await viewModel.loadRoutines()
            }
        }
        .accessibilityIdentifier("routine_list_view")
    }

    // MARK: - Subviews

    private var routineListView: some View {
        List {
            if !viewModel.urgentRoutines.isEmpty {
                Section("곧 해야 할 것") {
                    ForEach(viewModel.urgentRoutines) { routine in
                        RoutineRow(routine: routine, viewModel: viewModel)
                    }
                }
                .accessibilityIdentifier("routine_section_urgent")
            }

            if !viewModel.relaxedRoutines.isEmpty {
                Section("여유 있음") {
                    ForEach(viewModel.relaxedRoutines) { routine in
                        RoutineRow(routine: routine, viewModel: viewModel)
                    }
                }
                .accessibilityIdentifier("routine_section_relaxed")
            }
        }
    }

    private var emptyView: some View {
        VStack(spacing: 16) {
            Image(systemName: "arrow.clockwise.circle")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)
            Text("루틴이 없습니다")
                .font(.headline)
            Text("+ 버튼으로 루틴을 추가해보세요")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - RoutineRow

/// 루틴 목록의 개별 행 뷰.
private struct RoutineRow: View {

    let routine: Routine
    let viewModel: RoutineViewModel

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(routine.name)
                    .font(.body)
                Text("\(routine.intervalDays)일 주기")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            dDayLabel
        }
        .accessibilityIdentifier("routine_row_\(routine.name)")
        .swipeActions(edge: .leading) {
            Button {
                Task {
                    await viewModel.completeRoutine(id: routine.id)
                }
            } label: {
                Label("완료", systemImage: "checkmark.circle")
            }
            .tint(.green)
            .accessibilityIdentifier("complete_button")
        }
        .swipeActions(edge: .trailing) {
            Button(role: .destructive) {
                Task {
                    await viewModel.deleteRoutine(id: routine.id)
                }
            } label: {
                Label("삭제", systemImage: "trash")
            }
        }
    }

    private var dDayLabel: some View {
        let (text, color) = dDayDisplay(routine.dDay)
        return Text(text)
            .font(.caption.bold())
            .foregroundStyle(color)
            .accessibilityIdentifier("routine_dday_label")
    }

    private func dDayDisplay(_ dDay: Int) -> (String, Color) {
        if dDay == 0 {
            return ("D-Day", .orange)
        } else if dDay > 0 {
            return ("D-\(dDay)", .blue)
        } else {
            return ("D+\(-dDay)", .red)
        }
    }
}
