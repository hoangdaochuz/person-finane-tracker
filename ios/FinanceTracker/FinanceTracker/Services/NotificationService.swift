import Foundation
import UserNotifications
import CoreData
import Combine

class NotificationService: ObservableObject {
    private let apiService: APIServiceProtocol
    private let coreDataStack: CoreDataStack
    private var notificationQueue: [TransactionDTO] = []
    private let center = UNUserNotificationCenter.current()

    var objectWillChange = ObservableObjectPublisher()

    init(apiService: APIServiceProtocol, coreDataStack: CoreDataStack = .shared) {
        self.apiService = apiService
        self.coreDataStack = coreDataStack
        requestNotificationPermissions()
    }

    // MARK: - Permission Methods

    private func requestNotificationPermissions() {
        center.requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
            if granted {
                print("Notification permission granted")
            } else if let error = error {
                print("Notification permission error: \(error)")
            }
        }
    }

    // MARK: - Notification Handling

    func handleNotification(_ notification: UNNotification, completion: @escaping (Bool) -> Void) {
        // Validate notification content
        guard validateNotification(notification) else {
            completion(false)
            return
        }

        // Check if it's a duplicate
        if isDuplicate(notification) {
            completion(true)
            return
        }

        // Process based on priority
        if isHighPriority(notification) {
            processNotification(notification, completion: completion)
        } else {
            processNotification(notification, completion: completion)
        }
    }

    func processNotificationInBackground(_ notification: UNNotification, completion: @escaping (Bool) -> Void) {
        DispatchQueue.global(qos: .background).async {
            self.processNotification(notification) { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }

    // MARK: - Validation Methods

    func validateNotification(_ notification: UNNotification) -> Bool {
        guard let content = notification.request.content.userInfo as? [String: Any] else {
            return false
        }

        // Check for required transaction fields
        guard let _ = content["amount"] as? Double,
              let _ = content["date"] as? TimeInterval else {
            return false
        }

        // Validate content
        let title = notification.request.content.title
        let body = notification.request.content.body

        return !title.isEmpty && !body.isEmpty
    }

    // MARK: - Queue Management

    func addToQueue(_ notification: UNNotification) {
        // Parse notification to transaction DTO
        guard let content = notification.request.content.userInfo as? [String: Any],
              let amount = content["amount"] as? Double,
              let date = content["date"] as? TimeInterval,
              let source = content["source"] as? String else {
            return
        }

        let transactionDTO = TransactionDTO(
            id: UUID(),
            amount: amount,
            type: content["type"] as? String ?? "expense",
            merchant: content["merchant"] as? String,
            category: content["category"] as? String,
            source: source,
            date: Date(timeIntervalSince1970: date),
            remoteID: content["remoteID"] as? String
        )

        notificationQueue.append(transactionDTO)
    }

    func getQueueCount() -> Int {
        return notificationQueue.count
    }

    // MARK: - Filtering Methods

    func filterSpam(_ notification: UNNotification) -> Bool {
        let title = notification.request.content.title.lowercased()
        let body = notification.request.content.body.lowercased()

        let spamKeywords = [
            "winner", "congratulations", "you've won", "prize",
            "limited time", "act now", "urgent", "claim now"
        ]

        for keyword in spamKeywords {
            if title.contains(keyword) || body.contains(keyword) {
                return true
            }
        }

        return false
    }

    func isHighPriority(_ notification: UNNotification) -> Bool {
        guard let amount = notification.request.content.userInfo["amount"] as? Double else {
            return false
        }
        return amount > 500 // High threshold for high priority
    }

    func isDuplicate(_ notification: UNNotification) -> Bool {
        guard let content = notification.request.content.userInfo as? [String: Any],
              let amount = content["amount"] as? Double,
              let date = content["date"] as? TimeInterval else {
            return false
        }

        return notificationQueue.contains { existingTransaction in
            // Check if it's the same transaction (amount and date within 1 minute)
            return existingTransaction.amount == amount &&
                   abs(existingTransaction.date.timeIntervalSince1970 - date) < 60
        }
    }

    // MARK: - Private Methods

    private func processNotification(_ notification: UNNotification, completion: @escaping (Bool) -> Void) {
        guard let content = notification.request.content.userInfo as? [String: Any],
              let amount = content["amount"] as? Double,
              let date = content["date"] as? TimeInterval,
              let source = content["source"] as? String else {
            completion(false)
            return
        }

        // Create transaction DTO
        let transactionDTO = TransactionDTO(
            id: UUID(),
            amount: amount,
            type: content["type"] as? String ?? "expense",
            merchant: content["merchant"] as? String,
            category: content["category"] as? String,
            source: source,
            date: Date(timeIntervalSince1970: date),
            remoteID: content["remoteID"] as? String
        )

        // Convert to Core Data Transaction
        let transaction = transactionDTO.toTransaction(in: coreDataStack.viewContext)

        // Save to Core Data
        saveTransactionToCache(transaction)

        // Sync with API
        Task {
            do {
                _ = try await apiService.createTransaction(transaction)
                completion(true)
            } catch {
                print("Failed to sync transaction to API: \(error)")
                completion(false)
            }
        }
    }

    private func saveTransactionToCache(_ transaction: Transaction) {
        do {
            try coreDataStack.save()
            print("Transaction saved to cache: \(transaction.amount)")
        } catch {
            print("Failed to save transaction to Core Data: \(error)")

            // Try to delete the failed transaction to avoid orphaned records
            coreDataStack.viewContext.delete(transaction)

            // Retry save after cleanup
            do {
                try coreDataStack.save()
            } catch {
                // Log the error but don't crash - notification will be retried later
                print("Failed to cleanup after transaction save failure: \(error)")
            }
        }
    }
}
