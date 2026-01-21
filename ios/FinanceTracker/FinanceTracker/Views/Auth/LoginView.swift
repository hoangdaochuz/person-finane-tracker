import SwiftUI

struct LoginView: View {
    @StateObject private var viewModel: AuthViewModel
    @Environment(\.dismiss) var dismiss
    @State private var showingRegister = false

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
                Text("Finance Tracker")
                    .font(Typography.largeTitle)
                    .foregroundColor(ColorPalette.textPrimary)

                Text("Track your finances with ease")
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

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(Typography.caption)
                        .foregroundColor(ColorPalette.danger)
                }

                GradientButton(
                    title: "Sign In",
                    action: {
                        Task {
                            await viewModel.login()
                            if viewModel.isAuthenticated {
                                dismiss()
                            }
                        }
                    },
                    isDisabled: !viewModel.isValid || viewModel.isLoading
                )
            }
            .padding()

            HStack(spacing: 4) {
                Text("Don't have an account?")
                    .font(Typography.subheadline)
                    .foregroundColor(ColorPalette.textSecondary)

                Button("Sign Up") {
                    showingRegister = true
                }
                .font(Typography.subheadlineMedium)
                .foregroundColor(ColorPalette.primaryIndigo)
            }

            Spacer()
        }
        .padding()
        .sheet(isPresented: $showingRegister) {
            RegisterView(authManager: viewModel.authManager)
        }
    }
}

struct LoginView_Previews: PreviewProvider {
    static var previews: some View {
        // Preview requires AuthManager which depends on API service
        Text("Login Preview")
    }
}