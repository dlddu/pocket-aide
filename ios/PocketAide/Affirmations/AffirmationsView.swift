import DesignSystem
import PocketAideAPI
import SwiftUI

struct AffirmationsView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @StateObject private var viewModel: AffirmationsViewModel

    @State private var sheetMode: PriorityEditSheet.Mode?

    init() {
        _viewModel = StateObject(wrappedValue: AffirmationsViewModel(api: nil))
    }

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(.affirmations).ignoresSafeArea()
            VStack(spacing: 0) {
                ScreenHeader(area: .affirmations, title: "자주 읽어줘야 할 것", titleFamily: .serif) {
                    Button {
                        sheetMode = .create
                    } label: {
                        Image(systemName: "plus")
                            .font(.system(size: 14, weight: .bold))
                            .frame(width: 36, height: 36)
                            .background(DesignTokens.Color.card(.affirmations).opacity(0.4))
                            .overlay(
                                Circle().stroke(DesignTokens.Color.rule(.affirmations), lineWidth: 1)
                            )
                            .clipShape(Circle())
                            .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    }
                    .accessibilityIdentifier("affirmations.add.button")
                }

                List {
                    Section {
                        heroSection
                            .padding(.bottom, DesignTokens.Spacing.xl)
                            .listRowInsets(EdgeInsets(
                                top: DesignTokens.Spacing.sm,
                                leading: DesignTokens.Spacing.xl,
                                bottom: 0,
                                trailing: DesignTokens.Spacing.xl
                            ))
                            .listRowSeparator(.hidden)
                            .listRowBackground(Color.clear)
                    }

                    listSection
                }
                .listStyle(.plain)
                .scrollContentBackground(.hidden)
                .background(DesignTokens.Color.surface(.affirmations))
            }
        }
        .task {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .overlay {
            if let mode = sheetMode {
                PriorityEditSheet(
                    mode: mode,
                    onSave: { text, priority in
                        let captured = mode
                        sheetMode = nil
                        Task {
                            switch captured {
                            case .create:
                                await viewModel.add(text: text, priority: priority)
                            case .edit(let existing):
                                await viewModel.update(id: existing.id, text: text, priority: priority)
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
                .ignoresSafeArea(.container, edges: .bottom)
                .transition(.move(edge: .bottom).combined(with: .opacity))
                // No outer accessibilityIdentifier here — iOS 26 cascades it
                // to every leaf inside the sheet, clobbering sheet.title,
                // sheet.text.field, sheet.save.button, sheet.cancel.button.
            }
        }
        .animation(.easeInOut(duration: 0.18), value: sheetMode)
        .toolbarVisibility(sheetMode == nil ? .automatic : .hidden, for: .tabBar)
    }

    @ViewBuilder
    private var listSection: some View {
        if !viewModel.items.isEmpty {
            Section {
                ForEach(viewModel.items) { item in
                    listRow(for: item)
                        .listRowInsets(EdgeInsets(
                            top: DesignTokens.Spacing.sm / 2,
                            leading: DesignTokens.Spacing.xl,
                            bottom: DesignTokens.Spacing.sm / 2,
                            trailing: DesignTokens.Spacing.xl
                        ))
                        .listRowSeparator(.hidden)
                        .listRowBackground(Color.clear)
                        .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                            Button {
                                Task { await viewModel.delete(id: item.id) }
                            } label: {
                                Label("삭제", systemImage: "trash")
                            }
                            .tint(DesignTokens.Color.destructive(.affirmations))
                            .accessibilityIdentifier("affirmations.row.\(item.id).delete")
                        }
                }
            } header: {
                HStack {
                    Text("전체 \(viewModel.items.count)개")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    Spacer()
                }
                .textCase(nil)
                .listRowInsets(EdgeInsets(
                    top: 0,
                    leading: DesignTokens.Spacing.xl,
                    bottom: DesignTokens.Spacing.sm,
                    trailing: DesignTokens.Spacing.xl
                ))
                .listRowBackground(Color.clear)
                .accessibilityIdentifier("affirmations.list.header")
            }

            Section {
                priorityLegend
                    .listRowInsets(EdgeInsets(
                        top: DesignTokens.Spacing.xl,
                        leading: DesignTokens.Spacing.xl,
                        bottom: DesignTokens.Spacing.xxl,
                        trailing: DesignTokens.Spacing.xl
                    ))
                    .listRowSeparator(.hidden)
                    .listRowBackground(Color.clear)
            }
        }
    }

    private var priorityLegend: some View {
        HStack(spacing: DesignTokens.Spacing.md) {
            legendItem(.high, "자주 노출")
            legendItem(.normal, "보통")
            legendItem(.low, "가끔")
            Spacer(minLength: 0)
        }
        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
        .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.7))
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(.affirmations).opacity(0.6))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .accessibilityIdentifier("affirmations.priority.legend")
    }

    private func legendItem(_ priority: AffirmationPriority, _ label: String) -> some View {
        HStack(spacing: 6) {
            PriorityDots.horizontal(for: priority)
            Text(label)
        }
    }

    private func listRow(for item: Affirmation) -> some View {
        Card(area: .affirmations, padding: .small) {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                PriorityDots.vertical(for: item.priority)
                Text(item.text)
                    .font(DesignTokens.Typography.font(size: 15.5, family: .serif))
                    .lineHeight(.multiple(factor: 1.375))
                    .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
        .contentShape(Rectangle())
        .onLongPressGesture {
            sheetMode = .edit(item)
        }
        .accessibilityIdentifier("affirmations.row.\(item.id)")
    }

}

private extension AffirmationsView {
    @ViewBuilder
    var heroSection: some View {
        if let hero = viewModel.heroItem {
            Card(area: .affirmations, padding: .large) {
                VStack(alignment: .leading, spacing: 0) {
                    HStack(spacing: DesignTokens.Spacing.sm) {
                        Circle()
                            .fill(DesignTokens.Color.accent(.affirmations))
                            .frame(width: 6, height: 6)
                        Text("오늘 회전")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))
                            .textCase(.uppercase)
                            .tracking(2.2)
                            .foregroundStyle(DesignTokens.Color.accent(.affirmations))
                    }
                    .padding(.bottom, DesignTokens.Spacing.md)
                    Text(hero.text)
                        .font(DesignTokens.Typography.font(size: 22, weight: .medium, family: .serif))
                        .lineHeight(.multiple(factor: 1.45))
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                        .fixedSize(horizontal: false, vertical: true)
                        .accessibilityIdentifier("affirmations.hero.text")
                    VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
                        let priorityValue = Text(hero.priority.displayName).bold().foregroundStyle(DesignTokens.Color.ink(.affirmations))
                        Text("우선순위 \(priorityValue)")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                            .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))
                        HStack {
                            Spacer()
                            Button {
                                viewModel.rotateHero()
                            } label: {
                                HStack(spacing: DesignTokens.Spacing.xs) {
                                    Image(systemName: "arrow.triangle.2.circlepath")
                                        .font(.system(size: 12, weight: .semibold))
                                    Text("다른 다짐 보기")
                                }
                                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                                .foregroundStyle(DesignTokens.Color.accent(.affirmations))
                            }
                            .accessibilityIdentifier("affirmations.hero.rotate")
                        }
                    }
                    .padding(.top, DesignTokens.Spacing.xl)
                    .overlay(alignment: .top) {
                        Rectangle()
                            .fill(DesignTokens.Color.rule(.affirmations))
                            .frame(height: 1)
                    }
                    .padding(.top, DesignTokens.Spacing.xl)
                }
            }
            .overlay {
                Text("\"")
                    .font(DesignTokens.Typography.font(size: 120, family: .serif))
                    .lineHeight(.multiple(factor: 1))
                    .foregroundStyle(DesignTokens.Color.accent(.affirmations).opacity(0.08))
                    .fixedSize()
                    .padding(.leading, -12)
                    .padding(.top, -12)
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    .clipShape(RoundedRectangle(cornerRadius: 28, style: .continuous))
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
            }
            .accessibilityIdentifier("affirmations.hero.card")
        } else if viewModel.isLoading {
            Card(area: .affirmations) {
                ProgressView()
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, DesignTokens.Spacing.xl)
            }
        } else {
            Card(area: .affirmations, padding: .large) {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                    Text("첫 다짐을 추가해 보세요")
                        .font(DesignTokens.Typography.font(size: 18, weight: .bold, family: .serif))
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    Text("우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                        .lineHeight(.multiple(factor: 1.625))
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))
                }
            }
            .accessibilityIdentifier("affirmations.empty.state")
        }
    }
}

private enum PriorityDots {
    static func filled(_ priority: AffirmationPriority) -> Int {
        switch priority {
        case .high: return 3
        case .normal: return 2
        case .low: return 1
        }
    }

    static func horizontal(for priority: AffirmationPriority) -> some View {
        HStack(spacing: 6) {
            ForEach(0..<3, id: \.self) { idx in
                Circle()
                    .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3))
                    .frame(width: 6, height: 6)
            }
        }
    }

    static func vertical(for priority: AffirmationPriority) -> some View {
        VStack(spacing: 2) {
            ForEach(0..<3, id: \.self) { idx in
                Circle()
                    .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3))
                    .frame(width: 6, height: 6)
            }
        }
        .padding(.top, 4)
    }
}
