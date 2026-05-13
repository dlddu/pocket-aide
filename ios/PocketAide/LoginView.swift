import DesignSystem
import SwiftUI

struct LoginView: View {
    @EnvironmentObject private var auth: AppAuthCoordinator

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(.aiChat)
                .ignoresSafeArea()

            VStack(spacing: DesignTokens.Spacing.xxl) {
                Spacer()
                brand
                Spacer()
                signInBlock
                healthFooter
            }
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.vertical, DesignTokens.Spacing.xxl)
        }
    }

    private var brand: some View {
        VStack(spacing: DesignTokens.Spacing.sm) {
            Text("PocketAide")
                .font(.system(size: DesignTokens.Typography.h1, weight: .bold))
                .foregroundStyle(DesignTokens.Color.ink(.aiChat))
                .accessibilityIdentifier("BrandTitle")
            Text("당신의 작은 비서")
                .font(.system(size: DesignTokens.Typography.bodyLg))
                .foregroundStyle(DesignTokens.Color.accent(.aiChat))
                .accessibilityIdentifier("BrandTagline")
        }
    }

    private var signInBlock: some View {
        VStack(spacing: DesignTokens.Spacing.md) {
            Button {
                Task { await auth.signIn() }
            } label: {
                Text(auth.isSigningIn ? "Signing in…" : "Sign in with OIDC")
                    .font(.system(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, DesignTokens.Spacing.md)
            }
            .buttonStyle(.borderedProminent)
            .tint(DesignTokens.Color.accent(.aiChat))
            .disabled(auth.isSigningIn || auth.oidc == nil)
            .accessibilityIdentifier("SignInButton")

            if let err = auth.signInError {
                Text(err)
                    .font(.system(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.accent(.personal))
                    .multilineTextAlignment(.center)
                    .accessibilityIdentifier("SignInError")
            }
        }
    }

    private var healthFooter: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            Circle()
                .fill(auth.healthOK ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.accent(.personal))
                .frame(width: DesignTokens.Spacing.sm, height: DesignTokens.Spacing.sm)
            Text(auth.healthOK ? "Health: OK" : "Health: \(auth.healthError ?? "checking…")")
                .font(.system(size: DesignTokens.Typography.caption2xs))
                .foregroundStyle(DesignTokens.Color.ink(.aiChat).opacity(0.7))
                .accessibilityIdentifier("HealthStatus")
        }
    }
}
