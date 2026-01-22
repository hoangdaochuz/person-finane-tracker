import Foundation
import Combine
import LocalAuthentication

class AuthManager: ObservableObject {
    @Published var isAuthenticated = false
    @Published var currentUser: User?

    private let apiService: APIServiceProtocol
    private let keychainManager: KeychainManager
    private let userDefaults: UserDefaults

    init(apiService: APIServiceProtocol, keychainManager: KeychainManager = KeychainManager(), userDefaults: UserDefaults = .standard) {
        self.apiService = apiService
        self.keychainManager = keychainManager
        self.userDefaults = userDefaults
        checkAuthStatus()
    }

    func login(email: String, password: String) async throws {
        let (user, token) = try await apiService.login(email: email, password: password)

        if let apiKey = user.apiKey {
            keychainManager.store(key: "finance_tracker_api_key", value: apiKey)
        }

        // Store JWT token in keychain
        keychainManager.store(key: "finance_tracker_jwt_token", value: token)

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
            saveUserData(user)
        }
    }

    func register(email: String, password: String, name: String? = nil) async throws {
        let (user, token) = try await apiService.register(email: email, password: password, name: name)

        if let apiKey = user.apiKey {
            keychainManager.store(key: "finance_tracker_api_key", value: apiKey)
        }

        // Store JWT token in keychain
        keychainManager.store(key: "finance_tracker_jwt_token", value: token)

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
            saveUserData(user)
        }
    }

    func logout() {
        keychainManager.delete(key: "finance_tracker_api_key")
        keychainManager.delete(key: "finance_tracker_jwt_token")
        clearUserData()
        currentUser = nil
        isAuthenticated = false
    }

    private func checkAuthStatus() {
        // Check if API key exists in keychain
        guard let apiKey = keychainManager.retrieve(key: "finance_tracker_api_key") else {
            isAuthenticated = false
            currentUser = nil
            return
        }

        // Restore user from UserDefaults if available
        if let email = userDefaults.string(forKey: "user_email"),
           let name = userDefaults.string(forKey: "user_name") {
            currentUser = User(
                email: email,
                name: name,
                apiKey: apiKey,
                isBiometricEnabled: userDefaults.bool(forKey: "user_biometric_enabled")
            )
            isAuthenticated = true
        } else {
            // API key exists but no user data - treat as unauthenticated
            isAuthenticated = false
            currentUser = nil
            keychainManager.delete(key: "finance_tracker_api_key")
        }
    }

    private func saveUserData(_ user: User) {
        userDefaults.set(user.email, forKey: "user_email")
        userDefaults.set(user.name, forKey: "user_name")
        userDefaults.set(user.isBiometricEnabled, forKey: "user_biometric_enabled")
    }

    private func clearUserData() {
        userDefaults.removeObject(forKey: "user_email")
        userDefaults.removeObject(forKey: "user_name")
        userDefaults.removeObject(forKey: "user_biometric_enabled")
    }

    // MARK: - Biometric Authentication Methods

    func authenticateWithFaceID(completion: @escaping (Bool) -> Void) {
        let context = LAContext()
        var error: NSError?

        if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
            context.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, localizedReason: "Authenticate to access Finance Tracker") { success, error in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        } else {
            // Fallback to passcode authentication
            completion(false)
        }
    }

    func isBiometricAuthenticationAvailable() -> Bool {
        let context = LAContext()
        var error: NSError?
        return context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)
    }

    func setSessionTimeout(_ timeout: TimeInterval) {
        userDefaults.set(timeout, forKey: "session_timeout")
        userDefaults.set(Date().timeIntervalSince1970, forKey: "last_activity_time")
    }

    func isSessionActive() -> Bool {
        guard userDefaults.object(forKey: "session_timeout") != nil,
              userDefaults.object(forKey: "last_activity_time") != nil else {
            return true // Default to active if no settings exist
        }

        let timeout = userDefaults.double(forKey: "session_timeout")
        let lastActivityTime = userDefaults.double(forKey: "last_activity_time")

        let elapsedTime = Date().timeIntervalSince1970 - lastActivityTime
        return elapsedTime < timeout
    }

    func saveAuthenticationState(_ state: Bool, for userId: String) {
        userDefaults.set(state, forKey: "auth_state_\(userId)")
    }

    func getAuthenticationState(for userId: String) -> Bool {
        return userDefaults.bool(forKey: "auth_state_\(userId)")
    }

    func clearAuthenticationState(for userId: String) {
        userDefaults.removeObject(forKey: "auth_state_\(userId)")
    }

    func isLockedOut() -> Bool {
        // Implement lockout logic after multiple failed attempts
        let failedAttempts = userDefaults.integer(forKey: "failed_auth_attempts")
        return failedAttempts >= 5
    }
}

protocol APIServiceProtocol {
    func login(email: String, password: String) async throws -> (User, String)
    func register(email: String, password: String, name: String?) async throws -> (User, String)
    func createTransaction(_ transaction: Transaction) async throws -> Transaction
    func getTransactions(page: Int, limit: Int) async throws -> [Transaction]
    func getAnalytics(period: TimePeriod) async throws -> Analytics
    func getSummary() async throws -> SummaryResponse
}