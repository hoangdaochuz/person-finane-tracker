import Foundation
import Combine

@MainActor
class TransactionsViewModel: ObservableObject {
    @Published var transactions: [Transaction] = []
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var selectedType: TransactionType?
    @Published var selectedSource: String?
    @Published var searchText: String = ""

    private let apiService: APIServiceProtocol
    private var currentPage = 1
    private let pageSize = 20
    private var hasMorePages = true

    var hasMorePagesInternal: Bool {
        return hasMorePages
    }

    init(apiService: APIServiceProtocol) {
        self.apiService = apiService
    }

    func loadTransactions() async {
        isLoading = true
        errorMessage = nil

        do {
            let newTransactions = try await apiService.getTransactions(page: currentPage, limit: pageSize)

            if currentPage == 1 {
                transactions = newTransactions
            } else {
                transactions.append(contentsOf: newTransactions)
            }

            hasMorePages = newTransactions.count == pageSize
            isLoading = false
        } catch {
            isLoading = false
            errorMessage = error.localizedDescription
        }
    }

    func loadMore() async {
        guard !isLoading && hasMorePages else { return }
        currentPage += 1
        await loadTransactions()
    }

    func refresh() async {
        currentPage = 1
        await loadTransactions()
    }

    var filteredTransactions: [Transaction] {
        transactions.filter { transaction in
            if let type = selectedType, TransactionType(rawValue: transaction.type) != type {
                return false
            }
            if let source = selectedSource, transaction.source != source {
                return false
            }
            if !searchText.isEmpty {
                let search = searchText.lowercased()
                return transaction.merchant?.lowercased().contains(search) == true
                    || transaction.category?.lowercased().contains(search) == true
            }
            return true
        }
    }

    var uniqueSources: [String] {
        Set(transactions.compactMap { $0.source })
            .sorted()
    }
}