import Foundation
import CoreData

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