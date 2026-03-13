// LoginView.swift
// PocketAide

import SwiftUI

/// 로그인 화면. 서버 주소, 이메일, 비밀번호 입력 필드와 로그인 버튼을 제공합니다.
struct LoginView: View {

    @EnvironmentObject var authViewModel: AuthViewModel

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            Text("Pocket Aide")
                .font(.largeTitle)
                .fontWeight(.bold)

            VStack(spacing: 16) {
                TextField("Server Address (e.g. http://localhost:8080)", text: $authViewModel.serverAddress)
                    .textFieldStyle(.roundedBorder)
                    .autocapitalization(.none)
                    .disableAutocorrection(true)
                    .accessibilityIdentifier("server_address_field")

                TextField("Email", text: $authViewModel.email)
                    .textFieldStyle(.roundedBorder)
                    .autocapitalization(.none)
                    .keyboardType(.emailAddress)
                    .disableAutocorrection(true)
                    .accessibilityIdentifier("email_field")

                SecureField("Password", text: $authViewModel.password)
                    .textFieldStyle(.roundedBorder)
                    .accessibilityIdentifier("password_field")
            }
            .padding(.horizontal)

            if let errorMessage = authViewModel.errorMessage {
                Text(errorMessage)
                    .foregroundColor(.red)
                    .font(.caption)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)
                    .accessibilityIdentifier("login_error_message")
            }

            Button(action: {
                authViewModel.login()
            }) {
                if authViewModel.isLoading {
                    ProgressView()
                        .frame(maxWidth: .infinity)
                } else {
                    Text("Login")
                        .frame(maxWidth: .infinity)
                }
            }
            .buttonStyle(.borderedProminent)
            .padding(.horizontal)
            .disabled(authViewModel.isLoading)
            .accessibilityIdentifier("login_button")

            Spacer()
        }
        .accessibilityIdentifier("login_view")
    }
}
