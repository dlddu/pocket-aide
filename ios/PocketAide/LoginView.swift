import DesignSystem
import SwiftUI

struct LoginView: View {
    @EnvironmentObject private var session: AppSession

    private let area: DesignTokens.Area = .aiChat

    var body: some View {
        ZStack {
            DesignTokens.Color.surface(area)
                .ignoresSafeArea()

            VStack(spacing: DesignTokens.Spacing.xl) {
                Spacer()

                VStack(spacing: DesignTokens.Spacing.md) {
                    Text("PocketAide")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.h1,
                            weight: .bold
                        ))
                        .foregroundStyle(DesignTokens.Color.ink(area))

                    Text("당신의 손바닥 위 동반자")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.bodyLg
                        ))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.65))
                }

                Spacer()

                VStack(spacing: DesignTokens.Spacing.md) {
                    Button {
                        Task { await session.signIn() }
                    } label: {
                        HStack(spacing: DesignTokens.Spacing.sm) {
                            if session.isSigningIn {
                                ProgressView()
                                    .controlSize(.small)
                                    .tint(DesignTokens.Color.surface(area))
                            }
                            Text(session.isSigningIn ? "로그인 중…" : "로그인")
                                .font(DesignTokens.Typography.font(
                                    size: DesignTokens.Typography.bodyLg,
                                    weight: .bold
                                ))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, DesignTokens.Spacing.md)
                        .background(DesignTokens.Color.ink(area))
                        .foregroundStyle(DesignTokens.Color.surface(area))
                        .clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.chip))
                    }
                    .buttonStyle(.plain)
                    .disabled(session.isSigningIn)
                    .accessibilityIdentifier("login.signin")

                    Text("OIDC를 통해 로그인합니다")
                        .font(DesignTokens.Typography.font(
                            size: DesignTokens.Typography.captionSm
                        ))
                        .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.5))
                }
                .padding(.horizontal, DesignTokens.Spacing.xxl)
                .padding(.bottom, DesignTokens.Spacing.xxl)
            }
        }
        .dsToast(
            area: area,
            message: Binding(
                get: { session.signInError },
                set: { session.signInError = $0 }
            )
        )
    }
}

#Preview {
    LoginView()
        .environmentObject(AppSession())
}
