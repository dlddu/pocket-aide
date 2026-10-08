import DesignSystem
import PocketAideAPI
import SwiftUI

struct RoutinesTab: View {
    var body: some View { RoutinesView() }
}

struct RoutinesView: View {
    private enum SheetMode: Equatable {
        case add
        case addStep(Routine)
        case history(Routine, [RoutineHistoryDay])
    }

    @EnvironmentObject private var auth: AppAuthCoordinator
    @StateObject private var viewModel: RoutinesViewModel
    @State private var sheetMode: SheetMode?

    init() {
        _viewModel = StateObject(wrappedValue: RoutinesViewModel(api: nil))
    }

    private static let weekdayFormatter: DateFormatter = {
        let f = DateFormatter()
        f.locale = Locale(identifier: "ko_KR")
        f.dateFormat = "M월 d일 · EEEE"
        return f
    }()

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(.routines).ignoresSafeArea()
            VStack(spacing: 0) {
                header
                list
            }
        }
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .accessibilityIdentifier("routines.screen")
        .overlay { sheet.ignoresSafeArea(.container, edges: .bottom) }
        .animation(.easeInOut(duration: 0.18), value: sheetMode)
        .toolbarVisibility(sheetMode == nil ? .automatic : .hidden, for: .tabBar)
    }

    private var header: some View {
        ScreenHeader(area: .routines, title: "루틴") {
            VStack(alignment: .trailing, spacing: 6) {
                Button {
                    sheetMode = .add
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "plus")
                            .font(.system(size: 11, weight: .bold))
                        Text("새 루틴")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                    }
                    .padding(.horizontal, 12)
                    .padding(.vertical, 7)
                    .background(Capsule().fill(DesignTokens.Color.ink(.routines)))
                    .foregroundStyle(DesignTokens.Color.surface(.routines))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("routines.add.button")
                Text(Self.weekdayFormatter.string(from: Date()))
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                    .accessibilityIdentifier("routines.today")
            }
        }
    }

    private var list: some View {
        List {
            if let message = viewModel.errorMessage {
                RoutineErrorRow(message: message, canRetry: viewModel.canRetry) {
                    Task { await viewModel.retry() }
                }
                .listRowBackground(Color.clear)
            }
            if viewModel.routines.isEmpty {
                emptyState
            } else {
                Section {
                    ForEach(viewModel.today) { routine in
                        todayRow(routine)
                    }
                } header: {
                    sectionTitle("오늘 · \(viewModel.today.count)")
                }
                if !viewModel.restingToday.isEmpty {
                    Section {
                        ForEach(viewModel.restingToday) { routine in
                            restingRow(routine)
                        }
                    } header: {
                        sectionTitle("오늘 쉬는 루틴")
                    }
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.load() }
    }

    private func sectionTitle(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
            .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.7))
            .textCase(nil)
    }

    @ViewBuilder
    private var emptyState: some View {
        Group {
            if viewModel.isLoading {
                ProgressView().frame(maxWidth: .infinity)
            } else {
                Text("아직 루틴이 없습니다. 새 루틴으로 반복되는 하루의 흐름을 적어 보세요.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                    .accessibilityIdentifier("routines.empty.state")
            }
        }
        .padding(.vertical, DesignTokens.Spacing.xl)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private func todayRow(_ routine: RoutineDay) -> some View {
        RoutineCard(
            routine: routine,
            onToggle: { step in
                Task { await viewModel.toggle(step, in: routine) }
            },
            onAddStep: {
                if let full = viewModel.routine(id: routine.id) {
                    sheetMode = .addStep(full)
                }
            },
            onHistory: { openHistory(routine.id) }
        )
        .listRowInsets(EdgeInsets(top: 6, leading: DesignTokens.Spacing.xl, bottom: 6, trailing: DesignTokens.Spacing.xl))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .swipeActions(edge: .trailing, allowsFullSwipe: false) {
            deleteButton(routine.id)
        }
    }

    private func weekdayBar(_ weekdays: Int) -> some View {
        HStack(spacing: DesignTokens.Spacing.xs) {
            ForEach(RoutineWeekdays.displayOrder, id: \.self) { index in
                let scheduled = RoutineWeekdays.contains(weekdays, index)
                Text(RoutineWeekdays.symbols[index])
                    .font(DesignTokens.Typography.font(
                        size: DesignTokens.Typography.caption2xs,
                        weight: scheduled ? .bold : .regular
                    ))
                    .foregroundStyle(scheduled
                        ? DesignTokens.Color.accent(.routines)
                        : DesignTokens.Color.ink(.routines).opacity(0.4))
                    .frame(maxWidth: .infinity)
            }
        }
    }

    private func restingRow(_ routine: Routine) -> some View {
        Button {
            openHistory(routine.id)
        } label: {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(routine.name)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                            .foregroundStyle(DesignTokens.Color.ink(.routines))
                        Text("\(routine.scheduleSummary) · 단계 \(routine.steps.count)개")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                            .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                    }
                    Spacer(minLength: 0)
                    Image(systemName: "calendar")
                        .foregroundStyle(DesignTokens.Color.accent(.routines))
                }
                if routine.cadence == .weekdays || routine.cadence == .weekly {
                    weekdayBar(routine.weekdays)
                }
            }
            .padding(14)
            .background(DesignTokens.Color.card(.routines))
            .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                    .stroke(DesignTokens.Color.rule(.routines), lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("routines.resting.\(routine.id)")
        .listRowInsets(EdgeInsets(top: 4, leading: DesignTokens.Spacing.xl, bottom: 4, trailing: DesignTokens.Spacing.xl))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .swipeActions(edge: .trailing, allowsFullSwipe: false) {
            deleteButton(routine.id)
        }
    }

    private func deleteButton(_ id: Int64) -> some View {
        Button {
            Task { await viewModel.delete(id: id) }
        } label: {
            Label("삭제", systemImage: "trash")
        }
        .tint(DesignTokens.Color.destructive(.routines))
    }

    private func openHistory(_ id: Int64) {
        guard let routine = viewModel.routine(id: id) else { return }
        Task {
            let days = await viewModel.history(for: id)
            sheetMode = .history(routine, days)
        }
    }

    @ViewBuilder
    private var sheet: some View {
        switch sheetMode {
        case .add:
            RoutineAddSheet(
                startDay: viewModel.dayKey,
                onSave: { draft in
                    sheetMode = nil
                    Task { await viewModel.create(draft) }
                },
                onCancel: { sheetMode = nil }
            )
            .transition(.move(edge: .bottom).combined(with: .opacity))
        case .addStep(let routine):
            RoutineStepAddSheet(
                routine: routine,
                onSave: { title in
                    sheetMode = nil
                    Task { await viewModel.addStep(to: routine.id, title: title) }
                },
                onDeleteStep: { step in
                    sheetMode = nil
                    Task { await viewModel.deleteStep(step.id, from: routine.id) }
                },
                onCancel: { sheetMode = nil }
            )
            .transition(.move(edge: .bottom).combined(with: .opacity))
        case .history(let routine, let days):
            RoutineHistorySheet(routine: routine, days: days, onClose: { sheetMode = nil })
                .transition(.move(edge: .bottom).combined(with: .opacity))
        case nil:
            EmptyView()
        }
    }
}

private struct RoutineCard: View {
    let routine: RoutineDay
    let onToggle: (RoutineDayStep) -> Void
    let onAddStep: () -> Void
    let onHistory: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            HStack(alignment: .top) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(routine.name)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.titleMd, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.routines))
                    Text("\(routine.scheduleSummary) · \(routine.done) / \(routine.total) 완료")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                        .accessibilityIdentifier("routines.card.\(routine.id).summary")
                }
                Spacer(minLength: 0)
                Text(routine.completed ? "완료" : routine.done == 0 ? "대기" : "\(routine.progressPercent)%")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.titleMd, weight: .bold))
                    .foregroundStyle(!routine.completed && routine.done == 0
                        ? DesignTokens.Color.ink(.routines).opacity(0.4)
                        : DesignTokens.Color.accent(.routines))
                    .accessibilityIdentifier("routines.card.\(routine.id).progress")
            }
            progressBar
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                ForEach(routine.steps) { step in
                    stepRow(step)
                }
            }
            HStack(spacing: DesignTokens.Spacing.md) {
                Button(action: onAddStep) {
                    Label("단계", systemImage: "plus")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                }
                .buttonStyle(.borderless)
                .accessibilityIdentifier("routines.card.\(routine.id).steps.button")
                Button(action: onHistory) {
                    Label("이력", systemImage: "calendar")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                }
                .buttonStyle(.borderless)
                .accessibilityIdentifier("routines.card.\(routine.id).history.button")
            }
            .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.7))
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(.routines))
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                .stroke(
                    routine.completed ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.rule(.routines),
                    lineWidth: routine.completed ? 2 : 1
                )
        )
        .accessibilityIdentifier("routines.card.\(routine.id)")
    }

    private var progressBar: some View {
        GeometryReader { proxy in
            ZStack(alignment: .leading) {
                Capsule().fill(DesignTokens.Color.soft(.routines))
                Capsule()
                    .fill(DesignTokens.Color.accent(.routines))
                    .frame(width: proxy.size.width * routine.progress)
            }
        }
        .frame(height: 6)
    }

    private func stepRow(_ step: RoutineDayStep) -> some View {
        let state: String = step.checked ? "checked" : "unchecked"
        return Button {
            onToggle(step)
        } label: {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Image(systemName: step.checked ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 18))
                    .foregroundStyle(step.checked ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.rule(.routines))
                Text(step.title)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(step.checked ? 0.5 : 1))
                    .strikethrough(step.checked)
                Spacer(minLength: 0)
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.borderless)
        .accessibilityIdentifier("routines.step.\(step.id)")
        .accessibilityValue(state)
    }
}
