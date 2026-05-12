import DesignSystem
import PocketAideAPI
import PocketAideAuth
import PocketAideStorage
import SwiftUI

struct HelloWorldView: View {
    @StateObject private var model = HelloWorldViewModel()

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.xxl) {
                header
                signInSection
                paletteGrid
            }
            .padding(.horizontal, DesignTokens.Spacing.xl)
            .padding(.vertical, DesignTokens.Spacing.xxl)
        }
        .background(DesignTokens.Color.surface(.system))
        .task { await model.checkHealth() }
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

    private var signInSection: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
            HStack(spacing: DesignTokens.Spacing.sm) {
                Circle()
                    .fill(model.healthOK ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.accent(.personal))
                    .frame(width: 10, height: 10)
                Text(model.healthOK ? "Health: OK" : "Health: \(model.healthError ?? "checking…")")
                    .font(.system(size: DesignTokens.Typography.captionSm))
                    .foregroundStyle(DesignTokens.Color.ink(.system))
                    .accessibilityIdentifier("HealthStatus")
            }

            if let user = model.signedInUser {
                Text("Signed in as sub=\(user.sub) (id=\(user.id))")
                    .font(.system(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(.aiChat))
                    .accessibilityIdentifier("SignedInLabel")
                Button("Sign out", action: model.signOut)
                    .buttonStyle(.bordered)
            } else {
                Button {
                    Task { await model.signIn() }
                } label: {
                    Text(model.isSigningIn ? "Signing in…" : "Sign in with OIDC")
                        .font(.system(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                        .padding(.horizontal, DesignTokens.Spacing.lg)
                        .padding(.vertical, DesignTokens.Spacing.md)
                }
                .buttonStyle(.borderedProminent)
                .tint(DesignTokens.Color.accent(.aiChat))
                .disabled(model.isSigningIn)
                .accessibilityIdentifier("SignInButton")

                if let err = model.signInError {
                    Text(err)
                        .font(.system(size: DesignTokens.Typography.captionXs))
                        .foregroundStyle(DesignTokens.Color.accent(.personal))
                }
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

@MainActor
final class HelloWorldViewModel: ObservableObject {
    @Published var healthOK = false
    @Published var healthError: String?
    @Published var signedInUser: MeResponse?
    @Published var isSigningIn = false
    @Published var signInError: String?

    private let api: APIClient?
    private let oidc: OIDCClient?
    private let tokenStore: TokenStoring

    init() {
        let store = KeychainTokenStore(accessGroup: nil)
        self.tokenStore = store
        if let api = try? APIClient.fromBundle(.main, tokenStore: store) {
            self.api = api
            self.oidc = OIDCClient(api: api, tokenStore: store)
        } else {
            self.api = nil
            self.oidc = nil
            self.healthError = "BackendBaseURL missing"
        }
    }

    func checkHealth() async {
        guard let api else { return }
        do {
            try await api.health()
            healthOK = true
            await refreshMe()
        } catch {
            healthOK = false
            healthError = String(describing: error)
        }
    }

    func refreshMe() async {
        guard let api else { return }
        do {
            signedInUser = try await api.me()
        } catch {
            signedInUser = nil
        }
    }

    func signIn() async {
        guard let oidc else { return }
        isSigningIn = true
        signInError = nil
        defer { isSigningIn = false }
        do {
            _ = try await oidc.signIn()
            await refreshMe()
        } catch {
            signInError = String(describing: error)
        }
    }

    func signOut() {
        try? oidc?.signOut()
        signedInUser = nil
    }
}

#Preview {
    HelloWorldView()
}
