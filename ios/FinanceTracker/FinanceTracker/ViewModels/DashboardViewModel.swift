import Foundation
import Combine

@MainActor
class DashboardViewModel: ObservableObject {
    @Published var recentTransactions: [Transaction] = []
    @Published var balance: Double = 0
    @Published var totalIncome: Double = 0
    @Published var totalExpenses: Double = 0
    @Published var isLoading = true
    @Published var errorMessage: String?

    private let apiService: APIServiceProtocol
    private var cancellables = Set<AnyCancellable>()

    init(apiService: APIServiceProtocol) {
        self.apiService = apiService
    }

    func loadData() async {
        isLoading = true
        errorMessage = nil

        do {
            let transactions = try await apiService.getTransactions(page: 1, limit: 5)
            self.recentTransactions = transactions

            let summary = try await apiService.getSummary()
            self.balance = summary.balance
            self.totalIncome = summary.totalIncome
            self.totalExpenses = summary.totalExpenses

            isLoading = false
        } catch {
            isLoading = false
            errorMessage = error.localizedDescription
        }
    }

    func refresh() async {
        await loadData()
    }
}