import Foundation
import Combine

@MainActor
class AuthViewModel: ObservableObject {
    @Published var email = ""
    @Published var password = ""
    @Published var confirmPassword = ""
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var isAuthenticated = false

    let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    func login() async {
        isLoading = true
        errorMessage = nil

        do {
            try await authManager.login(email: email, password: password)
            isAuthenticated = true
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    func register() async {
        isLoading = true
        errorMessage = nil

        guard password == confirmPassword else {
            errorMessage = "Passwords do not match"
            isLoading = false
            return
        }

        do {
            try await authManager.register(email: email, password: password)
            isAuthenticated = true
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    var isValid: Bool {
        !email.isEmpty && !password.isEmpty && email.contains("@")
    }

    var isValidForRegistration: Bool {
        isValid && !confirmPassword.isEmpty && password == confirmPassword
    }
}