import DesignSystem
import PocketAideAPI
import SwiftUI

struct ExternalNotificationsView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator
    @Environment(\.scenePhase) private var scenePhase
    @StateObject private var viewModel = ExternalNotificationsViewModel(api: nil, onMarkedRead: WidgetRefresher.reloadAll)
    @State private var selected: ExternalNotification?

    let reloadToken: Int
    @Binding var openedNotificationID: Int64?

    private static let area: DesignTokens.Area = .prMonitor

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(Self.area).ignoresSafeArea()
            VStack(spacing: 0) {
                header
                list
            }
        }
        .task(id: reloadToken) {
            if viewModel.api == nil, let api = auth.api {
                viewModel.replaceAPI(api)
            }
            await viewModel.load()
        }
        .onChange(of: scenePhase) { _, phase in
            if phase == .active {
                Task { await viewModel.load() }
            }
        }
        .onChange(of: openedNotificationID, initial: true) { _, id in
            guard let id else { return }
            openedNotificationID = nil
            Task { await open(id) }
        }
        .onChange(of: selected?.id) { _, _ in
            guard let item = selected else { return }
            Task { await viewModel.markRead(item) }
        }
        .sheet(item: $selected) { item in
            ExternalNotificationDetailSheet(item: item) { selected = nil }
                .presentationDetents([.large])
        }
        .accessibilityIdentifier("notifications.screen")
    }

    private func open(_ id: Int64) async {
        if viewModel.api == nil, let api = auth.api {
            viewModel.replaceAPI(api)
        }
        if let item = await viewModel.notification(id: id) {
            selected = item
        }
    }

    private var header: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(DesignTokens.Color.accent(Self.area))
                        .frame(width: 8, height: 8)
                    Text("NOTIFY · LOG")
                        .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                        .tracking(2.4)
                        .foregroundStyle(DesignTokens.Color.accent(Self.area))
                }
                Text("알림 기록")
                    .font(DesignTokens.Typography.font(size: 26, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                    .accessibilityIdentifier("notifications.title")
                Text("자동화가 보낸 알림 · 보낸 호출자와 요청 주체")
                    .font(DesignTokens.Typography.font(size: 11.5))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
            }
            Spacer()
            if viewModel.unreadCount > 0 {
                VStack(spacing: 2) {
                    Text("\(viewModel.unreadCount)")
                        .font(.system(size: 15, weight: .bold, design: .monospaced))
                        .accessibilityIdentifier("notifications.unread.count")
                    Text("새 알림")
                        .font(.system(size: 8, weight: .bold, design: .monospaced))
                }
                .foregroundStyle(.white)
                .padding(.horizontal, 10)
                .padding(.vertical, DesignTokens.Spacing.xs)
                .background(DesignTokens.Color.accent(Self.area))
                .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
            }
        }
        .padding(.horizontal, DesignTokens.Spacing.xl)
        .padding(.top, DesignTokens.Spacing.md)
        .padding(.bottom, DesignTokens.Spacing.md)
    }

    private var list: some View {
        List {
            if let message = viewModel.errorMessage {
                errorBanner(message)
            }
            if viewModel.items.isEmpty {
                emptyState
            } else {
                sectionHeader
                ForEach(viewModel.items) { item in
                    row(for: item)
                }
                if viewModel.hasMore {
                    loadingMore
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
        .refreshable { await viewModel.load() }
    }

    private var sectionHeader: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Rectangle()
                .fill(DesignTokens.Color.accent(Self.area))
                .frame(width: 4, height: 12)
            Text("최근 알림")
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold, design: .monospaced))
                .foregroundStyle(DesignTokens.Color.accent(Self.area))
            Rectangle()
                .fill(DesignTokens.Color.rule(Self.area))
                .frame(height: 1)
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private func errorBanner(_ message: String) -> some View {
        HStack(alignment: .top, spacing: DesignTokens.Spacing.sm) {
            Image(systemName: "exclamationmark.circle")
                .foregroundStyle(DesignTokens.StatusColor.failure)
            VStack(alignment: .leading, spacing: 2) {
                Text("알림 기록을 불러오지 못했습니다")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area))
                Text(message)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                    .lineLimit(2)
            }
            Spacer(minLength: 0)
            Button("다시 시도") {
                Task { await viewModel.load() }
            }
            .buttonStyle(.borderless)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
            .foregroundStyle(DesignTokens.Color.accent(Self.area))
            .accessibilityIdentifier("notifications.error.retry")
        }
        .padding(.horizontal, DesignTokens.Spacing.md)
        .padding(.vertical, 10)
        .background(DesignTokens.Color.soft(Self.area))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.accent(Self.area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("notifications.error.banner")
    }

    @ViewBuilder
    private var emptyState: some View {
        Group {
            if viewModel.isLoading || (!viewModel.hasLoaded && viewModel.errorMessage == nil) {
                ProgressView().frame(maxWidth: .infinity)
            } else if viewModel.errorMessage == nil {
                VStack(spacing: 0) {
                    Image(systemName: "bell")
                        .font(.system(size: 22, weight: .medium))
                        .foregroundStyle(DesignTokens.Color.accent(Self.area))
                        .frame(width: 56, height: 56)
                        .background(Circle().fill(DesignTokens.Color.soft(Self.area)))
                    Text("받은 알림이 없어요.")
                        .font(DesignTokens.Typography.font(size: 15, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area))
                        .padding(.top, DesignTokens.Spacing.lg)
                        .accessibilityIdentifier("notifications.empty")
                    Text("등록한 호출자가 보낸 알림이 여기에 남습니다.")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                        .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.55))
                        .padding(.top, 6)
                }
                .frame(maxWidth: .infinity)
                .padding(.top, 96)
            }
        }
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
    }

    private var loadingMore: some View {
        Text("이전 기록 불러오는 중…")
            .font(.system(size: 11, design: .monospaced))
            .foregroundStyle(DesignTokens.Color.ink(Self.area).opacity(0.45))
            .frame(maxWidth: .infinity)
            .padding(.vertical, DesignTokens.Spacing.sm)
            .listRowSeparator(.hidden)
            .listRowBackground(Color.clear)
            .task(id: viewModel.items.last?.id) { await viewModel.loadMore() }
    }

    private func row(for item: ExternalNotification) -> some View {
        Button {
            selected = item
        } label: {
            ExternalNotificationRowCard(item: item, area: Self.area)
        }
        .buttonStyle(.plain)
        .listRowSeparator(.hidden)
        .listRowBackground(Color.clear)
        .listRowInsets(EdgeInsets(top: 5, leading: DesignTokens.Spacing.xl, bottom: 5, trailing: DesignTokens.Spacing.xl))
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("notifications.row.\(item.id)")
    }
}

private struct ExternalNotificationRowCard: View {
    let item: ExternalNotification
    let area: DesignTokens.Area

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Text(ExternalNotificationFormatting.callerLabel(item))
                    .font(.system(size: 10.5, weight: .bold, design: .monospaced))
                    .lineLimit(1)
                    .padding(.horizontal, DesignTokens.Spacing.sm)
                    .padding(.vertical, 2)
                    .background(DesignTokens.Color.soft(area))
                    .foregroundStyle(DesignTokens.StatusColor.arrivalGlow)
                    .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
                Spacer(minLength: 0)
                if item.isUnread {
                    Text("새 알림")
                        .font(.system(size: 10, weight: .bold, design: .monospaced))
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(DesignTokens.Color.accent(area))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 2, style: .continuous))
                        .accessibilityIdentifier("notifications.row.unread")
                }
                Text(ExternalNotificationFormatting.time(item))
                    .font(.system(size: 10.5, design: .monospaced))
                    .monospacedDigit()
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.45))
            }
            Text(item.title)
                .font(DesignTokens.Typography.font(size: 13.5, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .lineLimit(2)
            if !item.body.isEmpty {
                Text(item.body)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                    .lineLimit(3)
            }
            HStack(alignment: .firstTextBaseline, spacing: DesignTokens.Spacing.sm) {
                Text(ExternalNotificationFormatting.requesterLabel(item))
                    .font(.system(size: 10.5, design: .monospaced))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(item.requester.isEmpty ? 0.45 : 0.55))
                    .lineLimit(2)
                    .frame(maxWidth: .infinity, alignment: .leading)
                Text(ExternalNotificationFormatting.deliveryLabel(item))
                    .font(.system(size: 10.5, design: .monospaced))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.45))
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, DesignTokens.Spacing.md)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(DesignTokens.Color.card(area))
        .overlay(alignment: .leading) {
            if item.isUnread {
                Rectangle()
                    .fill(DesignTokens.Color.accent(area))
                    .frame(width: 3)
            }
        }
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}
