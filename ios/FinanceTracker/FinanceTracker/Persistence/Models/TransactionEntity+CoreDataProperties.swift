import Foundation
import CoreData

extension TransactionEntity {
    @nonobjc public class func fetchRequest() -> NSFetchRequest<TransactionEntity> {
        return NSFetchRequest<TransactionEntity>(entityName: "TransactionEntity")
    }

    @NSManaged public var amount: Double
    @NSManaged public var category: String?
    @NSManaged public var date: Date
    @NSManaged public var id: UUID
    @NSManaged public var merchant: String?
    @NSManaged public var remoteID: String?
    @NSManaged public var source: String  // Changed from optional to non-optional
    @NSManaged public var type: String
}

extension TransactionEntity: Identifiable {}

// MARK: - Transaction Type Computed Property
extension TransactionEntity {
    var transactionType: TransactionType {
        get {
            return TransactionType(rawValue: type) ?? .expense
        }
        set {
            type = newValue.rawValue
        }
    }
}
