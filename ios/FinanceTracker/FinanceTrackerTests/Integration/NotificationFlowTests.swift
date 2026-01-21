import XCTest
import UserNotifications
@testable import FinanceTracker

final class NotificationFlowTests: XCTestCase {

    private var notificationService: NotificationService!
    private var mockAPIService: MockAPIService!

    override func setUp() {
        super.setUp()
        mockAPIService = MockAPIService()
        notificationService = NotificationService(apiService: mockAPIService)
    }

    override func tearDown() {
        notificationService = nil
        mockAPIService = nil
        super.tearDown()
    }

    // MARK: - Permission Tests

    func testNotificationPermissionRequest() {
        // Given
        let expectation = XCTestExpectation(description: "Permission request completes")

        // When
        notificationService.requestNotificationPermissions { granted, error in
            // Then
            XCTAssertNotNil(granted)
            XCTAssertNil(error)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testNotificationPermissionDenied() {
        // Given
        let expectation = XCTestExpectation(description: "Permission denied handled")

        // When
        notificationService.requestNotificationPermissions { granted, error in
            // Then
            XCTAssertFalse(granted)
            XCTAssertNotNil(error)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    // MARK: - Notification Handling Tests

    func testTransactionNotificationHandling() {
        // Given
        let notification = createMockTransactionNotification()
        let expectation = XCTestExpectation(description: "Transaction notification processed")

        // When
        notificationService.handleNotification(notification) { success in
            // Then
            XCTAssertTrue(success)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testNotificationContentValidation() {
        // Given
        let invalidNotification = createMockNotification(title: "", body: "")

        // When
        let isValid = notificationService.validateNotification(invalidNotification)

        // Then
        XCTAssertFalse(isValid)
    }

    func testNotificationPriorityHandling() {
        // Given
        let highPriorityNotification = createMockTransactionNotification(amount: 1000)
        let normalNotification = createMockTransactionNotification(amount: 50)

        // When
        let highPriorityHandled = notificationService.handleNotification(highPriorityNotification)
        let normalPriorityHandled = notificationService.handleNotification(normalNotification)

        // Then
        XCTAssertTrue(highPriorityHandled)
        XCTAssertTrue(normalPriorityHandled)
    }

    // MARK: - Background Processing Tests

    func testBackgroundNotificationProcessing() {
        // Given
        let notification = createMockTransactionNotification()
        let expectation = XCTestExpectation(description: "Background processing completes")

        // When
        notificationService.processNotificationInBackground(notification) { success in
            // Then
            XCTAssertTrue(success)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 10.0)
    }

    func testNotificationQueueManagement() {
        // Given
        let notifications = (1...10).map { i in
            createMockTransactionNotification(title: "Notification \(i)")
        }

        // When
        for notification in notifications {
            notificationService.addToQueue(notification)
        }

        let queueCount = notificationService.getQueueCount()

        // Then
        XCTAssertEqual(queueCount, 10)
    }

    // MARK: - Error Handling Tests

    func testNetworkErrorNotificationHandling() {
        // Given
        mockAPIService.shouldFail = true
        let notification = createMockTransactionNotification()
        let expectation = XCTestExpectation(description: "Network error handled")

        // When
        notificationService.handleNotification(notification) { success in
            // Then
            XCTAssertFalse(success)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testMalformedNotificationDataHandling() {
        // Given
        let malformedNotification = UNMutableNotificationContent()
        malformedNotification.title = "Test"
        malformedNotification.body = "Test body"
        malformedNotification.userInfo = ["invalid_key": "value"] // Missing required fields

        // When
        let isValid = notificationService.validateNotification(malformedNotification)

        // Then
        XCTAssertFalse(isValid)
    }

    // MARK: - Content Filtering Tests

    func testNotificationContentFiltering() {
        // Given
        let spamNotification = createMockTransactionNotification(
            title: "🎉 WINNER! 🎉",
            body: "You've won $1000!",
            amount: 0
        )

        let validNotification = createMockTransactionNotification(
            title: "Payment Received",
            body: "John Doe sent you $100",
            amount: 100
        )

        // When
        let spamFiltered = notificationService.filterSpam(spamNotification)
        let validNotFiltered = notificationService.filterSpam(validNotification)

        // Then
        XCTAssertTrue(spamFiltered)
        XCTAssertFalse(validNotFiltered)
    }

    func testDuplicateNotificationDetection() {
        // Given
        let notification = createMockTransactionNotification()
        notificationService.addToQueue(notification)

        // When
        let isDuplicate = notificationService.isDuplicate(notification)

        // Then
        XCTAssertTrue(isDuplicate)
    }

    // MARK: - Helper Methods

    private func createMockTransactionNotification(
        title: String = "Transaction Alert",
        body: String = "A new transaction has been detected",
        amount: Double = 100
    ) -> UNNotification {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default
        content.userInfo = [
            "type": "transaction",
            "amount": amount,
            "date": Date().timeIntervalSince1970,
            "sender": "Bank App",
            "recipient": "Your Account"
        ]

        return UNNotification(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
    }

    private func createMockNotification(title: String, body: String) -> UNNotification {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default

        return UNNotification(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
    }
}

// MARK: - Mock API Service for Testing

class MockAPIService: APIServiceProtocol {
    var shouldFail = false
    var responseDelay: TimeInterval = 0.1

    func fetchTransactionDetails(for identifier: String, completion: @escaping (Result<TransactionDetails, APIError>) -> Void) {
        if shouldFail {
            completion(.failure(.networkError))
        } else {
            let details = TransactionDetails(
                amount: 100.0,
                currency: "USD",
                date: Date(),
                sender: "Test Sender",
                recipient: "Test Recipient",
                type: .transfer,
                description: "Test transaction"
            )
            completion(.success(details))
        }
    }

    func fetchAccountBalance(for account: String, completion: @escaping (Result<Double, APIError>) -> Void) {
        if shouldFail {
            completion(.failure(.networkError))
        } else {
            completion(.success(5000.0))
        }
    }

    func fetchTransactionHistory(limit: Int, offset: Int, completion: @escaping (Result<[Transaction], APIError>) -> Void) {
        if shouldFail {
            completion(.failure(.networkError))
        } else {
            let transactions = [
                Transaction(
                    amount: 100.0,
                    currency: "USD",
                    date: Date(),
                    sender: "Test Sender",
                    recipient: "Test Recipient",
                    type: .transfer,
                    description: "Test transaction"
                )
            ]
            completion(.success(transactions))
        }
    }

    func submitTransactionFeedback(_ feedback: TransactionFeedback, completion: @escaping (Result<Bool, APIError>) -> Void) {
        if shouldFail {
            completion(.failure(.networkError))
        } else {
            completion(.success(true))
        }
    }

    func updateNotificationSettings(settings: NotificationSettings, completion: @escaping (Result<Bool, APIError>) -> Void) {
        if shouldFail {
            completion(.failure(.networkError))
        } else {
            completion(.success(true))
        }
    }
}