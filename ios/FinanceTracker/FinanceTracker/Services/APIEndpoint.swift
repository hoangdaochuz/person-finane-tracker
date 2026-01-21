import Foundation

enum APIEndpoint {
    case login(email: String, password: String)
    case register(email: String, password: String)
    case createTransaction(Transaction)
    case getTransactions(page: Int, limit: Int)
    case getAnalytics(period: TimePeriod)
    case getSummary

    var path: String {
        switch self {
        case .login: return "/api/v1/auth/login"
        case .register: return "/api/v1/auth/register"
        case .createTransaction: return "/api/v1/transactions"
        case .getTransactions: return "/api/v1/transactions"
        case .getAnalytics: return "/api/v1/analytics"
        case .getSummary: return "/api/v1/summary"
        }
    }

    var method: String {
        switch self {
        case .login, .register, .createTransaction:
            return "POST"
        case .getTransactions, .getAnalytics, .getSummary:
            return "GET"
        }
    }
}