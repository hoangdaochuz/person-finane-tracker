import XCTest
@testable import FinanceTracker

class MockAPIServiceForDashboard: APIServiceProtocol {
    var mockTransactions: [Transaction] = []
    var mockSummary: SummaryResponse?
    var shouldThrowError = false

    func getTransactions(page: Int, limit: Int) async throws -> [Transaction] {
        if shouldThrowError { throw APIError.serverError }
        return Array(mockTransactions.prefix(limit))
    }

    func getAnalytics(period: TimePeriod) async throws -> Analytics {
        if shouldThrowError { throw APIError.serverError }
        return Analytics(
            totalIncome: 5000,
            totalExpenses: 3000,
            balance: 2000,
            categoryBreakdown: [],
            sourceBreakdown: [],
            period: period
        )
    }

    func getSummary() async throws -> SummaryResponse {
        if shouldThrowError { throw APIError.serverError }
        return mockSummary ?? SummaryResponse(balance: 2000, totalIncome: 5000, totalExpenses: 3000)
    }

    func login(email: String, password: String) async throws -> User {
        if shouldThrowError { throw APIError.unauthorized }
        return User(email: email, apiKey: "mock_api_key")
    }

    func register(email: String, password: String) async throws -> User {
        if shouldThrowError { throw APIError.clientError("Registration failed") }
        return User(email: email, apiKey: "mock_api_key")
    }

    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        return transaction
    }
}