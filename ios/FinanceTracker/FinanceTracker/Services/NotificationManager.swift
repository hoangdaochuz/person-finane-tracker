import Foundation
import UserNotifications
import CoreData
import Combine

class NotificationManager: NSObject, ObservableObject {
    private let transactionParser: TransactionParser
    private let apiService: APIService
    private let coreDataStack: CoreDataStack

    var onNewTransaction: ((Transaction) -> Void)?

    // Required for ObservableObject conformance when no @Published properties exist
    let objectWillChange = ObservableObjectPublisher()

    init(
        transactionParser: TransactionParser = TransactionParser(),
        apiService: APIService = APIService(baseURL: Config.baseURL),
        coreDataStack: CoreDataStack = .shared
    ) {
        self.transactionParser = transactionParser
        self.apiService = apiService
        self.coreDataStack = coreDataStack
        super.init()
        requestNotificationAuthorization()
    }

    func requestNotificationAuthorization() {
        let center = UNUserNotificationCenter.current()
        center.requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
            if granted {
                print("Notification authorization granted")
            } else if let error = error {
                print("Notification authorization error: \(error)")
            }
        }
        center.delegate = self
    }

    private func processNotification(_ notification: UNNotification) {
        guard let content = notification.request.content.userInfo as? [String: Any],
              let aps = content["aps"] as? [String: Any],
              let alert = aps["alert"] as? [String: Any],
              let body = alert["body"] as? String else {
            return
        }

        let source = notification.request.content.categoryIdentifier

        guard let transactionDTO = transactionParser.parse(notificationText: body, source: source) else {
            print("Failed to parse transaction from notification")
            return
        }

        // Convert DTO to Core Data Transaction
        let transaction = transactionDTO.toTransaction(in: coreDataStack.viewContext)
        saveTransactionToCache(transaction)

        Task { @MainActor in
            do {
                let savedTransaction = try await apiService.createTransaction(transaction)
                onNewTransaction?(savedTransaction)
            } catch {
                print("Failed to sync transaction to API: \(error)")
            }
        }
    }

    private func saveTransactionToCache(_ transaction: Transaction) {
        do {
            try coreDataStack.viewContext.save()
            print("Transaction saved to cache: \(transaction.amount)")
        } catch {
            print("Failed to save transaction to Core Data: \(error)")
        }
    }
}

extension NotificationManager: UNUserNotificationCenterDelegate {
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        processNotification(response.notification)
        completionHandler()
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        processNotification(notification)

        #if os(iOS)
        completionHandler([.banner, .sound, .badge])
        #else
        completionHandler([])
        #endif
    }
}
