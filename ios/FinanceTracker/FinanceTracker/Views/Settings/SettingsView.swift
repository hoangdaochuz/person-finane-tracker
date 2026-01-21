import SwiftUI
import LocalAuthentication

struct SettingsView: View {
    @EnvironmentObject var authManager: AuthManager
    @State private var isBiometricEnabled = false
    @State private var showingBiometricError = false
    @State private var biometricError: Error?

    var body: some View {
        NavigationView {
            List {
                Section {
                    HStack(spacing: 16) {
                        Circle()
                            .fill(ColorPalette.primaryGradient)
                            .frame(width: 60, height: 60)

                        VStack(alignment: .leading, spacing: 4) {
                            Text(authManager.currentUser?.name ?? "User")
                                .font(Typography.bodyBold)
                                .foregroundColor(ColorPalette.textPrimary)

                            Text(authManager.currentUser?.email ?? "")
                                .font(Typography.subheadline)
                                .foregroundColor(ColorPalette.textSecondary)
                        }

                        Spacer()
                    }
                    .padding(.vertical, 8)
                }

                Section("Security") {
                    Toggle("Biometric Authentication", isOn: $isBiometricEnabled)
                        .onChange(of: isBiometricEnabled) { newValue in
                            enableBiometric(newValue)
                        }
                }

                Section("Notifications") {
                    HStack {
                        Text("Notification Access")
                            .font(Typography.body)
                            .foregroundColor(ColorPalette.textPrimary)

                        Spacer()

                        Image(systemName: "checkmark.circle.fill")
                            .foregroundColor(ColorPalette.success)
                    }

                    HStack {
                        Text("Connected Sources")
                            .font(Typography.body)
                            .foregroundColor(ColorPalette.textPrimary)

                        Spacer()

                        Text("3 sources")
                            .font(Typography.subheadline)
                            .foregroundColor(ColorPalette.textSecondary)
                    }
                }

                Section("About") {
                    HStack {
                        Text("Version")
                            .font(Typography.body)
                            .foregroundColor(ColorPalette.textPrimary)

                        Spacer()

                        Text("1.0.0")
                            .font(Typography.subheadline)
                            .foregroundColor(ColorPalette.textSecondary)
                    }
                }

                Section {
                    Button {
                        authManager.logout()
                    } label: {
                        HStack {
                            Spacer()
                            Text("Log Out")
                                .font(Typography.bodyMedium)
                                .foregroundColor(ColorPalette.danger)
                            Spacer()
                        }
                    }
                }
            }
            .navigationTitle("Settings")
            .alert("Biometric Error", isPresented: $showingBiometricError) {
                Button("OK", role: .cancel) {}
            } message: {
                if let error = biometricError {
                    Text(error.localizedDescription)
                }
            }
        }
    }

    private func enableBiometric(_ enable: Bool) {
        guard enable else {
            authManager.currentUser?.isBiometricEnabled = false
            return
        }

        let context = LAContext()
        var error: NSError?

        if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
            context.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, localizedReason: "Enable biometric authentication for Finance Tracker") { success, error in
                DispatchQueue.main.async {
                    if success {
                        authManager.currentUser?.isBiometricEnabled = true
                    } else {
                        isBiometricEnabled = false
                        if let error = error {
                            biometricError = error
                            showingBiometricError = true
                        }
                    }
                }
            }
        } else {
            isBiometricEnabled = false
            biometricError = error
            showingBiometricError = true
        }
    }
}

struct SettingsView_Previews: PreviewProvider {
    static var previews: some View {
        // Create a simple mock API service
        class MockAPIService: APIServiceProtocol {
            func login(email: String, password: String) async throws -> User {
                User(email: email, name: "Test User", isBiometricEnabled: false)
            }

            func register(email: String, password: String) async throws -> User {
                User(email: email, name: "Test User", isBiometricEnabled: false)
            }

            func createTransaction(_ transaction: Transaction) async throws -> Transaction {
                transaction
            }

            func getTransactions(page: Int, limit: Int) async throws -> [Transaction] {
                []
            }

            func getAnalytics(period: TimePeriod) async throws -> Analytics {
                Analytics(
                    totalIncome: 0,
                    totalExpenses: 0,
                    netBalance: 0,
                    totalTransactions: 0,
                    averageTransactionAmount: 0,
                    categorySummaries: [],
                    sourceSummaries: [],
                    topCategories: [],
                    topSources: [],
                    timePeriod: .month,
                    dateRange: Date()...Date()
                )
            }

            func getSummary() async throws -> SummaryResponse {
                SummaryResponse(balance: 0, totalIncome: 0, totalExpenses: 0)
            }
        }

        let authManager = AuthManager(apiService: MockAPIService())
        return SettingsView()
            .environmentObject(authManager)
    }
}