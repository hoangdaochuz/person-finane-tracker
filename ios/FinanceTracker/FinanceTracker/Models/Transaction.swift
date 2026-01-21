//
//  Transaction.swift
//  FinanceTracker
//
//  Created by Claude on 2025-01-20.
//

import Foundation
import CoreData

@objc(Transaction)
public class Transaction: NSManagedObject {
    @NSManaged public var id: UUID
    @NSManaged public var amount: Double
    @NSManaged public var type: String
    @NSManaged public var merchant: String?
    @NSManaged public var category: String?
    @NSManaged public var source: String
    @NSManaged public var date: Date
    @NSManaged public var remoteID: String?
}

extension Transaction {
    convenience init(context: NSManagedObjectContext,
                   amount: Double,
                   type: TransactionType,
                   merchant: String? = nil,
                   category: String? = nil,
                   source: String,
                   date: Date = Date(),
                   remoteID: String? = nil) {

        guard let entity = NSEntityDescription.entity(forEntityName: "TransactionEntity", in: context) else {
            fatalError("TransactionEntity not found in Core Data model")
        }
        self.init(entity: entity, insertInto: context)

        self.id = UUID()
        self.amount = amount
        self.type = type.rawValue
        self.merchant = merchant
        self.category = category
        self.source = source
        self.date = date
        self.remoteID = remoteID
    }

    // For SwiftUI Previews - creates a temporary instance not attached to any context
    static func preview(
        amount: Double,
        type: TransactionType,
        merchant: String? = nil,
        category: String? = nil,
        source: String,
        date: Date = Date(),
        remoteID: String? = nil
    ) -> Transaction {
        let transaction = Transaction()
        transaction.id = UUID()
        transaction.amount = amount
        transaction.type = type.rawValue
        transaction.merchant = merchant
        transaction.category = category
        transaction.source = source
        transaction.date = date
        transaction.remoteID = remoteID
        return transaction
    }
}

// MARK: - Identifiable
extension Transaction: Identifiable {
    public var identifier: UUID {
        return id
    }
}

// MARK: - Transaction Type Computed Property
extension Transaction {
    var transactionType: TransactionType {
        get {
            return TransactionType(rawValue: type) ?? .expense
        }
        set {
            type = newValue.rawValue
        }
    }
}