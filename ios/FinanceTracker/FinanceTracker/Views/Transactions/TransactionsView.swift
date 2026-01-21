import SwiftUI

struct TransactionsView: View {
    @StateObject private var viewModel: TransactionsViewModel
    @State private var selectedTransaction: Transaction?

    init(apiService: APIServiceProtocol) {
        _viewModel = StateObject(wrappedValue: TransactionsViewModel(apiService: apiService))
    }

    var body: some View {
        VStack(spacing: 0) {
            filterSection

            if viewModel.filteredTransactions.isEmpty && !viewModel.isLoading {
                emptyState
            } else {
                List {
                    ForEach(viewModel.filteredTransactions) { transaction in
                        Button {
                            selectedTransaction = transaction
                        } label: {
                            TransactionCell(transaction: transaction)
                                .listRowInsets(EdgeInsets(top: 4, leading: 16, bottom: 4, trailing: 16))
                                .listRowSeparator(.hidden)
                                .listRowBackground(Color.clear)
                        }

                        if transaction.id == viewModel.filteredTransactions.last?.id && viewModel.hasMorePagesInternal {
                            Color.clear
                                .onAppear {
                                    Task {
                                        await viewModel.loadMore()
                                    }
                                }
                        }
                    }
                }
                .listStyle(.plain)
                .background(ColorPalette.background)
            }
        }
        .background(ColorPalette.background)
        .refreshable {
            await viewModel.refresh()
        }
        .sheet(item: $selectedTransaction) { transaction in
            TransactionDetailView(transaction: transaction)
        }
        .searchable(text: $viewModel.searchText, prompt: "Search transactions...")
        .task {
            if viewModel.transactions.isEmpty {
                await viewModel.loadTransactions()
            }
        }
    }

    private var filterSection: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                FilterChip(
                    title: "All",
                    isSelected: viewModel.selectedType == nil
                ) {
                    viewModel.selectedType = nil
                }

                ForEach(TransactionType.allCases, id: \.self) { type in
                    FilterChip(
                        title: type.displayName,
                        isSelected: viewModel.selectedType == type
                    ) {
                        viewModel.selectedType = type
                    }
                }

                Spacer()
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
        }
        .background(ColorPalette.cardBackground)
    }

    private var emptyState: some View {
        VStack(spacing: 16) {
            Image(systemName: "tray")
                .font(.system(size: 60))
                .foregroundColor(ColorPalette.textSecondary)

            Text("No transactions found")
                .font(Typography.body)
                .foregroundColor(ColorPalette.textSecondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct FilterChip: View {
    let title: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(Typography.subheadlineMedium)
                .foregroundColor(isSelected ? .white : ColorPalette.textPrimary)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(isSelected ? ColorPalette.primaryIndigo : Color.gray.opacity(0.1))
                .cornerRadius(20)
        }
    }
}

struct TransactionsView_Previews: PreviewProvider {
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
                    Transaction.preview(amount: 1500, type: .income, merchant: "Salary", category: "Income", source: "BCA"),
                    Transaction.preview(amount: 50, type: .expense, merchant: "Coffee", category: "Food", source: "Gopay"),
                    Transaction.preview(amount: 25, type: .expense, merchant: "Lunch", category: "Food", source: "OVO")
                ]
            }

            func getAnalytics(period: TimePeriod) async throws -> Analytics {
                Analytics(
                    totalIncome: 1500,
                    totalExpenses: 75,
                    netBalance: 1425,
                    totalTransactions: 3,
                    averageTransactionAmount: 525,
                    categorySummaries: [],
                    sourceSummaries: [],
                    topCategories: [],
                    topSources: [],
                    timePeriod: .month,
                    dateRange: Date()...Date()
                )
            }

            func getSummary() async throws -> SummaryResponse {
                SummaryResponse(balance: 1425, totalIncome: 1500, totalExpenses: 75)
            }
        }

        return TransactionsView(apiService: MockAPIService())
    }
}