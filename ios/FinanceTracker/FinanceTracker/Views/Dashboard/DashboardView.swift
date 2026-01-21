import SwiftUI

struct DashboardView: View {
    @StateObject private var viewModel: DashboardViewModel
    @State private var showingAddTransaction = false

    init(viewModel: DashboardViewModel) {
        _viewModel = StateObject(wrappedValue: viewModel)
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 20) {
                headerSection
                statsSection
                recentTransactionsSection
            }
            .padding()
        }
        .background(ColorPalette.background)
        .refreshable {
            await viewModel.refresh()
        }
        .overlay {
            if viewModel.isLoading && viewModel.recentTransactions.isEmpty {
                ProgressView()
            }
        }
        .sheet(isPresented: $showingAddTransaction) {
            Text("Add Transaction View")
        }
    }

    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text("Good \(timeOfDay)")
                    .font(Typography.subheadline)
                    .foregroundColor(ColorPalette.textSecondary)

                Text("Dashboard")
                    .font(Typography.largeTitle)
                    .foregroundColor(ColorPalette.textPrimary)
            }

            Spacer()

            Button {
                // TODO: Open notifications
            } label: {
                Image(systemName: "bell.badge")
                    .font(.title3)
                    .foregroundColor(ColorPalette.textPrimary)
            }
        }
    }

    private var statsSection: some View {
        VStack(spacing: 12) {
            StatCard(
                icon: "dollarsign.circle.fill",
                title: "Balance",
                value: formatCurrency(viewModel.balance),
                trend: nil,
                isPositive: true
            )

            HStack(spacing: 12) {
                StatCard(
                    icon: "arrow.down.circle.fill",
                    title: "Income",
                    value: formatCurrency(viewModel.totalIncome),
                    trend: nil,
                    isPositive: true
                )

                StatCard(
                    icon: "arrow.up.circle.fill",
                    title: "Expenses",
                    value: formatCurrency(viewModel.totalExpenses),
                    trend: nil,
                    isPositive: true
                )
            }
        }
    }

    private var recentTransactionsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Recent Transactions")
                    .font(Typography.title3)
                    .foregroundColor(ColorPalette.textPrimary)

                Spacer()

                Button("See All") {
                    // TODO: Navigate to transactions list
                }
                .font(Typography.subheadlineMedium)
                .foregroundColor(ColorPalette.primaryIndigo)
            }

            if viewModel.recentTransactions.isEmpty {
                Text("No transactions yet")
                    .font(Typography.body)
                    .foregroundColor(ColorPalette.textSecondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.vertical, 20)
            } else {
                VStack(spacing: 0) {
                    ForEach(viewModel.recentTransactions) { transaction in
                        TransactionCell(transaction: transaction)

                        if transaction.id != viewModel.recentTransactions.last?.id {
                            Divider()
                                .padding(.leading, 68)
                        }
                    }
                }
                .background(ColorPalette.cardBackground)
                .cornerRadius(16)
                .shadow(color: .black.opacity(0.05), radius: 10, x: 0, y: 4)
            }
        }
    }

    private var timeOfDay: String {
        let hour = Calendar.current.component(.hour, from: Date())
        switch hour {
        case 0..<12: return "Morning"
        case 12..<18: return "Afternoon"
        default: return "Evening"
        }
    }

    private func formatCurrency(_ value: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        return formatter.string(from: NSNumber(value: value)) ?? "$0.00"
    }
}

struct DashboardView_Previews: PreviewProvider {
    static var previews: some View {
        // Create a simple mock API service
        class MockAPIService: APIServiceProtocol {
            func login(email: String, password: String) async throws -> User {
                User(email: email, isBiometricEnabled: false)
            }

            func register(email: String, password: String) async throws -> User {
                User(email: email, isBiometricEnabled: false)
            }

            func createTransaction(_ transaction: Transaction) async throws -> Transaction {
                transaction
            }

            func getTransactions(page: Int, limit: Int) async throws -> [Transaction] {
                [
                    Transaction.preview(amount: 1500, type: .income, merchant: "Salary", source: "BCA"),
                    Transaction.preview(amount: 50, type: .expense, merchant: "Coffee", source: "Gopay")
                ]
            }

            func getAnalytics(period: TimePeriod) async throws -> Analytics {
                Analytics(
                    totalIncome: 5000,
                    totalExpenses: 3000,
                    netBalance: 2000,
                    totalTransactions: 2,
                    averageTransactionAmount: 775,
                    categorySummaries: [],
                    sourceSummaries: [],
                    topCategories: [],
                    topSources: [],
                    timePeriod: .month,
                    dateRange: Date()...Date()
                )
            }

            func getSummary() async throws -> SummaryResponse {
                SummaryResponse(balance: 2000, totalIncome: 5000, totalExpenses: 3000)
            }
        }

        let viewModel = DashboardViewModel(apiService: MockAPIService())
        return DashboardView(viewModel: viewModel)
    }
}