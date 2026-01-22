import Foundation
import CoreData

// MARK: - Authentication Response Models

// Backend returns: {"token": "jwt", "user": {...}}
struct AuthResponse: Codable {
    let token: String
    let user: UserDTO

    // Convert to iOS User model
    func toUser() -> User {
        return user.toUser()
    }
}

// UserDTO matches the backend's UserResponse JSON format
struct UserDTO: Codable {
    let id: Int64           // Database ID (not used by iOS)
    let uuid: UUID          // This is what iOS uses as the user's id
    let email: String
    let name: String?
    let apiKey: String      // Backend uses api_key (snake_case) but Codable handles mapping
    let isActive: Bool      // Backend uses is_active

    enum CodingKeys: String, CodingKey {
        case id
        case uuid
        case email
        case name
        case apiKey = "api_key"
        case isActive = "is_active"
    }

    // Convert to iOS User model
    func toUser() -> User {
        // Use the UUID as the user's id (not the int64 database id)
        return User(
            id: uuid,
            email: email,
            name: name,
            apiKey: apiKey,
            isBiometricEnabled: false // Default to false, user can enable later
        )
    }
}

// Legacy LoginResponse - kept for compatibility
struct LoginResponse: Codable {
    let apiKey: String
    let user: User
}

// MARK: - Transaction DTO for API responses
struct TransactionDTO: Codable {
    let id: UUID
    let amount: Double
    let type: String
    let merchant: String?
    let category: String?
    let source: String
    let date: Date
    let remoteID: String?

    // Convert DTO to Core Data Transaction
    func toTransaction(in context: NSManagedObjectContext) -> Transaction {
        let transaction = Transaction(context: context,
                                     amount: amount,
                                     type: TransactionType(rawValue: type) ?? .expense,
                                     merchant: merchant,
                                     category: category,
                                     source: source,
                                     date: date,
                                     remoteID: remoteID)
        transaction.id = id
        return transaction
    }

    // Create DTO from Core Data Transaction
    static func from(_ transaction: Transaction) -> TransactionDTO {
        return TransactionDTO(
            id: transaction.id,
            amount: transaction.amount,
            type: transaction.type,
            merchant: transaction.merchant,
            category: transaction.category,
            source: transaction.source,
            date: transaction.date,
            remoteID: transaction.remoteID
        )
    }
}

struct TransactionsResponse: Codable {
    let transactions: [TransactionDTO]
    let page: Int
    let limit: Int
    let total: Int
}

struct CreateTransactionResponse: Codable {
    let transaction: TransactionDTO
}

struct ErrorResponse: Codable {
    let error: String
    let message: String
}

struct SummaryResponse: Codable {
    let balance: Double
    let totalIncome: Double
    let totalExpenses: Double
}