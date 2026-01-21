import Foundation
import CoreData

class APIService: APIServiceProtocol {
    private let session: URLSessionProtocol
    private let baseURL: String
    private let keychainManager: KeychainManager
    private let coreDataStack: CoreDataStack

    var apiKey: String? {
        keychainManager.retrieve(key: "finance_tracker_api_key")
    }

    init(session: URLSessionProtocol = URLSession.shared, baseURL: String, keychainManager: KeychainManager = KeychainManager(), coreDataStack: CoreDataStack = .shared) {
        self.session = session
        self.baseURL = baseURL
        self.keychainManager = keychainManager
        self.coreDataStack = coreDataStack
    }

    func login(email: String, password: String) async throws -> User {
        var request = createRequest(for: .login(email: email, password: password))
        request.httpBody = try JSONEncoder().encode(["email": email, "password": password])

        let (data, response) = try await session.data(for: request)
        return try handleResponse(data, response: response)
    }

    func register(email: String, password: String) async throws -> User {
        var request = createRequest(for: .register(email: email, password: password))
        request.httpBody = try JSONEncoder().encode(["email": email, "password": password])

        let (data, response) = try await session.data(for: request)
        return try handleResponse(data, response: response)
    }

    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        // Convert Transaction to TransactionDTO for encoding
        let transactionDTO = TransactionDTO.from(transaction)
        var request = createAuthenticatedRequest(for: .createTransaction(transaction))
        request.httpBody = try JSONEncoder().encode(transactionDTO)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let (data, response) = try await session.data(for: request)
        let createResponse: CreateTransactionResponse = try handleResponse(data, response: response)
        // Convert DTO back to Transaction
        return createResponse.transaction.toTransaction(in: coreDataStack.viewContext)
    }

    func getTransactions(page: Int = 1, limit: Int = 20) async throws -> [Transaction] {
        let request = createAuthenticatedRequest(for: .getTransactions(page: page, limit: limit))
        let (data, response) = try await session.data(for: request)
        let result: TransactionsResponse = try handleResponse(data, response: response)
        // Convert DTOs to Transactions
        return result.transactions.map { $0.toTransaction(in: coreDataStack.viewContext) }
    }

    func getAnalytics(period: TimePeriod) async throws -> Analytics {
        let request = createAuthenticatedRequest(for: .getAnalytics(period: period))
        let (data, response) = try await session.data(for: request)
        return try handleResponse(data, response: response)
    }

    func getSummary() async throws -> SummaryResponse {
        let request = createAuthenticatedRequest(for: .getSummary)
        let (data, response) = try await session.data(for: request)
        return try handleResponse(data, response: response)
    }

    private func createRequest(for endpoint: APIEndpoint) -> URLRequest {
        var components = URLComponents(string: baseURL + endpoint.path)

        if case .getTransactions(let page, let limit) = endpoint {
            components?.queryItems = [
                URLQueryItem(name: "page", value: String(page)),
                URLQueryItem(name: "limit", value: String(limit))
            ]
        } else if case .getAnalytics(let period) = endpoint {
            components?.queryItems = [
                URLQueryItem(name: "period", value: period.rawValue)
            ]
        }

        var request = URLRequest(url: components!.url!)
        request.httpMethod = endpoint.method
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        return request
    }

    private func createAuthenticatedRequest(for endpoint: APIEndpoint) -> URLRequest {
        var request = createRequest(for: endpoint)

        if let apiKey = apiKey {
            request.setValue(apiKey, forHTTPHeaderField: "X-API-Key")
        }

        return request
    }

    private func handleResponse<T: Decodable>(_ data: Data, response: URLResponse) throws -> T {
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        switch httpResponse.statusCode {
        case 200...299:
            return try JSONDecoder().decode(T.self, from: data)
        case 401:
            throw APIError.unauthorized
        case 400...499:
            let errorResponse = try? JSONDecoder().decode(ErrorResponse.self, from: data)
            throw APIError.clientError(errorResponse?.message ?? "Unknown error")
        case 500...599:
            throw APIError.serverError
        default:
            throw APIError.unknown
        }
    }
}

enum APIError: LocalizedError {
    case invalidResponse
    case unauthorized
    case clientError(String)
    case serverError
    case unknown

    var errorDescription: String? {
        switch self {
        case .invalidResponse: return "Invalid response from server"
        case .unauthorized: return "Unauthorized. Please log in again."
        case .clientError(let message): return message
        case .serverError: return "Server error. Please try again later."
        case .unknown: return "An unknown error occurred"
        }
    }
}

protocol URLSessionProtocol {
    func data(for request: URLRequest) async throws -> (Data, URLResponse)
}

extension URLSession: URLSessionProtocol {}
