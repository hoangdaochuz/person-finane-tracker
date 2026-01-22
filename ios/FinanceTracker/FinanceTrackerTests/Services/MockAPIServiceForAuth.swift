import XCTest
@testable import FinanceTracker

class MockAPIServiceForAuth: APIServiceProtocol {
    var mockUser: User?
    var shouldThrowError = false

    func login(email: String, password: String) async throws -> User {
        if shouldThrowError {
            throw APIError.unauthorized
        }
        return mockUser ?? User(email: email, apiKey: "mock_api_key")
    }

    func register(email: String, password: String, name: String? = nil) async throws -> User {
        if shouldThrowError {
            throw APIError.clientError("Registration failed")
        }
        return User(email: email, apiKey: "mock_api_key")
    }

    func getTransactions(page: Int, limit: Int) async throws -> [Transaction] {
        return []
    }

    func getAnalytics(period: TimePeriod) async throws -> Analytics {
        return Analytics(
            totalIncome: 0,
            totalExpenses: 0,
            balance: 0,
            categoryBreakdown: [],
            sourceBreakdown: [],
            period: period
        )
    }

    func getSummary() async throws -> SummaryResponse {
        return SummaryResponse(balance: 0, totalIncome: 0, totalExpenses: 0)
    }

    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        return transaction
    }
}