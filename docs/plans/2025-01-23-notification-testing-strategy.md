# Notification Testing Strategy for Finance Tracker

**Date:** 2025-01-23
**Status:** Approved
**Author:** Design Document

## Overview

This document outlines a comprehensive testing strategy for the Personal Finance Tracker's notification handling system. The app receives transaction notifications from Vietnamese banks (Vietcombank, Techcombank, BIDV) and e-wallets (MoMo, ZaloPay, Viettel Money), parses them, and stores transactions locally and remotely.

### Testing Goals

- **Data Accuracy**: Ensure transactions are parsed correctly from various notification formats
- **Integration Reliability**: Verify the notification flow works end-to-end
- **Edge Cases**: Handle malformed, duplicate, and spam notifications

---

## Architecture: Multi-Layer Testing Pyramid

```
                    ┌─────────────┐
                    │   Manual    │  As needed
                    │   Testing   │
                    └─────────────┘
                  ┌─────────────────┐
                  │     UI Tests    │  ~10%
                  └─────────────────┘
                ┌─────────────────────┐
                │  Integration Tests  │  ~20%
                └─────────────────────┘
              ┌───────────────────────────┐
              │        Unit Tests         │  ~70%
              └───────────────────────────┘
```

### Layer Breakdown

| Layer | Focus | Percentage | Execution Time |
|-------|-------|------------|----------------|
| Unit Tests | Individual components | 70% | Milliseconds |
| Integration Tests | Component interactions | 20% | Seconds |
| UI Tests | User-facing behavior | 10% | Seconds to minutes |
| Manual Testing | Real-world validation | As needed | N/A |

---

## Section 1: Unit Tests Layer

### Responsibility
Test `TransactionParser` in isolation - the critical component that converts notification text into structured transaction data.

### Test File Structure
```
FinanceTrackerTests/Services/TransactionParserTests.swift
FinanceTrackerTests/Samples/NotificationTestSamples.swift
```

### Test Cases

#### Amount Extraction Tests
```swift
func testExtractAmount_VNDFormat()
func testExtractAmount_USDFormat()
func testExtractAmount_WithDots()
func testExtractAmount_InvalidFormat_ReturnsNil()
```

#### Transaction Type Detection Tests
```swift
func testDetermineType_IncomeKeywords()
func testDetermineType_ExpenseKeywords()
func testDetermineType_DefaultsToExpense()
```

#### Category Extraction Tests
```swift
func testExtractCategory_FoodTransactions()
func testExtractCategory_Transportation()
func testExtractCategory_Transfer()
func testExtractCategory_UnknownCategory()
```

#### Merchant Extraction Tests
```swift
func testExtractMerchant_ValidPattern()
func testExtractMerchant_NoPatternFound()
```

#### End-to-End Parsing Tests
```swift
func testParse_CompleteBankNotification()
func testParse_WalletNotification()
func testParse_InvalidNotification_ReturnsNil()
```

### Notification Test Samples

Create `NotificationTestSamples.swift` with Vietnamese bank and wallet notification examples:

```swift
enum NotificationSample {
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

    // MoMo
    static let momoReceive = """
    Ban nhan 200.000d tu TRAN THI B
    qua MoMo. So du: 1.500.000d
    """

    static let momoPay = """
    GD thanh cong. Da tru 55.000d tu vi MoMo.
    Mua ma the The Coffee House. 23/01/26.
    """

    // ZaloPay
    static let zaloPayTransfer = """
    Chuyen tien thanh cong. 300.000 VND
    den LE VAN C. SD ZaloPay: 2.000.000 VND
    """

    // BIDV
    static let bidvAtmWithdrawal = """
    BIDV: Ban rut 2.000.000 VND tu ATM
    tai 123 Nguyen Trai. 23/01/26 10:15.
    """

    // Viettel Money
    static let viettelTopup = """
    Nap thanh cong 100.000d vao dt 0912345678.
    SD: 500.000d. 23/01/26.
    """
}
```

### Benefits
- Runs in milliseconds
- No external dependencies
- Easy to add new test cases
- Tests can run in parallel

---

## Section 2: Integration Tests Layer

### Responsibility
Verify that components work together correctly - `NotificationService`, `NotificationManager`, `APIService`, and `CoreData`.

### Test File Structure
```
FinanceTrackerTests/Integration/NotificationFlowTests.swift (extends existing)
FinanceTrackerTests/Integration/CoreDataIntegrationTests.swift
FinanceTrackerTests/Helpers/NotificationTestHelpers.swift
FinanceTrackerTests/Helpers/MockAPIServiceFactory.swift
```

### NotificationService Tests

#### Notification Processing Flow
```swift
func testProcessNotification_ValidNotification_SavesToCoreData()
func testProcessNotification_APIFailure_RetriesSuccessfully()
func testProcessNotification_DuplicateDetection_WorksCorrectly()
```

#### Queue Management
```swift
func testNotificationQueue_AddMultiple_ProcessesInOrder()
func testNotificationQueue_HighPriority_ProcessesFirst()
```

#### Spam Filtering
```swift
func testSpamFilter_RealBankNotices_PassThrough()
func testSpamFilter_PromotionalNotifications_Blocked()
```

#### Background Processing
```swift
func testBackgroundProcessing_AppInBackground_QueuesNotification()
func testBackgroundProcessing_AppComesForeground_SyncsToAPI()
```

### NotificationManager Tests

#### Delegate Handling
```swift
func testUNUserNotificationCenterDelegate_ReceivesNotification()
func testDelegate_ParsesWithTransactionParser_CreatesTransaction()
```

#### API Integration
```swift
func testAPIIntegration_CreateTransaction_Success()
func testAPIIntegration_NetworkError_StoresLocallyForRetry()
```

### CoreData Integration Tests
```swift
func testCoreData_TransactionSaved_CanBeRetrieved()
func testCoreData_TransactionDeleted_RemovedCorrectly()
```

### Mock Objects

```swift
class MockAPIService: APIServiceProtocol {
    var shouldFail = false
    var createTransactionCalled = false
    var receivedTransactions: [Transaction] = []

    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        createTransactionCalled = true
        receivedTransactions.append(transaction)
        if shouldFail { throw APIError.networkError }
        return transaction
    }
}

class InMemoryCoreDataStack {
    // Creates in-memory CoreData for fast tests
}
```

### Test Helpers

```swift
struct NotificationTestBuilder {
    func buildBankNotification(
        amount: Double,
        merchant: String? = nil,
        type: TransactionType = .expense
    ) -> UNNotification

    func buildWalletNotification(
        source: String,
        amount: Double,
        isIncoming: Bool
    ) -> UNNotification

    func buildMalformedNotification() -> UNNotification
    func buildSpamNotification() -> UNNotification
}
```

### Benefits
- Tests component interactions without external dependencies
- Catches integration bugs before production
- Moderate execution time (seconds)
- Easy to debug failures

---

## Section 3: UI Tests Layer

### Responsibility
Verify that notifications trigger the correct user-facing behavior and UI updates using XCUITest.

### Test File Structure
```
FinanceTrackerUITests/DashboardUITests.swift
FinanceTrackerUITests/AnalyticsUITests.swift
FinanceTrackerUITests/Helpers/UITestHelpers.swift
```

### Dashboard UI Tests

```swift
func testNewTransaction_AppearsInDashboard() {
    // Given: App is open on dashboard
    app.launch()

    // When: Simulate receiving a transaction notification
    simulateTransactionNotification(amount: 50000, merchant: "Starbucks")

    // Then: Transaction appears in recent transactions list
    XCTAssertTrue(app.staticTexts["Starbucks"].exists)
    XCTAssertTrue(app.staticTexts["50.000 VND"].exists)
}

func testTransactionUpdate_BalanceChanges() {
    // Given: App shows current balance
    let initialBalance = getDisplayedBalance()

    // When: Receive debit notification
    simulateTransactionNotification(amount: 100000, type: .expense)

    // Then: Balance is updated
    waitForBalanceUpdate()
    let newBalance = getDisplayedBalance()
    XCTAssertEqual(newBalance, initialBalance - 100000)
}
```

### Analytics UI Tests

```swift
func testMultipleTransactions_AnalyticsUpdatesCorrectly() {
    // Given: Dashboard is open
    app.launch()

    // When: Receive multiple food-related notifications
    simulateTransactionNotification(amount: 50000, category: "Food", merchant: "KFC")
    simulateTransactionNotification(amount: 75000, category: "Food", merchant: "Lotteria")
    simulateTransactionNotification(amount: 45000, category: "Food", merchant: "Highlands")

    // Then: Navigate to analytics and verify totals
    app.tabBars.buttons["Analytics"].tap()
    XCTAssertTrue(app.staticTexts["Food: 170.000 VND"].exists)
}
```

### Transaction Detail UI Tests

```swift
func testTapTransaction_ShowsDetailView() {
    // Given: Transaction in dashboard
    app.launch()
    simulateTransactionNotification(amount: 100000)

    // When: Tap on transaction
    app.buttons["TransactionCell"].tap()

    // Then: Detail view shows full information
    XCTAssertTrue(app.staticTexts["100.000 VND"].exists)
    XCTAssertTrue(app.staticTexts["Details"].exists)
}
```

### Background-to-Foreground Tests

```swift
func testBackgroundNotification_TapOpensAppToDashboard() {
    // Given: App is in background
    app.launch()
    XCUIDevice.shared.press(.home)

    // When: Receive notification and tap it
    sendPushNotificationWhileInBackground()
    tapNotification()

    // Then: App opens and shows the transaction
    XCTAssertTrue(app.state == .runningForeground)
    XCTAssertTrue(app.otherElements["DashboardView"].exists)
}
```

### Key Challenges & Solutions

| Challenge | Solution |
|-----------|----------|
| System notification dialogs | Special XCUITest handling or manual testing |
| Background processing | XCUITest lifecycle control or scheme configuration |
| Push notifications | Local notification simulation or APNs sandbox |

---

## Section 4: Manual/Device Testing Layer

### Responsibility
Validate actual integration with real bank/wallet apps on physical devices - scenarios automated tests cannot cover.

### Manual Testing Scenarios

#### 1. Real Bank Notification Testing

**Setup:**
- Install finance tracker on physical device
- Have accounts with target banks
- Enable notification permissions

**Test Cases:**
- Make a purchase → Verify transaction appears
- Transfer money → Verify debit/credit notifications captured
- Receive money → Verify income transaction recorded
- ATM withdrawal → Verify amount and location parsing
- Bill payment → Verify merchant and category detection

#### 2. E-Wallet Notification Testing

**Target Apps:** MoMo, ZaloPay, Viettel Money

**Test Cases:**
- Receive money → Verify incoming transaction recorded
- Send money → Verify outgoing transaction recorded
- Top up phone credit → Verify merchant detected
- Pay bills → Verify category assigned
- QR code payment → Verify merchant name extracted

#### 3. Edge Case Scenarios

**Notification Deluge:**
- Receive 20+ notifications in quick succession
- Verify queue processing and no duplicates

**Network Issues:**
- Enable airplane mode, trigger transactions
- Verify offline caching works
- Disable airplane mode → Verify sync occurs

**App States:**
- Kill app completely
- Receive notifications
- Open app → Verify all notifications processed

#### 4. Device-Specific Testing

**Testing Matrix:**
- Different iOS versions (16, 17, 18)
- Different device sizes (SE, iPhone 15, Pro Max)
- Low-end devices → Verify performance

### Manual Test Checklist

Create `docs/ManualTestChecklist.md`:

```markdown
# Manual Notification Testing Checklist

## Device & Environment
- [ ] Device: ___________
- [ ] iOS Version: ___________
- [ ] Test Date: ___________

## Bank Notifications
| Bank | Debit | Credit | Transfer | Bill Pay | ATM |
|------|-------|--------|----------|----------|-----|
| Vietcombank | [ ] | [ ] | [ ] | [ ] | [ ] |
| Techcombank | [ ] | [ ] | [ ] | [ ] | [ ] |
| BIDV | [ ] | [ ] | [ ] | [ ] | [ ] |

## E-Wallet Notifications
| Wallet | Receive | Send | Top-up | Bill Pay |
|--------|---------|------|--------|----------|
| MoMo | [ ] | [ ] | [ ] | [ ] |
| ZaloPay | [ ] | [ ] | [ ] | [ ] |
| Viettel Money | [ ] | [ ] | [ ] | [ ] |

## Edge Cases
- [ ] 20+ rapid notifications
- [ ] Offline mode transactions
- [ ] App killed state
- [ ] Low battery mode
- [ ] Low storage warning

## Issues Found
1. ___________________________
2. ___________________________
```

### Beta Testing Program

**Distribution:**
- Use TestFlight for controlled distribution
- Provide testing checklist
- Create in-app feedback button

**Feedback Collection:**
- Capture notification text
- Capture app's parsed result
- Device info and iOS version
- Steps to reproduce

---

## Section 5: Test Infrastructure and Tooling

### File Structure

```
FinanceTrackerTests/
├── Helpers/
│   ├── NotificationTestHelpers.swift
│   ├── TestDataGenerator.swift
│   └── InMemoryCoreDataStack.swift
├── Services/
│   ├── TransactionParserTests.swift
│   ├── NotificationServiceTests.swift
│   └── MockAPIServiceFactory.swift
├── Integration/
│   ├── NotificationFlowTests.swift
│   └── CoreDataIntegrationTests.swift
└── Samples/
    └── NotificationTestSamples.swift

FinanceTrackerUITests/
├── Helpers/
│   └── UITestHelpers.swift
├── DashboardUITests.swift
└── AnalyticsUITests.swift
```

### Core Test Utilities

#### NotificationTestBuilder.swift

```swift
struct NotificationTestBuilder {
    var source: String = "vietcombank"
    var amount: Double = 100000
    var type: TransactionType = .expense
    var merchant: String?
    var category: String?
    var date: Date = Date()

    func build() -> UNNotification {
        let content = UNMutableNotificationContent()
        content.title = "\(type == .income ? "Credit" : "Debit") Alert"
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
        // Returns realistic notification text based on params
    }
}
```

#### TestDataGenerator.swift

```swift
struct TransactionTestDataGenerator {
    static func randomTransaction() -> TransactionDTO
    static func randomTransactions(count: Int) -> [TransactionDTO]
    static func transactionForCategory(_ category: String) -> TransactionDTO

    // Property-based testing helpers
    static func generateValidAmounts() -> [Double]
    static func generateInvalidAmounts() -> [String]
    static func generateMalformedNotifications() -> [String]
}
```

#### InMemoryCoreDataStack.swift

```swift
class InMemoryCoreDataStack {
    static let shared = InMemoryCoreDataStack()

    lazy var viewContext: NSManagedObjectContext = {
        // Creates in-memory persistent store for fast, isolated tests
    }()

    func clearAllData() {
        // Wipes all data between tests
    }
}
```

### CI/CD Configuration

**.github/workflows/test.yml**

```yaml
name: iOS Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Unit Tests
        run: |
          xcodebuild test \
            -scheme FinanceTracker \
            -destination 'platform=iOS Simulator,name=iPhone 15' \
            -only-testing:FinanceTrackerTests/TransactionParserTests

  integration-tests:
    runs-on: macos-latest
    steps:
      - name: Run Integration Tests
        run: xcodebuild test -only-testing:FinanceTrackerTests/Integration

  ui-tests:
    runs-on: macos-latest
    steps:
      - name: Run UI Tests
        run: xcodebuild test -only-testing:FinanceTrackerUITests
```

### Development Workflow

1. **Write test first** (TDD approach)
2. **Run unit tests** during development (Cmd+U in Xcode)
3. **Commit triggers CI** for full test suite
4. **PR blocks on test failures**
5. **Manual testing** before release to production

---

## Section 6: Implementation Roadmap

### Phase 1: Foundation (Week 1)

- [ ] Set up test infrastructure (helpers, mocks, in-memory CoreData)
- [ ] Collect notification samples from banks/wallets
- [ ] Verify existing tests run successfully

### Phase 2: Unit Tests (Week 1-2)

- [ ] Expand TransactionParser tests with Vietnamese samples
- [ ] Add validation logic tests (spam, duplicates, priority)
- [ ] Test edge cases (malformed, empty, duplicate)

### Phase 3: Integration Tests (Week 2-3)

- [ ] Enhance NotificationService tests
- [ ] Add CoreData integration tests
- [ ] Add end-to-end flow tests

### Phase 4: UI Tests (Week 3-4)

- [ ] Set up UI test infrastructure
- [ ] Implement dashboard UI tests
- [ ] Implement analytics UI tests

### Phase 5: CI/CD & Automation (Week 4)

- [ ] Set up GitHub Actions workflow
- [ ] Add test coverage reporting
- [ ] Set minimum coverage thresholds (70%)

### Phase 6: Manual Testing (Week 5)

- [ ] Create manual testing checklist
- [ ] Set up TestFlight beta
- [ ] Conduct internal device testing

### Phase 7: Iteration & Improvement (Ongoing)

- [ ] Review test results and fix failures
- [ ] Expand test coverage for new banks/wallets
- [ ] Optimize slow tests

---

## File Checklist After Implementation

```
FinanceTrackerTests/
├── Helpers/
│   ├── NotificationTestHelpers.swift
│   ├── TestDataGenerator.swift
│   └── InMemoryCoreDataStack.swift
├── Services/
│   ├── TransactionParserTests.swift
│   ├── NotificationServiceTests.swift
│   └── MockAPIServiceFactory.swift
├── Integration/
│   ├── NotificationFlowTests.swift
│   └── CoreDataIntegrationTests.swift
└── Samples/
    └── NotificationTestSamples.swift

FinanceTrackerUITests/
├── Helpers/
│   └── UITestHelpers.swift
├── DashboardUITests.swift
└── AnalyticsUITests.swift

.github/
└── workflows/
    └── test.yml

docs/
├── plans/
│   └── 2025-01-23-notification-testing-strategy.md
└── ManualTestChecklist.md
```

---

## Conclusion

This testing strategy provides a comprehensive approach to validating notification handling from Vietnamese banks and e-wallets. By implementing tests at multiple levels - unit, integration, UI, and manual - we ensure:

1. **Data accuracy** through TransactionParser unit tests
2. **Integration reliability** through component interaction tests
3. **Edge case handling** through spam filtering, duplicate detection, and malformed notification tests
4. **Real-world validation** through manual testing on physical devices

The implementation roadmap provides a clear path forward, with each phase building on the previous one. The test infrastructure ensures tests are maintainable, fast, and reliable.

---

## References

- Existing code: `ios/FinanceTracker/FinanceTracker/Services/`
- Existing tests: `ios/FinanceTracker/FinanceTrackerTests/`
- UserNotifications framework documentation
- XCUITest documentation
