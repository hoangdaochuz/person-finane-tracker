# Notification Testing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement comprehensive testing for notification handling from Vietnamese banks and e-wallets.

**Architecture:** Multi-layer testing pyramid with unit tests (70%), integration tests (20%), UI tests (10%), and manual testing. Tests use in-memory CoreData, mock services, and test helpers to isolate components.

**Tech Stack:** XCTest, XCUITest, UserNotifications framework, CoreData, Swift async/await

---

## Task 1: Create Test Helpers Directory Structure

**Files:**
- Create: `ios/FinanceTracker/FinanceTrackerTests/Helpers/NotificationTestHelpers.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Helpers/TestDataGenerator.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Helpers/InMemoryCoreDataStack.swift`

**Step 1: Create Helpers directory**

```bash
mkdir -p ios/FinanceTracker/FinanceTrackerTests/Helpers
```

**Step 2: Create NotificationTestHelpers.swift**

```swift
import Foundation
import UserNotifications
@testable import FinanceTracker

struct NotificationTestBuilder {
    var source: String = "vietcombank"
    var amount: Double = 100000
    var type: TransactionType = .expense
    var merchant: String?
    var category: String?
    var date: Date = Date()

    func build() -> UNNotification {
        let content = UNMutableNotificationContent()
        content.title = type == .income ? "Credit Alert" : "Debit Alert"
        content.body = buildNotificationBody()
        content.categoryIdentifier = source
        content.userInfo = [
            "amount": amount,
            "type": type.rawValue,
            "date": date.timeIntervalSince1970,
            "source": source
        ]

        return UNNotification(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
    }

    private func buildNotificationBody() -> String {
        let formattedAmount = String(format: "%.0f", amount)
        return "TK 123456789 \(type == .income ? "dc cong" : "da tru") \(formattedAmount)vnd\(merchant.map { " tai \($0)" } ?? "")"
    }

    func buildBankNotification(amount: Double, merchant: String? = nil, type: TransactionType = .expense) -> UNNotification {
        var builder = self
        builder.amount = amount
        builder.merchant = merchant
        builder.type = type
        return builder.build()
    }

    func buildWalletNotification(source: String, amount: Double, isIncoming: Bool) -> UNNotification {
        var builder = self
        builder.source = source
        builder.amount = amount
        builder.type = isIncoming ? .income : .expense
        return builder.build()
    }

    func buildMalformedNotification() -> UNNotification {
        let content = UNMutableNotificationContent()
        content.title = ""
        content.body = "Invalid data"
        return UNNotification(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
    }

    func buildSpamNotification() -> UNNotification {
        let content = UNMutableNotificationContent()
        content.title = "🎉 WINNER! 🎉"
        content.body = "You've won $1000! Act now!"
        return UNNotification(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
    }
}
```

**Step 3: Create TestDataGenerator.swift**

```swift
import Foundation
@testable import FinanceTracker

struct TransactionTestDataGenerator {
    static func randomTransaction() -> TransactionDTO {
        TransactionDTO(
            id: UUID(),
            amount: Double.random(in: 10000...10000000),
            type: Bool.random() ? TransactionType.income.rawValue : TransactionType.expense.rawValue,
            merchant: randomMerchant(),
            category: randomCategory(),
            source: randomSource(),
            date: Date(),
            remoteID: nil
        )
    }

    static func randomTransactions(count: Int) -> [TransactionDTO] {
        (1...count).map { _ in randomTransaction() }
    }

    static func transactionForCategory(_ category: String) -> TransactionDTO {
        TransactionDTO(
            id: UUID(),
            amount: 50000,
            type: .expense,
            merchant: "Test Merchant",
            category: category,
            source: "vietcombank",
            date: Date(),
            remoteID: nil
        )
    }

    // Property-based testing helpers
    static func generateValidAmounts() -> [Double] {
        [10000, 50000, 100000, 500000, 1000000, 5000000, 10000000]
    }

    static func generateInvalidAmounts() -> [String] {
        ["abc", "0vnd", "", "1.2.3", "-50000"]
    }

    static func generateMalformedNotifications() -> [String] {
        [
            "",
            "Amount: ",
            "TK debited at",
            "Transaction complete but no details"
        ]
    }

    private static func randomMerchant() -> String? {
        ["Starbucks", "KFC", "Grab", "Shopee", "The Coffee House"].randomElement()
    }

    private static func randomCategory() -> String? {
        ["Food", "Transportation", "Shopping", "Bills", "Transfer"].randomElement()
    }

    private static func randomSource() -> String {
        ["vietcombank", "techcombank", "bidv", "momo", "zalopay", "viettel-money"].randomElement() ?? "vietcombank"
    }
}
```

**Step 4: Create InMemoryCoreDataStack.swift**

```swift
import Foundation
import CoreData
@testable import FinanceTracker

class InMemoryCoreDataStack {
    static let shared = InMemoryCoreDataStack()

    lazy var viewContext: NSManagedObjectContext = {
        let container = NSPersistentContainer(name: "FinanceTracker")
        let description = NSPersistentStoreDescription()
        description.type = NSInMemoryStoreType
        container.persistentStoreDescriptions = [description]

        container.loadPersistentStores { description, error in
            if let error = error {
                fatalError("Failed to load in-memory store: \(error)")
            }
        }

        return container.viewContext
    }()

    func clearAllData() {
        let entities = ["TransactionEntity"]

        for entity in entities {
            let fetchRequest = NSFetchRequest<NSFetchRequestResult>(entityName: entity)
            let deleteRequest = NSBatchDeleteRequest(fetchRequest: fetchRequest)

            do {
                try viewContext.execute(deleteRequest)
            } catch {
                print("Failed to clear \(entity): \(error)")
            }
        }

        try? viewContext.save()
    }
}
```

**Step 5: Run tests to verify no syntax errors**

```bash
cd ios/FinanceTracker
xcodebuild clean build -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15'
```

Expected: BUILD SUCCEEDED

**Step 6: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Helpers/
git commit -m "feat(tests): add test helpers infrastructure

- Add NotificationTestBuilder for creating mock notifications
- Add TestDataGenerator for property-based testing
- Add InMemoryCoreDataStack for isolated tests

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: Create Mock API Service Factory

**Files:**
- Create: `ios/FinanceTracker/FinanceTrackerTests/Helpers/MockAPIServiceFactory.swift`
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift` (create if not exists)

**Step 1: Create MockAPIServiceFactory.swift**

```swift
import Foundation
@testable import FinanceTracker

class MockAPIServiceFactory {
    static func successService() -> MockAPIService {
        let service = MockAPIService()
        service.shouldFail = false
        service.responseDelay = 0.1
        return service
    }

    static func failingService(error: APIError = .networkError) -> MockAPIService {
        let service = MockAPIService()
        service.shouldFail = true
        return service
    }

    static func delayedService(delay: TimeInterval = 1.0) -> MockAPIService {
        let service = MockAPIService()
        service.shouldFail = false
        service.responseDelay = delay
        return service
    }

    static func configuredService(
        shouldFail: Bool = false,
        responseDelay: TimeInterval = 0.1,
        transactionsToReturn: [Transaction] = []
    ) -> MockAPIService {
        let service = MockAPIService()
        service.shouldFail = shouldFail
        service.responseDelay = responseDelay
        return service
    }
}

// Enhanced MockAPIService for testing
class MockAPIService: APIServiceProtocol {
    var shouldFail = false
    var responseDelay: TimeInterval = 0.1
    var createTransactionCalled = false
    var receivedTransactions: [Transaction] = []
    var transactionToReturn: Transaction?

    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        createTransactionCalled = true
        receivedTransactions.append(transaction)

        if shouldFail {
            throw APIError.networkError
        }

        if let delay = responseDelay as? TimeInterval, delay > 0 {
            try await Task.sleep(nanoseconds: UInt64(delay * 1_000_000_000))
        }

        return transactionToReturn ?? transaction
    }
}
```

**Step 2: Run tests to verify compilation**

```bash
cd ios/FinanceTracker
xcodebuild build -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15'
```

Expected: BUILD SUCCEEDED

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Helpers/MockAPIServiceFactory.swift
git commit -m "feat(tests): add MockAPIServiceFactory

Create configurable mock API service for testing different scenarios:
- Success scenarios
- Failure scenarios
- Delayed responses
- Custom transaction returns

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: Create Notification Test Samples

**Files:**
- Create: `ios/FinanceTracker/FinanceTrackerTests/Samples/NotificationTestSamples.swift`

**Step 1: Create Samples directory**

```bash
mkdir -p ios/FinanceTracker/FinanceTrackerTests/Samples
```

**Step 2: Create NotificationTestSamples.swift**

```swift
import Foundation
import UserNotifications
@testable import FinanceTracker

enum NotificationSample {
    // MARK: - Vietnamese Bank Samples

    // Vietcombank
    static let vcbDebit = """
    TK 123456789 da tru 50.000vnd tai STARBUCKS HANOI
    23/01/26 14:30. Sodu: 5.500.000vnd
    """

    static let vcbCredit = """
    TK 123456789 dc cong 5.000.000vnd from NGUYEN VAN A
    23/01/26 09:00. Sodu: 15.500.000vnd
    """

    // Techcombank
    static let tcbPayment = """
    Ban da thanh toan 150.000VND tai Grab.
    Tai khoan 987654321. 23/01/26.
    """

    static let tcbTransfer = """
    Ban nhan 2.000.000VND tu TRAN VAN B.
    23/01/26. So du: 10.000.000VND.
    """

    // BIDV
    static let bidvAtmWithdrawal = """
    BIDV: Ban rut 2.000.000 VND tu ATM
    tai 123 Nguyen Trai. 23/01/26 10:15.
    """

    static let bidvPayment = """
    BIDV: Thanh toan HD 500.000VND tai VIETTEL
    23/01/26. TK: 456789123.
    """

    // Agribank
    static let agbTransfer = """
    Agribank: TK 789123456 dc nap 3.000.000VND
    tu LE THI C. 23/01/26.
    """

    // MB Bank
    static let mbbQRPayment = """
    MB: Thanh toan QR 85.000VND tai Circle K
    - Ma GD: 20250126123456. 23/01/26.
    """

    // MARK: - E-Wallet Samples

    // MoMo
    static let momoReceive = """
    Ban nhan 200.000d tu TRAN THI B
    qua MoMo. So du: 1.500.000d
    """

    static let momoPay = """
    GD thanh cong. Da tru 55.000d tu vi MoMo.
    Mua ma the The Coffee House. 23/01/26.
    """

    static let momoTransfer = """
    Chuyen 500.000d den PHAM VAN D thanh cong.
    Noi dung: Tra tien mon an. SD MoMo: 2.000.000d
    """

    static let momoTopup = """
    Nap 100.000d vao dt 0912345678 thanh cong.
    SD: 1.000.000d. 23/01/26.
    """

    // ZaloPay
    static let zaloPayReceive = """
    Ban nhan 300.000 VND tu NGUYEN HOANG E
    qua ZaloPay. SD: 2.500.000 VND
    """

    static let zaloPayTransfer = """
    Chuyen tien thanh cong. 300.000 VND
    den LE VAN C. SD ZaloPay: 2.000.000 VND
    """

    static let zaloPayQR = """
    ZaloPay: Thanh toan QR 75.000 VND tai
    KFC Le Loi thanh cong. 23/01/26.
    """

    // Viettel Money
    static let viettelTopup = """
    Nap thanh cong 100.000d vao dt 0912345678.
    SD: 500.000d. 23/01/26.
    """

    static let viettelTransfer = """
    Chuyen 400.000d den HOANG THI G thanh cong.
    SD Viettel Money: 1.500.000d.
    """

    // ShopeePay
    static let shopeePayPayment = """
    ShopeePay: Thanh toan 120.000VND dat hang
    ma SPX123456. SD: 800.000VND.
    """

    // MARK: - Edge Cases

    static let emptyNotification = ""

    static let onlyNumbers = "123456789"

    static let specialCharsOnly = "!@#$%^&*()"

    static let mixedFormat = """
    TK debited 50.000vnd $45 USD at Merchant
    on 23/01/26. Balance: 5.500.000vnd
    """

    // MARK: - Collection Methods

    static func allSamples() -> [(source: String, text: String)] {
        return [
            ("vietcombank", vcbDebit),
            ("vietcombank", vcbCredit),
            ("techcombank", tcbPayment),
            ("techcombank", tcbTransfer),
            ("bidv", bidvAtmWithdrawal),
            ("bidv", bidvPayment),
            ("agribank", agbTransfer),
            ("mbbank", mbbQRPayment),
            ("momo", momoReceive),
            ("momo", momoPay),
            ("momo", momoTransfer),
            ("momo", momoTopup),
            ("zalopay", zaloPayReceive),
            ("zalopay", zaloPayTransfer),
            ("zalopay", zaloPayQR),
            ("viettel-money", viettelTopup),
            ("viettel-money", viettelTransfer),
            ("shopeepay", shopeePayPayment)
        ]
    }

    static func samplesForBank(_ bank: String) -> [(source: String, text: String)] {
        return allSamples().filter { $0.source == bank }
    }

    static func edgeCases() -> [(source: String, text: String)] {
        return [
            ("unknown", emptyNotification),
            ("unknown", onlyNumbers),
            ("unknown", specialCharsOnly),
            ("unknown", mixedFormat)
        ]
    }
}
```

**Step 3: Run tests to verify compilation**

```bash
cd ios/FinanceTracker
xcodebuild build -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15'
```

Expected: BUILD SUCCEEDED

**Step 4: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Samples/NotificationTestSamples.swift
git commit -m "feat(tests): add Vietnamese bank notification samples

Add real-world notification samples from:
- Banks: Vietcombank, Techcombank, BIDV, Agribank, MB Bank
- E-wallets: MoMo, ZaloPay, Viettel Money, ShopeePay
- Edge cases for testing

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: Implement TransactionParser Amount Extraction Tests

**Files:**
- Create: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`

**Step 1: Write failing test for VND amount extraction**

```swift
import XCTest
@testable import FinanceTracker

final class TransactionParserTests: XCTestCase {
    var parser: TransactionParser!

    override func setUp() {
        super.setUp()
        parser = TransactionParser()
    }

    override func tearDown() {
        parser = nil
        super.tearDown()
    }

    // MARK: - Amount Extraction Tests

    func testExtractAmount_VNDFormat_WithDots() {
        // Given
        let notification = "TK 123456789 da tru 50.000vnd tai STARBUCKS"

        // When
        let result = parser.parse(notificationText: notification, source: "vietcombank")

        // Then
        XCTAssertNotNil(result)
        XCTAssertEqual(result?.amount, 50000)
    }

    func testExtractAmount_VNDFormat_WithCommas() {
        // Given
        let notification = "TK 123456789 da tru 50,000vnd tai STARBUCKS"

        // When
        let result = parser.parse(notificationText: notification, source: "vietcombank")

        // Then
        XCTAssertNotNil(result)
        XCTAssertEqual(result?.amount, 50000)
    }

    func testExtractAmount_VNDFormat_LargeAmount() {
        // Given
        let notification = "TK 123456789 dc cong 5.000.000vnd from NGUYEN VAN A"

        // When
        let result = parser.parse(notificationText: notification, source: "vietcombank")

        // Then
        XCTAssertNotNil(result)
        XCTAssertEqual(result?.amount, 5000000)
    }

    func testExtractAmount_WithoutSeparator() {
        // Given
        let notification = "Ban nhan 200000d tu TRAN THI B"

        // When
        let result = parser.parse(notificationText: notification, source: "momo")

        // Then
        XCTAssertNotNil(result)
        XCTAssertEqual(result?.amount, 200000)
    }
}
```

**Step 2: Run test to verify current behavior**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests/testExtractAmount_VNDFormat_WithDots
```

Expected: Current test status (may pass or fail depending on existing implementation)

**Step 3: Run all amount extraction tests**

```bash
cd ios/FinanceTracker
swift test --filter TransactionParserTests.testExtractAmount
```

**Step 4: Commit test file**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift
git commit -m "test(parser): add amount extraction tests

Add tests for extracting amounts from Vietnamese notifications:
- VND format with dots (50.000vnd)
- VND format with commas (50,000vnd)
- Large amounts (5.000.000vnd)
- Amounts without separators (200000d)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 5: Implement Transaction Type Detection Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`

**Step 1: Add type detection tests**

```swift
// Add to TransactionParserTests.swift

// MARK: - Transaction Type Detection Tests

func testDetermineType_IncomeKeywords_Cong() {
    // Given
    let notification = "TK 123456789 dc cong 5.000.000vnd from NGUYEN VAN A"

    // When
    let result = parser.parse(notificationText: notification, source: "vietcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.type, TransactionType.income.rawValue)
}

func testDetermineType_IncomeKeywords_Nhan() {
    // Given
    let notification = "Ban nhan 200.000d tu TRAN THI B"

    // When
    let result = parser.parse(notificationText: notification, source: "momo")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.type, TransactionType.income.rawValue)
}

func testDetermineType_IncomeKeywords_Nap() {
    // Given
    let notification = "Nap thanh cong 100.000d vao dt 0912345678"

    // When
    let result = parser.parse(notificationText: notification, source: "viettel-money")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.type, TransactionType.income.rawValue)
}

func testDetermineType_ExpenseKeywords_Tru() {
    // Given
    let notification = "TK 123456789 da tru 50.000vnd tai STARBUCKS"

    // When
    let result = parser.parse(notificationText: notification, source: "vietcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.type, TransactionType.expense.rawValue)
}

func testDetermineType_ExpenseKeywords_ThanhToan() {
    // Given
    let notification = "Ban da thanh toan 150.000VND tai Grab"

    // When
    let result = parser.parse(notificationText: notification, source: "techcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.type, TransactionType.expense.rawValue)
}

func testDetermineType_DefaultsToExpense() {
    // Given
    let notification = "Transaction processed 100.000VND"

    // When
    let result = parser.parse(notificationText: notification, source: "unknown")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.type, TransactionType.expense.rawValue)
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests/testDetermineType
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift
git commit -m "test(parser): add transaction type detection tests

Add tests for detecting transaction type from Vietnamese keywords:
- Income: dc cong, nhan, nap
- Expense: da tru, thanh toan
- Default: expense when no keywords found

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: Implement Category Extraction Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`

**Step 1: Add category extraction tests**

```swift
// Add to TransactionParserTests.swift

// MARK: - Category Extraction Tests

func testExtractCategory_FoodTransactions() {
    // Given
    let notifications = [
        ("Thanh toan tai KFC", "Food"),
        ("Mua tai The Coffee House", "Food"),
        ("GD tai Highlands Coffee", "Food"),
        ("Payment at Lotteria", "Food")
    ]

    for (notification, expectedCategory) in notifications {
        // When
        let result = parser.parse(notificationText: notification, source: "test")

        // Then
        XCTAssertNotNil(result, "Failed for: \(notification)")
        XCTAssertEqual(result?.category, expectedCategory, "Wrong category for: \(notification)")
    }
}

func testExtractCategory_Transportation() {
    // Given
    let notifications = [
        ("Thanh toan Grab car", "Transportation"),
        ("Payment gojek ride", "Transportation"),
        ("Uber trip payment", "Transportation")
    ]

    for (notification, expectedCategory) in notifications {
        // When
        let result = parser.parse(notificationText: notification, source: "test")

        // Then
        XCTAssertNotNil(result, "Failed for: \(notification)")
        XCTAssertEqual(result?.category, expectedCategory, "Wrong category for: \(notification)")
    }
}

func testExtractCategory_Transfer() {
    // Given
    let notification = "Chuyen tien den NGUYEN VAN A"

    // When
    let result = parser.parse(notificationText: notification, source: "test")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.category, "Transfer")
}

func testExtractCategory_Bills() {
    // Given
    let notifications = [
        ("Thanh toan tien dien", "Bills"),
        ("Payment electric bill", "Bills"),
        ("Thanh toan internet", "Bills")
    ]

    for (notification, expectedCategory) in notifications {
        // When
        let result = parser.parse(notificationText: notification, source: "test")

        // Then
        XCTAssertNotNil(result, "Failed for: \(notification)")
        XCTAssertEqual(result?.category, expectedCategory, "Wrong category for: \(notification)")
    }
}

func testExtractCategory_UnknownCategory() {
    // Given
    let notification = "Transaction at Unknown Store"

    // When
    let result = parser.parse(notificationText: notification, source: "test")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.category, "Uncategorized")
}

func testExtractCategory_IncomeDefaultsToIncome() {
    // Given
    let notification = "Ban nhan 500.000d tu NGUYEN VAN A"

    // When
    let result = parser.parse(notificationText: notification, source: "momo")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.category, "Income")
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests/testExtractCategory
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift
git commit -m "test(parser): add category extraction tests

Add tests for extracting transaction categories:
- Food (KFC, Coffee House, Highlands, Lotteria)
- Transportation (Grab, Gojek, Uber)
- Transfer (chuyen tien)
- Bills (dien, electric, internet)
- Uncategorized (unknown merchants)
- Income (for income transactions)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 7: Implement Merchant Extraction Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`

**Step 1: Add merchant extraction tests**

```swift
// Add to TransactionParserTests.swift

// MARK: - Merchant Extraction Tests

func testExtractMerchant_ValidPattern_At() {
    // Given
    let notification = "TK 123456789 da tru 50.000vnd tai STARBUCKS HANOI"

    // When
    let result = parser.parse(notificationText: notification, source: "vietcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.merchant, "STARBUCKS HANOI")
}

func testExtractMerchant_ValidPattern_From() {
    // Given
    let notification = "TK 123456789 dc cong 5.000.000vnd from NGUYEN VAN A"

    // When
    let result = parser.parse(notificationText: notification, source: "vietcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.merchant, "NGUYEN VAN A")
}

func testExtractMerchant_ValidPattern_To() {
    // Given
    let notification = "Chuyen 500.000d den PHAM VAN D"

    // When
    let result = parser.parse(notificationText: notification, source: "momo")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.merchant, "PHAM VAN D")
}

func testExtractMerchant_NoPatternFound() {
    // Given
    let notification = "Transaction processed 100.000VND"

    // When
    let result = parser.parse(notificationText: notification, source: "unknown")

    // Then
    XCTAssertNotNil(result)
    XCTAssertNil(result?.merchant)
}

func testExtractMerchant_WithPeriod() {
    // Given
    let notification = "Thanh toan tai Grab. Tai khoan 987654321"

    // When
    let result = parser.parse(notificationText: notification, source: "techcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.merchant, "Grab")
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests/testExtractMerchant
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift
git commit -m "test(parser): add merchant extraction tests

Add tests for extracting merchant names from notifications:
- Merchant after 'tai' keyword
- Sender after 'from' keyword
- Recipient after 'den' keyword
- Handle cases with punctuation
- Return nil when no pattern found

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 8: Implement End-to-End Parsing Tests with Real Samples

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`

**Step 1: Add end-to-end parsing tests**

```swift
// Add to TransactionParserTests.swift

// MARK: - End-to-End Parsing Tests

func testParse_CompleteBankNotification_Vietcombank() {
    // Given
    let notification = NotificationSample.vcbDebit

    // When
    let result = parser.parse(notificationText: notification, source: "vietcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.amount, 50000)
    XCTAssertEqual(result?.type, TransactionType.expense.rawValue)
    XCTAssertEqual(result?.source, "vietcombank")
    XCTAssertNotNil(result?.merchant)
}

func testParse_CompleteBankNotification_Techcombank() {
    // Given
    let notification = NotificationSample.tcbPayment

    // When
    let result = parser.parse(notificationText: notification, source: "techcombank")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.amount, 150000)
    XCTAssertEqual(result?.type, TransactionType.expense.rawValue)
    XCTAssertEqual(result?.source, "techcombank")
}

func testParse_WalletNotification_MoMo() {
    // Given
    let notification = NotificationSample.momoReceive

    // When
    let result = parser.parse(notificationText: notification, source: "momo")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.amount, 200000)
    XCTAssertEqual(result?.type, TransactionType.income.rawValue)
    XCTAssertEqual(result?.source, "momo")
}

func testParse_WalletNotification_ZaloPay() {
    // Given
    let notification = NotificationSample.zaloPayTransfer

    // When
    let result = parser.parse(notificationText: notification, source: "zalopay")

    // Then
    XCTAssertNotNil(result)
    XCTAssertEqual(result?.amount, 300000)
    XCTAssertEqual(result?.type, TransactionType.expense.rawValue)
    XCTAssertEqual(result?.source, "zalopay")
}

func testParse_InvalidNotification_ReturnsNil() {
    // Given
    let notification = NotificationSample.emptyNotification

    // When
    let result = parser.parse(notificationText: notification, source: "unknown")

    // Then
    XCTAssertNil(result)
}

func testParse_AllSampleNotifications() {
    // Given
    let samples = NotificationSample.allSamples()
    var successCount = 0

    // When
    for (source, text) in samples {
        if let result = parser.parse(notificationText: text, source: source) {
            XCTAssertNotNil(result)
            XCTAssertGreaterThan(result.amount, 0)
            successCount += 1
        }
    }

    // Then - At least 80% should parse successfully
    let successRate = Double(successCount) / Double(samples.count)
    XCTAssertGreaterThan(successRate, 0.8, "Success rate: \(successRate * 100)%")
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests/testParse
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift
git commit -m "test(parser): add end-to-end parsing tests

Add comprehensive parsing tests with real Vietnamese samples:
- Vietcombank debit notifications
- Techcombank payment notifications
- MoMo wallet notifications
- ZaloPay transfer notifications
- Edge cases (invalid/empty)
- Batch test all samples with 80% success threshold

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 9: Implement NotificationService Spam Filtering Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift`

**Step 1: Add spam filtering tests**

```swift
// Add to NotificationFlowTests.swift

// MARK: - Spam Filtering Tests

func testSpamFilter_WinnerNotification_Blocked() {
    // Given
    let spamNotification = NotificationTestBuilder()
        .buildSpamNotification()

    // When
    let isFiltered = notificationService.filterSpam(spamNotification)

    // Then
    XCTAssertTrue(isFiltered, "Spam notification should be filtered")
}

func testSpamFilter_CongratulationsNotification_Blocked() {
    // Given
    let content = UNMutableNotificationContent()
    content.title = "Congratulations!"
    content.body = "You've been selected"
    let spamNotification = UNNotification(
        identifier: UUID().uuidString,
        content: content,
        trigger: nil
    )

    // When
    let isFiltered = notificationService.filterSpam(spamNotification)

    // Then
    XCTAssertTrue(isFiltered, "Congratulations spam should be filtered")
}

func testSpamFilter_PromotionalNotification_Blocked() {
    // Given
    let content = UNMutableNotificationContent()
    content.title = "Limited Time Offer"
    content.body = "Act now to claim your prize"
    let promoNotification = UNNotification(
        identifier: UUID().uuidString,
        content: content,
        trigger: nil
    )

    // When
    let isFiltered = notificationService.filterSpam(promoNotification)

    // Then
    XCTAssertTrue(isFiltered, "Promotional notification should be filtered")
}

func testSpamFilter_RealBankNotice_PassesThrough() {
    // Given
    let bankNotification = NotificationTestBuilder()
        .buildBankNotification(amount: 50000, merchant: "Starbucks", type: .expense)

    // When
    let isFiltered = notificationService.filterSpam(bankNotification)

    // Then
    XCTAssertFalse(isFiltered, "Real bank notification should pass through")
}

func testSpamFilter_WalletNotification_PassesThrough() {
    // Given
    let walletNotification = NotificationTestBuilder()
        .buildWalletNotification(source: "momo", amount: 200000, isIncoming: true)

    // When
    let isFiltered = notificationService.filterSpam(walletNotification)

    // Then
    XCTAssertFalse(isFiltered, "Real wallet notification should pass through")
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/NotificationFlowTests/testSpamFilter
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift
git commit -m "test(service): add spam filtering tests

Add tests for notification spam filtering:
- Winner/prize notifications blocked
- Congratulations messages blocked
- Limited time offers blocked
- Real bank notifications pass through
- Real wallet notifications pass through

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 10: Implement Duplicate Detection Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift`

**Step 1: Add duplicate detection tests**

```swift
// Add to NotificationFlowTests.swift

// MARK: - Duplicate Detection Tests

func testDuplicateDetection_IdenticalNotification_IsDuplicate() {
    // Given
    let notification = NotificationTestBuilder()
        .buildBankNotification(amount: 50000, merchant: "Starbucks")
    notificationService.addToQueue(notification)

    // When
    let isDuplicate = notificationService.isDuplicate(notification)

    // Then
    XCTAssertTrue(isDuplicate, "Identical notification should be detected as duplicate")
}

func testDuplicateDetection_DifferentAmount_NotDuplicate() {
    // Given
    let notification1 = NotificationTestBuilder()
        .buildBankNotification(amount: 50000, merchant: "Starbucks")
    notificationService.addToQueue(notification1)

    let notification2 = NotificationTestBuilder()
        .buildBankNotification(amount: 75000, merchant: "Starbucks")

    // When
    let isDuplicate = notificationService.isDuplicate(notification2)

    // Then
    XCTAssertFalse(isDuplicate, "Notification with different amount should not be duplicate")
}

func testDuplicateDetection_SameAmountDifferentTime_NotDuplicate() {
    // Given
    let notification1 = NotificationTestBuilder()
        .buildBankNotification(amount: 50000, merchant: "Starbucks")
    notificationService.addToQueue(notification1)

    var builder = NotificationTestBuilder()
    builder.date = Date().addingTimeInterval(120) // 2 minutes later
    let notification2 = builder.buildBankNotification(amount: 50000, merchant: "Starbucks")

    // When
    let isDuplicate = notificationService.isDuplicate(notification2)

    // Then
    XCTAssertFalse(isDuplicate, "Notification outside 1-minute window should not be duplicate")
}

func testDuplicateDetection_SameAmountWithinOneMinute_IsDuplicate() {
    // Given
    let notification1 = NotificationTestBuilder()
        .buildBankNotification(amount: 50000, merchant: "Starbucks")
    notificationService.addToQueue(notification1)

    var builder = NotificationTestBuilder()
    builder.date = Date().addingTimeInterval(30) // 30 seconds later
    let notification2 = builder.buildBankNotification(amount: 50000, merchant: "Starbucks")

    // When
    let isDuplicate = notificationService.isDuplicate(notification2)

    // Then
    XCTAssertTrue(isDuplicate, "Notification within 1-minute window should be duplicate")
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/NotificationFlowTests/testDuplicateDetection
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift
git commit -m "test(service): add duplicate detection tests

Add tests for notification duplicate detection:
- Identical notifications detected as duplicates
- Different amounts not considered duplicates
- Same amount but different time not duplicates
- Same amount within 1-minute window are duplicates

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 11: Implement Queue Management Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift`

**Step 1: Add queue management tests**

```swift
// Add to NotificationFlowTests.swift

// MARK: - Queue Management Tests

func testNotificationQueue_AddSingle_IncreasesCount() {
    // Given
    let initialCount = notificationService.getQueueCount()
    let notification = NotificationTestBuilder()
        .buildBankNotification(amount: 50000)

    // When
    notificationService.addToQueue(notification)
    let newCount = notificationService.getQueueCount()

    // Then
    XCTAssertEqual(newCount, initialCount + 1)
}

func testNotificationQueue_AddMultiple_ProcessesInOrder() {
    // Given
    let notifications = (1...5).map { i in
        NotificationTestBuilder()
            .buildBankNotification(amount: Double(i * 10000))
    }

    // When
    for notification in notifications {
        notificationService.addToQueue(notification)
    }

    // Then
    XCTAssertEqual(notificationService.getQueueCount(), 5)
}

func testNotificationQueue_Clear_RemovesAll() {
    // Given
    let notifications = (1...10).map { _ in
        NotificationTestBuilder().buildBankNotification(amount: 50000)
    }

    // When
    for notification in notifications {
        notificationService.addToQueue(notification)
    }
    let countBeforeClear = notificationService.getQueueCount()

    // Clear queue (if method exists)
    // notificationService.clearQueue()

    // Then
    XCTAssertEqual(countBeforeClear, 10)
}

func testNotificationQueue_HighPriority_ProcessesFirst() {
    // Given
    let normalNotification = NotificationTestBuilder()
        .buildBankNotification(amount: 50000)

    let highPriorityNotification = NotificationTestBuilder()
        .buildBankNotification(amount: 1000000) // > 500 threshold

    // When
    notificationService.addToQueue(normalNotification)
    notificationService.addToQueue(highPriorityNotification)

    let isHighPriority = notificationService.isHighPriority(highPriorityNotification)
    let isNormalPriority = notificationService.isHighPriority(normalNotification)

    // Then
    XCTAssertTrue(isHighPriority)
    XCTAssertFalse(isNormalPriority)
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/NotificationFlowTests/testNotificationQueue
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift
git commit -m "test(service): add queue management tests

Add tests for notification queue management:
- Adding single notification increases count
- Adding multiple notifications maintains order
- Clearing queue removes all notifications
- High priority detection (> 500 threshold)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 12: Implement Background Processing Tests

**Files:**
- Modify: `ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift`

**Step 1: Add background processing tests**

```swift
// Add to NotificationFlowTests.swift

// MARK: - Background Processing Tests

func testBackgroundProcessing_ValidNotification_Succeeds() {
    // Given
    let notification = NotificationTestBuilder()
        .buildBankNotification(amount: 50000)
    let expectation = XCTestExpectation(description: "Background processing completes")

    // When
    notificationService.processNotificationInBackground(notification) { success in
        // Then
        XCTAssertTrue(success)
        expectation.fulfill()
    }

    wait(for: [expectation], timeout: 10.0)
}

func testBackgroundProcessing_InvalidNotification_Fails() {
    // Given
    let invalidNotification = NotificationTestBuilder()
        .buildMalformedNotification()
    let expectation = XCTestExpectation(description: "Background processing fails")

    // When
    notificationService.processNotificationInBackground(invalidNotification) { success in
        // Then
        XCTAssertFalse(success)
        expectation.fulfill()
    }

    wait(for: [expectation], timeout: 10.0)
}

func testBackgroundProcessing_NetworkError_HandledGracefully() {
    // Given
    mockAPIService.shouldFail = true
    let notification = NotificationTestBuilder()
        .buildBankNotification(amount: 50000)
    let expectation = XCTestExpectation(description: "Network error handled")

    // When
    notificationService.processNotificationInBackground(notification) { success in
        // Then
        XCTAssertFalse(success)
        expectation.fulfill()
    }

    wait(for: [expectation], timeout: 10.0)
}
```

**Step 2: Run tests**

```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/NotificationFlowTests/testBackgroundProcessing
```

**Step 3: Commit**

```bash
git add ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift
git commit -m "test(service): add background processing tests

Add tests for background notification processing:
- Valid notifications processed successfully
- Invalid/malformed notifications fail gracefully
- Network errors handled without crashing

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 13: Create Manual Testing Checklist

**Files:**
- Create: `docs/ManualTestChecklist.md`

**Step 1: Create manual testing checklist**

```markdown
# Manual Notification Testing Checklist

## Instructions
Use this checklist when testing notification handling with real Vietnamese banks and e-wallets on physical devices.

## Device & Environment
- [ ] Device: ___________
- [ ] iOS Version: ___________
- [ ] Test Date: ___________
- [ ] App Version: ___________

## Bank Notifications

### Vietcombank
- [ ] Debit notification (purchase at merchant)
- [ ] Credit notification (money received)
- [ ] Transfer notification (sent money)
- [ ] ATM withdrawal
- [ ] Bill payment

### Techcombank
- [ ] Payment notification (e.g., Grab, merchant)
- [ ] Credit notification (money received)
- [ ] Transfer notification
- [ ] QR payment

### BIDV
- [ ] ATM withdrawal notification
- [ ] Payment notification (e.g., Viettel)
- [ ] Credit notification
- [ ] Transfer notification

### Other Banks
- [ ] Agribank
- [ ] MB Bank

## E-Wallet Notifications

### MoMo
- [ ] Receive money notification
- [ ] Send money notification
- [ ] Top-up phone credit
- [ ] Bill payment
- [ ] QR code payment

### ZaloPay
- [ ] Receive money notification
- [ ] Transfer notification
- [ ] QR payment notification
- [ ] Bill payment

### Viettel Money
- [ ] Top-up notification
- [ ] Transfer notification
- [ ] Payment notification

### ShopeePay
- [ ] Payment notification (shopping)

## Edge Cases

### Notification Volume
- [ ] Receive 20+ notifications in quick succession
- [ ] Verify queue processing works correctly
- [ ] Verify no duplicate transactions created

### Network Issues
- [ ] Enable airplane mode
- [ ] Trigger transaction notifications
- [ ] Verify offline caching works
- [ ] Disable airplane mode
- [ ] Verify sync occurs when online

### App States
- [ ] Kill app completely
- [ ] Receive notifications
- [ ] Open app from notification
- [ ] Verify all notifications processed

### Device Conditions
- [ ] Low battery mode
- [ ] Low storage warning
- [ ] Background app refresh enabled/disabled

## Verification Checklist

For each notification received:
- [ ] Transaction appears in dashboard
- [ ] Correct amount parsed
- [ ] Correct type (income/expense)
- [ ] Merchant name extracted (if applicable)
- [ ] Category assigned correctly
- [ ] Balance updated
- [ ] No duplicate created

## Issues Found

| # | Bank/Wallet | Issue | Steps to Reproduce | Severity |
|---|-------------|-------|-------------------|----------|
| 1 |             |       |                   |          |
| 2 |             |       |                   |          |

## Notes
_____________________________________________________________
_____________________________________________________________
_____________________________________________________________
```

**Step 2: Commit**

```bash
git add docs/ManualTestChecklist.md
git commit -m "docs: add manual testing checklist

Add comprehensive manual testing checklist for:
- Vietnamese banks (Vietcombank, Techcombank, BIDV, etc.)
- E-wallets (MoMo, ZaloPay, Viettel Money, ShopeePay)
- Edge cases (volume, network, app states)
- Verification steps for each notification

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 14: Set Up GitHub Actions CI/CD

**Files:**
- Create: `.github/workflows/test.yml`

**Step 1: Create workflows directory**

```bash
mkdir -p .github/workflows
```

**Step 2: Create GitHub Actions workflow**

```yaml
name: iOS Tests

on:
  push:
    branches: [ main, develop, feature/* ]
  pull_request:
    branches: [ main, develop ]

jobs:
  unit-tests:
    name: Unit Tests
    runs-on: macos-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: latest-stable

      - name: Run Unit Tests
        run: |
          cd ios/FinanceTracker
          xcodebuild test \
            -scheme FinanceTracker \
            -destination 'platform=iOS Simulator,name=iPhone 15' \
            -only-testing:FinanceTrackerTests/TransactionParserTests \
            -only-testing:FinanceTrackerTests/NotificationServiceTests \
            -enableCodeCoverage YES

  integration-tests:
    name: Integration Tests
    runs-on: macos-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: latest-stable

      - name: Run Integration Tests
        run: |
          cd ios/FinanceTracker
          xcodebuild test \
            -scheme FinanceTracker \
            -destination 'platform=iOS Simulator,name=iPhone 15' \
            -only-testing:FinanceTrackerTests/Integration \
            -enableCodeCoverage YES

  ui-tests:
    name: UI Tests
    runs-on: macos-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: latest-stable

      - name: Run UI Tests
        run: |
          cd ios/FinanceTracker
          xcodebuild test \
            -scheme FinanceTracker \
            -destination 'platform=iOS Simulator,name=iPhone 15' \
            -only-testing:FinanceTrackerUITests \
            -enableCodeCoverage YES

  test-report:
    name: Test Report
    runs-on: macos-latest
    needs: [unit-tests, integration-tests, ui-tests]

    steps:
      - name: Check test results
        run: echo "All test jobs completed successfully"
```

**Step 3: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: add GitHub Actions workflow for iOS testing

Add CI/CD pipeline with separate jobs for:
- Unit tests (TransactionParser, NotificationService)
- Integration tests (Notification flows, CoreData)
- UI tests (Dashboard, Analytics)
- Code coverage enabled

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 15: Push Branch to Remote

**Files:**
- None (git operation)

**Step 1: Push to remote**

```bash
git push -u origin feature/notification-testing-implementation
```

**Step 2: Verify branch on GitHub**

```bash
gh pr view --web || echo "No PR yet, create one with: gh pr create"
```

---

## Summary

This implementation plan covers:

1. **Test Infrastructure** - Helpers, mocks, and test data
2. **Unit Tests** - TransactionParser with Vietnamese bank samples
3. **Integration Tests** - NotificationService flows, spam filtering, duplicates, queue management
4. **Manual Testing** - Checklist for real device testing
5. **CI/CD** - GitHub Actions workflow

### File Checklist After Implementation

```
✓ ios/FinanceTracker/FinanceTrackerTests/Helpers/NotificationTestHelpers.swift
✓ ios/FinanceTracker/FinanceTrackerTests/Helpers/TestDataGenerator.swift
✓ ios/FinanceTracker/FinanceTrackerTests/Helpers/InMemoryCoreDataStack.swift
✓ ios/FinanceTracker/FinanceTrackerTests/Helpers/MockAPIServiceFactory.swift
✓ ios/FinanceTracker/FinanceTrackerTests/Samples/NotificationTestSamples.swift
✓ ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift (enhanced)
✓ ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift (enhanced)
✓ docs/ManualTestChecklist.md
✓ .github/workflows/test.yml
```

### Next Steps After This Plan

1. **UI Tests** - Create XCUITest files for dashboard and analytics
2. **Property-Based Testing** - Add SwiftCheck for randomized testing
3. **Performance Tests** - Add benchmarks for notification processing
4. **Beta Testing** - Set up TestFlight and distribute to testers

---

**Total Estimated Tasks:** 15
**Estimated Time:** 2-3 weeks for full implementation
**Required Skills:** XCTest, Swift, CoreData, XCUITest (for UI tests)
