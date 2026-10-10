import DesignSystem
import PocketAideAPI
import SwiftUI

extension TodoArea {
    var designArea: DesignTokens.Area {
        switch self {
        case .personal: return .personal
        case .work: return .work
        }
    }

    /// PRD-3 AC3: the two areas differ in wording as well as palette — the
    /// personal tab reads like a notebook, the work tab like a status board.
    var screenTitle: String {
        switch self {
        case .personal: return "내 일"
        case .work: return "회사"
        }
    }

    var searchPrompt: String {
        switch self {
        case .personal: return "개인 영역만 검색…"
        case .work: return "회사 영역만 검색…"
        }
    }

    var rowRadius: CGFloat {
        switch self {
        case .personal: return DesignTokens.Radius.card
        case .work: return DesignTokens.Radius.key
        }
    }

    func summary(open: Int, done: Int) -> String {
        switch self {
        case .personal: return "\(open)개 남음 · \(done)개 완료"
        case .work: return "\(open) OPEN · \(done) DONE"
        }
    }
}

struct PersonalTab: View {
    var body: some View { TodoListView(area: .personal) }
}

struct WorkTab: View {
    var body: some View { TodoListView(area: .work) }
}

/// PRD-3 개인/회사 투두. One view type serves both tabs, but each instance is
/// bound to a single area for its lifetime and only ever loads that area's
/// collection — there is no control that moves an item across (AC4).
struct TodoListView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @StateObject private var viewModel: TodoListViewModel
    @State private var sheetMode: TodoEditSheet.Mode?

    private let area: TodoArea

    init(area: TodoArea) {
        self.area = area
        _viewModel = StateObject(wrappedValue: TodoListViewModel(area: area, api: nil))
    }

    private var tone: DesignTokens.Area { area.designArea }
    private var idPrefix: String { "todos.\(area.rawValue)" }

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(tone).ignoresSafeArea()
            VStack(spacing: 0) {
                Rectangle()
                    .fill(DesignTokens.Color.accent(tone))
                    .frame(height: 3)
                    .accessibilityHidden(true)
                header
                searchField
                list
            }
        }
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .overlay { sheet }
        .animation(.easeInOut(duration: 0.18), value: sheetMode)
    }

    private var header: some View {
        ScreenHeader(
            area: tone,
            title: area.screenTitle,
            titleSize: 28,
            subtitle: area.summary(open: viewModel.openCount, done: viewModel.doneCount),
            subtitleIdentifier: "\(idPrefix).summary",
            subtitleMonospaced: area == .work,
            bottomPadding: DesignTokens.Spacing.md
        ) {
            EmptyView()
        } trailing: {
            Button {
                sheetMode = .create
            } label: {
                Image(systemName: "plus")
                    .font(.system(size: 18, weight: .bold))
                    .frame(width: 40, height: 40)
                    .background(DesignTokens.Color.accent(tone))
                    .clipShape(Circle())
                    .foregroundStyle(.white)
            }
            .accessibilityIdentifier("\(idPrefix).add.button")
        }
    }

    private var searchField: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "magnifyingglass")
                .foregroundStyle(DesignTokens.Color.ink(tone).opacity(0.45))
                .font(.system(size: 14))
            TextField(area.searchPrompt, text: $viewModel.query)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(tone))
                .autocorrectionDisabled()
                .accessibilityIdentifier("\(idPrefix).search.field")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, DesignTokens.Spacing.sm)
        .background(DesignTokens.Color.card(tone))
        .clipShape(RoundedRectangle(cornerRadius: area.rowRadius, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: area.rowRadius, style: .continuous)
                .stroke(DesignTokens.Color.rule(tone), lineWidth: 1)
        )
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.bottom, DesignTokens.Spacing.md)
    }

    private var list: some View {
        List {
            if let message = viewModel.errorMessage {
                Text(message)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.destructive(tone))
                    .listRowBackground(Color.clear)
                    .accessibilityIdentifier("\(idPrefix).error")
            }
            if viewModel.items.isEmpty {
                emptyState
            } else {
                section(title: area == .work ? "OPEN" : "할 일", items: viewModel.openItems)
                section(title: area == .work ? "DONE" : "완료", items: viewModel.doneItems)
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.load() }
    }

    @ViewBuilder
    private var emptyState: some View {
        Group {
            if viewModel.isLoading {
                ProgressView().frame(maxWidth: .infinity)
            } else {
                Text("아직 할 일이 없습니다. 우상단 + 버튼으로 추가하세요.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .lineHeight(.multiple(factor: 1.625))
                    .foregroundStyle(DesignTokens.Color.ink(tone).opacity(0.55))
                    .accessibilityIdentifier("\(idPrefix).empty.state")
            }
        }
        .padding(.vertical, DesignTokens.Spacing.xl)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    @ViewBuilder
    private func section(title: String, items: [TodoItem]) -> some View {
        if !items.isEmpty {
            Section {
                ForEach(items) { item in
                    TodoRow(area: area, item: item) {
                        Task { await viewModel.toggleDone(item) }
                    }
                    .contentShape(Rectangle())
                    .onTapGesture { sheetMode = .edit(item) }
                    .listRowInsets(EdgeInsets(top: DesignTokens.Spacing.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: DesignTokens.Spacing.cardGap / 2, trailing: DesignTokens.Spacing.xl))
                    .listRowSeparator(.hidden)
                    .listRowBackground(Color.clear)
                    .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                        Button {
                            Task { await viewModel.delete(id: item.id) }
                        } label: {
                            Label("삭제", systemImage: "trash")
                        }
                        .tint(DesignTokens.Color.destructive(tone))
                    }
                }
            } header: {
                Text("\(title) · \(items.count)")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(tone).opacity(0.7))
                    .textCase(nil)
            }
        }
    }

    @ViewBuilder
    private var sheet: some View {
        if let mode = sheetMode {
            TodoEditSheet(
                area: area,
                mode: mode,
                onSave: { draft in
                    sheetMode = nil
                    Task {
                        switch mode {
                        case .create:
                            await viewModel.add(draft)
                        case .edit(let existing):
                            await viewModel.update(id: existing.id, draft: draft)
                        }
                    }
                },
                onCancel: { sheetMode = nil },
                onDelete: {
                    guard case let .edit(existing) = mode else { return }
                    sheetMode = nil
                    Task { await viewModel.delete(id: existing.id) }
                }
            )
            .transition(.move(edge: .bottom).combined(with: .opacity))
        }
    }
}

private struct TodoRow: View {
    let area: TodoArea
    let item: TodoItem
    let onToggle: () -> Void

    private var tone: DesignTokens.Area { area.designArea }

    var body: some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
            Button(action: onToggle) {
                Image(systemName: item.isDone ? "checkmark.square.fill" : "square")
                    .font(.system(size: 18))
                    .foregroundStyle(DesignTokens.Color.accent(tone))
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("todos.\(area.rawValue).row.\(item.id).toggle")
            VStack(alignment: .leading, spacing: 6) {
                Text(item.title)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .medium))
                    .lineHeight(.multiple(factor: 1.375))
                    .foregroundStyle(DesignTokens.Color.ink(tone))
                    .strikethrough(item.isDone)
                    .accessibilityIdentifier("todos.\(area.rawValue).row.\(item.id)")
                if !meta.isEmpty {
                    Text(meta)
                        .font(metaFont)
                        .foregroundStyle(DesignTokens.Color.ink(tone).opacity(0.55))
                        .accessibilityIdentifier("todos.\(area.rawValue).row.\(item.id).meta")
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(14)
        .background(DesignTokens.Color.card(tone).opacity(item.isDone ? 0.5 : 1))
        .clipShape(RoundedRectangle(cornerRadius: area.rowRadius, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: area.rowRadius, style: .continuous)
                .stroke(DesignTokens.Color.rule(tone), lineWidth: 1)
        )
        .opacity(item.isDone ? 0.6 : 1)
    }

    private var meta: String {
        var parts: [String] = []
        if let due = item.dueDate { parts.append(due) }
        if let priority = item.priority {
            parts.append(area == .work ? "P\(workRank(priority))" : priority.displayName)
        }
        if !item.memo.isEmpty { parts.append("메모: \(item.memo)") }
        return parts.joined(separator: " · ")
    }

    private func workRank(_ priority: TodoPriority) -> Int {
        switch priority {
        case .high: return 1
        case .normal: return 2
        case .low: return 3
        }
    }

    private var metaFont: Font {
        let base = DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)
        return area == .work ? base.monospaced() : base
    }
}
