import DesignSystem
import PocketAideAPI
import SwiftUI

struct AffirmationsView: View {
    private let area: DesignTokens.Area = .affirmations

    @StateObject private var viewModel = AffirmationsViewModel()
    @StateObject private var tts = TTSPlayer()

    @State private var sheetMode: PriorityEditSheet.Mode = .create
    @State private var isSheetPresented: Bool = false

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(area)
                .ignoresSafeArea()

            VStack(spacing: 0) {
                DSScreenHeader(
                    area: area,
                    areaLabel: "다짐",
                    title: "자주 읽어줘야 할 것",
                    titleFamily: .serif
                ) {
                    Button {
                        sheetMode = .create
                        isSheetPresented = true
                    } label: {
                        Image(systemName: "plus")
                            .font(.system(size: 14, weight: .bold))
                            .foregroundStyle(DesignTokens.Color.ink(area))
                            .frame(width: 36, height: 36)
                            .overlay(
                                Circle()
                                    .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
                            )
                    }
                    .buttonStyle(.plain)
                    .accessibilityIdentifier("affirmations.add")
                }

                ScrollView {
                    VStack(spacing: DesignTokens.Spacing.md) {
                        heroSection
                        listSection
                    }
                    .padding(.horizontal, DesignTokens.Spacing.xl)
                    .padding(.vertical, DesignTokens.Spacing.md)
                }
            }
        }
        .sheet(isPresented: $isSheetPresented) {
            PriorityEditSheet(
                mode: sheetMode,
                onSave: { text, priority in
                    await save(mode: sheetMode, text: text, priority: priority)
                },
                onCancel: dismissSheet
            )
            .presentationDetents([.medium, .large])
            .presentationDragIndicator(.visible)
            .presentationBackground(DesignTokens.Color.surface(area))
            .dsToast(area: area, message: $viewModel.loadError)
        }
        .dsToast(area: area, message: $viewModel.loadError)
        .task { await viewModel.load() }
    }

    private var heroSection: some View {
        Group {
            if let hero = viewModel.hero {
                DSCard(area: area, padding: .large) {
                    VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
                        HStack(spacing: DesignTokens.Spacing.xs) {
                            Circle()
                                .fill(DesignTokens.Color.accent(area))
                                .frame(width: 6, height: 6)
                            Text("오늘 회전")
                                .font(DesignTokens.Typography.font(
                                    size: DesignTokens.Typography.caption2xs,
                                    weight: .bold
                                ))
                                .tracking(2.4)
                                .textCase(.uppercase)
                                .foregroundStyle(DesignTokens.Color.accent(area))
                            Spacer()
                            Button {
                                viewModel.rotateHero()
                            } label: {
                                Image(systemName: "arrow.triangle.2.circlepath")
                                    .font(.system(size: 14, weight: .semibold))
                                    .foregroundStyle(DesignTokens.Color.accent(area))
                            }
                            .buttonStyle(.plain)
                            .accessibilityIdentifier("affirmations.rotate")
                        }

                        Text(hero.text)
                            .font(DesignTokens.Typography.font(
                                size: DesignTokens.Typography.h2,
                                weight: .medium,
                                family: .serif
                            ))
                            .foregroundStyle(DesignTokens.Color.ink(area))
                            .accessibilityIdentifier("affirmations.hero.text")

                        Divider().overlay(DesignTokens.Color.rule(area))

                        HStack {
                            HStack(spacing: DesignTokens.Spacing.xs) {
                                Text("우선순위")
                                    .font(DesignTokens.Typography.font(
                                        size: DesignTokens.Typography.captionXs
                                    ))
                                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                                Text(label(for: hero.priority))
                                    .font(DesignTokens.Typography.font(
                                        size: DesignTokens.Typography.captionXs,
                                        weight: .bold
                                    ))
                                    .foregroundStyle(DesignTokens.Color.ink(area))
                                PriorityDots(priority: hero.priority, layout: .horizontal)
                            }
                            Spacer()
                            Button {
                                tts.speak(hero.text)
                            } label: {
                                HStack(spacing: DesignTokens.Spacing.xs) {
                                    Image(systemName: "speaker.wave.2.fill")
                                        .font(.system(size: 12))
                                    Text("읽어주기")
                                        .font(DesignTokens.Typography.font(
                                            size: DesignTokens.Typography.captionSm,
                                            weight: .bold
                                        ))
                                }
                                .padding(.horizontal, DesignTokens.Spacing.md)
                                .padding(.vertical, DesignTokens.Spacing.sm)
                                .background(DesignTokens.Color.ink(area))
                                .foregroundStyle(DesignTokens.Color.surface(area))
                                .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.chip))
                            }
                            .buttonStyle(.plain)
                            .accessibilityIdentifier("affirmations.hero.tts")
                        }
                    }
                }
                .accessibilityIdentifier("affirmations.hero")
            } else if !viewModel.isLoading {
                emptyState
            }
        }
    }

    private var listSection: some View {
        Group {
            if !viewModel.items.isEmpty {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                    Text("전체 \(viewModel.items.count)개")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.body,
                            weight: .bold
                        ))
                        .foregroundStyle(DesignTokens.Color.ink(area))

                    LazyVStack(spacing: DesignTokens.Spacing.sm) {
                        ForEach(viewModel.items) { item in
                            listRow(for: item)
                        }
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func listRow(for item: Affirmation) -> some View {
        DSCard(area: area, padding: .small) {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                PriorityDots(priority: item.priority, layout: .vertical)
                    .padding(.top, 4)
                Text(item.text)
                    .font(DesignTokens.Typography.font(
                        size: DesignTokens.Typography.bodyLg,
                        family: .serif
                    ))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                    .frame(maxWidth: .infinity, alignment: .leading)
                Button {
                    tts.speak(item.text)
                } label: {
                    Image(systemName: "speaker.wave.2")
                        .font(.system(size: 16, weight: .regular))
                        .foregroundStyle(DesignTokens.Color.accent(area))
                }
                .buttonStyle(.plain)
            }
        }
        .accessibilityIdentifier("affirmations.row.\(item.id)")
        .contentShape(Rectangle())
        .onLongPressGesture {
            sheetMode = .edit(item)
            isSheetPresented = true
        }
    }

    private var emptyState: some View {
        VStack(spacing: DesignTokens.Spacing.sm) {
            Text("아직 등록된 다짐이 없습니다")
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.bodyLg
                ))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.65))
            Text("우측 상단 + 버튼으로 추가해보세요")
                .font(DesignTokens.Typography.font(
                    size: DesignTokens.Typography.captionSm
                ))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.45))
        }
        .padding(.vertical, DesignTokens.Spacing.xxl)
        .accessibilityIdentifier("affirmations.empty")
    }

    private func save(mode: PriorityEditSheet.Mode, text: String, priority: Priority) async {
        guard !text.isEmpty else { return }
        let success: Bool
        switch mode {
        case .create:
            success = await viewModel.add(text: text, priority: priority)
        case .edit(let existing):
            success = await viewModel.update(id: existing.id, text: text, priority: priority)
        }
        if success {
            dismissSheet()
        }
        // On failure the sheet stays open and the toast (attached to the sheet
        // content) surfaces the error so the user can correct and retry.
    }

    private func dismissSheet() {
        isSheetPresented = false
    }

    private func label(for priority: Priority) -> String {
        switch priority {
        case .high: return "높음"
        case .normal: return "보통"
        case .low: return "가끔"
        }
    }
}

#Preview {
    AffirmationsView()
}
