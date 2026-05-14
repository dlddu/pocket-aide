import DesignSystem
import PocketAideAPI
import SwiftUI

struct AffirmationsView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @StateObject private var viewModel: AffirmationsViewModel
    @StateObject private var tts = TTSPlayer()

    @State private var sheetMode: PriorityEditSheet.Mode?

    init() {
        let seed: UInt64? = {
            guard let raw = ProcessInfo.processInfo.environment["ROTATION_SEED"],
                  let parsed = UInt64(raw) else { return nil }
            return parsed
        }()
        _viewModel = StateObject(wrappedValue: AffirmationsViewModel(api: nil, rotationSeed: seed))
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

                ScrollView {
                    VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                        heroSection
                        listSection
                    }
                    .padding(.horizontal, DesignTokens.Spacing.xl)
                    .padding(.bottom, DesignTokens.Spacing.xxl)
                }
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
                    onCancel: { sheetMode = nil }
                )
                .transition(.move(edge: .bottom).combined(with: .opacity))
                // No outer accessibilityIdentifier here — iOS 26 cascades it
                // to every leaf inside the sheet, clobbering sheet.title,
                // sheet.text.field, sheet.save.button, sheet.cancel.button.
            }
        }
        .animation(.easeInOut(duration: 0.18), value: sheetMode)
    }

    @ViewBuilder
    private var heroSection: some View {
        if let hero = viewModel.heroItem {
            Card(area: .affirmations, padding: .large) {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
                    HStack(spacing: 6) {
                        Circle()
                            .fill(DesignTokens.Color.accent(.affirmations))
                            .frame(width: 6, height: 6)
                        AreaLabel(area: .affirmations, text: "오늘 회전")
                    }
                    Text(hero.text)
                        .font(DesignTokens.Typography.font(size: 22, weight: .medium, family: .serif))
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                        .fixedSize(horizontal: false, vertical: true)
                        .accessibilityIdentifier("affirmations.hero.text")
                    Divider().background(DesignTokens.Color.rule(.affirmations))
                    HStack {
                        HStack(spacing: 6) {
                            priorityDots(for: hero.priority)
                            Text("우선순위 \(hero.priority.displayName)")
                                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                                .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))
                        }
                        Spacer()
                        Button {
                            tts.speak(hero.text)
                        } label: {
                            HStack(spacing: 6) {
                                Image(systemName: "speaker.wave.2.fill")
                                    .font(.system(size: 12))
                                Text("읽어주기")
                                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                            }
                            .padding(.horizontal, 14)
                            .padding(.vertical, 8)
                            .background(Capsule().fill(DesignTokens.Color.ink(.affirmations)))
                            .foregroundStyle(DesignTokens.Color.surface(.affirmations))
                        }
                        .accessibilityIdentifier("affirmations.hero.tts")
                    }
                    HStack {
                        Spacer()
                        Button {
                            viewModel.rotateHero()
                        } label: {
                            Label("다른 다짐 보기", systemImage: "arrow.triangle.2.circlepath")
                                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                                .foregroundStyle(DesignTokens.Color.accent(.affirmations))
                        }
                        .accessibilityIdentifier("affirmations.hero.rotate")
                    }
                }
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
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))
                }
            }
            .accessibilityIdentifier("affirmations.empty.state")
        }
    }

    @ViewBuilder
    private var listSection: some View {
        if !viewModel.items.isEmpty {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                HStack {
                    Text("전체 \(viewModel.items.count)개")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    Spacer()
                }
                .accessibilityIdentifier("affirmations.list.header")

                VStack(spacing: DesignTokens.Spacing.sm) {
                    ForEach(viewModel.items) { item in
                        listRow(for: item)
                    }
                }
            }
        }
    }

    private func listRow(for item: Affirmation) -> some View {
        Card(area: .affirmations, padding: .small) {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                priorityDotsColumn(for: item.priority)
                Text(item.text)
                    .font(DesignTokens.Typography.font(size: 15.5, family: .serif))
                    .foregroundStyle(DesignTokens.Color.ink(.affirmations))
                    .frame(maxWidth: .infinity, alignment: .leading)
                Button {
                    tts.speak(item.text)
                } label: {
                    Image(systemName: "speaker.wave.2")
                        .font(.system(size: 16))
                        .foregroundStyle(DesignTokens.Color.accent(.affirmations))
                }
                .accessibilityIdentifier("affirmations.row.\(item.id).tts")
            }
        }
        .contentShape(Rectangle())
        .onLongPressGesture {
            sheetMode = .edit(item)
        }
        .accessibilityIdentifier("affirmations.row.\(item.id)")
    }

    private func priorityDots(for priority: AffirmationPriority) -> some View {
        HStack(spacing: 3) {
            ForEach(0..<3, id: \.self) { idx in
                Circle()
                    .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < dotsFilled(priority) ? 1 : 0.25))
                    .frame(width: 6, height: 6)
            }
        }
    }

    private func priorityDotsColumn(for priority: AffirmationPriority) -> some View {
        VStack(spacing: 2) {
            ForEach(0..<3, id: \.self) { idx in
                Circle()
                    .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < dotsFilled(priority) ? 1 : 0.25))
                    .frame(width: 6, height: 6)
            }
        }
        .padding(.top, 4)
    }

    private func dotsFilled(_ priority: AffirmationPriority) -> Int {
        switch priority {
        case .high: return 3
        case .normal: return 2
        case .low: return 1
        }
    }
}
