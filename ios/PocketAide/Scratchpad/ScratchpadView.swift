import DesignSystem
import PocketAideAPI
import SwiftUI

extension ScratchpadMoveTarget {
    var designArea: DesignTokens.Area {
        switch self {
        case .personal: return .personal
        case .work: return .work
        case .affirmation: return .affirmations
        case .routine: return .routines
        }
    }
}

struct ScratchpadTab: View {
    @ObservedObject var viewModel: ScratchpadViewModel

    var body: some View { ScratchpadView(viewModel: viewModel) }
}

struct ScratchpadView: View {
    private enum SheetMode: Equatable {
        case add
        case priority(Affirmation)
    }

    @EnvironmentObject private var auth: AppAuthCoordinator
    @ObservedObject var viewModel: ScratchpadViewModel
    @State private var sheetMode: SheetMode?

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.locale = Locale(identifier: "ko_KR")
        f.dateFormat = "HH:mm"
        return f
    }()

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(.scratchpad).ignoresSafeArea()
            VStack(spacing: 0) {
                header
                addButton
                list
            }
        }
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .accessibilityIdentifier("scratchpad.screen")
        .overlay { sheet }
        .animation(.easeInOut(duration: 0.18), value: sheetMode)
    }

    private var header: some View {
        ScreenHeader(area: .scratchpad, title: "임시 공간", subtitle: "캡처 부담 없이 일단 던져두는 곳") {
            VStack(alignment: .trailing, spacing: 2) {
                Text("\(viewModel.unclassifiedCount)")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.h1, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.accent(.scratchpad))
                    .accessibilityIdentifier("scratchpad.badge")
                Text("분류되지 않은 메모")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))
            }
        }
    }

    private var addButton: some View {
        Button {
            sheetMode = .add
        } label: {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Image(systemName: "plus")
                    .font(.system(size: 13, weight: .bold))
                Text("새 메모")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                Spacer(minLength: 0)
            }
            .padding(.horizontal, DesignTokens.Spacing.md)
            .padding(.vertical, 12)
            .background(DesignTokens.Color.card(.scratchpad))
            .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                    .stroke(DesignTokens.Color.rule(.scratchpad), lineWidth: 1)
            )
            .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
        }
        .buttonStyle(.plain)
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.bottom, DesignTokens.Spacing.sm)
        .accessibilityIdentifier("scratchpad.add.button")
    }

    private var list: some View {
        List {
            if let message = viewModel.errorMessage {
                Text(message)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.destructive(.scratchpad))
                    .listRowBackground(Color.clear)
                    .accessibilityIdentifier("scratchpad.error")
            }
            if viewModel.items.isEmpty {
                emptyState
            } else {
                ForEach(viewModel.sections) { section in
                    Section {
                        ForEach(section.items) { item in
                            row(for: item)
                        }
                    } header: {
                        HStack(spacing: DesignTokens.Spacing.sm) {
                            Text(section.title)
                                .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold))
                                .tracking(2.1)
                                .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.7))
                            Rectangle()
                                .fill(DesignTokens.Color.rule(.scratchpad))
                                .frame(height: 1)
                            Text("\(section.items.count) ITEMS")
                                .font(DesignTokens.Typography.font(size: 10.5))
                                .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.5))
                        }
                        .textCase(nil)
                    }
                }
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
                Text("분류할 메모가 없습니다. 떠오르는 대로 새 메모에 던져두세요.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .lineHeight(.multiple(factor: 1.625))
                    .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))
                    .accessibilityIdentifier("scratchpad.empty.state")
            }
        }
        .padding(.vertical, DesignTokens.Spacing.xl)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private func row(for item: ScratchpadItem) -> some View {
        ScratchpadCard(item: item, time: timeLabel(item)) { target in
            Task {
                if let affirmation = await viewModel.move(item, to: target) {
                    sheetMode = .priority(affirmation)
                }
            }
        }
        .listRowInsets(EdgeInsets(top: 4, leading: DesignTokens.Spacing.xl, bottom: 4, trailing: DesignTokens.Spacing.xl))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .swipeActions(edge: .trailing, allowsFullSwipe: false) {
            Button {
                Task { await viewModel.delete(id: item.id) }
            } label: {
                Label("삭제", systemImage: "trash")
            }
            .tint(DesignTokens.Color.destructive(.scratchpad))
        }
    }

    private func timeLabel(_ item: ScratchpadItem) -> String {
        Self.timeFormatter.string(from: Date(timeIntervalSince1970: TimeInterval(item.capturedAt)))
    }

    @ViewBuilder
    private var sheet: some View {
        switch sheetMode {
        case .add:
            ScratchpadAddSheet(
                onSave: { text in
                    sheetMode = nil
                    Task { await viewModel.add(text: text) }
                },
                onCancel: { sheetMode = nil }
            )
            .transition(.move(edge: .bottom).combined(with: .opacity))
        case .priority(let affirmation):
            PriorityEditSheet(
                mode: .edit(affirmation),
                onSave: { text, priority in
                    sheetMode = nil
                    Task { await viewModel.setPriority(of: affirmation, text: text, priority: priority) }
                },
                onCancel: { sheetMode = nil }
            )
            .transition(.move(edge: .bottom).combined(with: .opacity))
        case nil:
            EmptyView()
        }
    }
}

private struct ScratchpadCard: View {
    let item: ScratchpadItem
    let time: String
    let onMove: (ScratchpadMoveTarget) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .top, spacing: 12) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55))
                        .frame(width: 6, height: 6)
                    Text(item.source.displayName)
                        .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold))
                        .tracking(0.5)
                        .textCase(.uppercase)
                        .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))
                }
                Spacer(minLength: 0)
                Text(time)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))
            }
            .accessibilityElement(children: .ignore)
            .accessibilityLabel("\(item.source.displayName) · \(time)")
            .accessibilityAddTraits(.isStaticText)
            .accessibilityIdentifier("scratchpad.row.\(item.id).meta")
            .padding(.bottom, 6)
            Text(item.text)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                .lineHeight(.multiple(factor: 1.625))
                .foregroundStyle(DesignTokens.Color.ink(.scratchpad))
                .fixedSize(horizontal: false, vertical: true)
            HStack(spacing: 6) {
                ForEach(ScratchpadMoveTarget.allCases, id: \.self) { target in
                    Button {
                        onMove(target)
                    } label: {
                        Text(target.chipLabel)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .medium))
                            .padding(.horizontal, 10)
                            .padding(.vertical, DesignTokens.Spacing.xs)
                            .overlay(Capsule().stroke(DesignTokens.Color.accent(target.designArea), lineWidth: 1))
                            .foregroundStyle(DesignTokens.Color.accent(target.designArea))
                    }
                    .buttonStyle(.plain)
                    .accessibilityIdentifier("scratchpad.row.\(item.id).move.\(target.rawValue)")
                }
            }
            .padding(.top, DesignTokens.Spacing.md)
        }
        .padding(DesignTokens.Spacing.lg)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(.scratchpad))
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)
                .stroke(DesignTokens.Color.rule(.scratchpad), lineWidth: 1)
        )
        .accessibilityIdentifier("scratchpad.row.\(item.id)")
    }
}
