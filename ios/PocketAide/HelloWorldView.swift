import DesignSystem
import SwiftUI

struct HelloWorldView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.xxl) {
                header
                statusSection
                paletteGrid
            }
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.vertical, DesignTokens.Spacing.xxl)
        }
        .background(DesignTokens.Color.surface(.system))
        .task { await auth.refreshHealth() }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            Text("Hello, pocket-aide")
                .font(.system(size: DesignTokens.Typography.h1, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(.system))
            Text("design tokens online")
                .font(.system(size: DesignTokens.Typography.bodyLg))
                .foregroundStyle(DesignTokens.Color.accent(.aiChat))
        }
    }

    private var statusSection: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Circle()
                    .fill(auth.healthOK ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.accent(.personal))
                    .frame(width: 10, height: 10)
                Text(auth.healthOK ? "Health: OK" : "Health: \(auth.healthError ?? "checking…")")
                    .font(.system(size: DesignTokens.Typography.captionSm))
                    .foregroundStyle(DesignTokens.Color.ink(.system))
                    .accessibilityIdentifier("HealthStatus")
            }

            if let user = auth.me {
                Text("Signed in as sub=\(user.sub) (id=\(user.id))")
                    .font(.system(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(.aiChat))
                    .accessibilityIdentifier("SignedInLabel")
                Button("Sign out", action: auth.signOut)
                    .buttonStyle(.bordered)
                    .accessibilityIdentifier("SignOutButton")
            } else if let error = auth.meError {
                Text(error)
                    .font(.system(size: DesignTokens.Typography.captionSm))
                    .foregroundStyle(DesignTokens.Color.accent(.personal))
                    .accessibilityIdentifier("MeErrorLabel")
                HStack(spacing: DesignTokens.Spacing.sm) {
                    Button("Retry") {
                        Task { await auth.refreshMe() }
                    }
                    .buttonStyle(.bordered)
                    .accessibilityIdentifier("RetryMeButton")
                    Button("Sign out", action: auth.signOut)
                        .buttonStyle(.bordered)
                        .accessibilityIdentifier("SignOutButton")
                }
            } else {
                Text("Loading account…")
                    .font(.system(size: DesignTokens.Typography.captionSm))
                    .foregroundStyle(DesignTokens.Color.ink(.aiChat).opacity(0.6))
            }
        }
        .padding(DesignTokens.Spacing.lg)
        .background(DesignTokens.Color.soft(.aiChat))
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card))
    }

    private var paletteGrid: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            Text("Area palette")
                .font(.system(size: DesignTokens.Typography.titleMd, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(.system))

            LazyVGrid(columns: [GridItem(.adaptive(minimum: 150), spacing: DesignTokens.Spacing.md)], spacing: DesignTokens.Spacing.md) {
                ForEach(DesignTokens.Area.allCases, id: \.self) { area in
                    AreaChip(area: area)
                }
            }
        }
    }
}

private struct AreaChip: View {
    let area: DesignTokens.Area

    var body: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            Text(area.label)
                .font(.system(size: DesignTokens.Typography.captionXs, weight: .bold))
                .tracking(2)
                .textCase(.uppercase)
                .foregroundStyle(DesignTokens.Color.ink(area))

            HStack(spacing: 6) {
                swatch(DesignTokens.Color.surface(area))
                swatch(DesignTokens.Color.soft(area))
                swatch(DesignTokens.Color.accent(area))
                swatch(DesignTokens.Color.rule(area))
                swatch(DesignTokens.Color.ink(area))
            }
        }
        .padding(DesignTokens.Spacing.md)
        .background(DesignTokens.Color.surface(area))
        .overlay(
            RoundedRectangle(cornerRadius: DesignTokens.Radius.card)
                .stroke(DesignTokens.Color.rule(area), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card))
        .accessibilityIdentifier("AreaChip-\(area.rawValue)")
    }

    private func swatch(_ color: Color) -> some View {
        RoundedRectangle(cornerRadius: 6)
            .fill(color)
            .frame(width: 22, height: 22)
            .overlay(RoundedRectangle(cornerRadius: 6).stroke(Color.black.opacity(0.06), lineWidth: 0.5))
    }
}

#Preview {
    HelloWorldView()
        .environmentObject(AppAuthCoordinator())
}
