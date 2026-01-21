import SwiftUI

struct RegisterView: View {
    @StateObject private var viewModel: AuthViewModel
    @Environment(\.dismiss) var dismiss

    init(authManager: AuthManager) {
        _viewModel = StateObject(wrappedValue: AuthViewModel(authManager: authManager))
    }

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "chart.line.uptrend.xyaxis.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(ColorPalette.primaryGradient)

            VStack(spacing: 8) {
                Text("Create Account")
                    .font(Typography.largeTitle)
                    .foregroundColor(ColorPalette.textPrimary)

                Text("Start tracking your finances")
                    .font(Typography.body)
                    .foregroundColor(ColorPalette.textSecondary)
            }

            Spacer()

            VStack(spacing: 16) {
                TextField("Email", text: $viewModel.email)
                    .textFieldStyle(.roundedBorder)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)

                SecureField("Password", text: $viewModel.password)
                    .textFieldStyle(.roundedBorder)

                SecureField("Confirm Password", text: $viewModel.confirmPassword)
                    .textFieldStyle(.roundedBorder)

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(Typography.caption)
                        .foregroundColor(ColorPalette.danger)
                }

                GradientButton(
                    title: "Sign Up",
                    action: {
                        Task {
                            await viewModel.register()
                            if viewModel.isAuthenticated {
                                dismiss()
                            }
                        }
                    },
                    isDisabled: !viewModel.isValidForRegistration || viewModel.isLoading
                )
            }
            .padding()

            Spacer()
        }
        .padding()
        .overlay(alignment: .topLeading) {
            Button {
                dismiss()
            } label: {
                Image(systemName: "xmark.circle.fill")
                    .font(.title3)
                    .foregroundColor(.gray)
            }
            .padding()
        }
    }
}

struct RegisterView_Previews: PreviewProvider {
    static var previews: some View {
        // Preview requires AuthManager which depends on API service
        Text("Register Preview")
    }
}