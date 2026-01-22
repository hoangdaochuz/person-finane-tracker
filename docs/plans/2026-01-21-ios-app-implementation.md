# iOS Finance Tracker App Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a native iOS application for tracking financial transactions from bank/e-wallet notifications with a modern fintech dashboard and analytics.

**Architecture:** MVVM (Model-View-ViewModel) with SwiftUI, backend-first data strategy, local CoreData caching, notification-based transaction capture via UNUserNotificationCenter, and Keychain for secure API key storage.

**Tech Stack:** Swift 5.9+, SwiftUI, iOS 16.0+, Swift Charts, Core Data, URLSession, Keychain Services, XCTest

---

## Prerequisites

- Xcode 15.0+
- macOS 14.0+
- iOS 16.0+ deployment target
- Backend API running (Go/Gin framework)

---

## Task 1: Xcode Project Setup

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/FinanceTrackerApp.swift`
- Create: `ios/FinanceTracker/FinanceTracker/ContentView.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Info.plist`
- Modify: `ios/FinanceTracker/FinanceTracker.xcodeproj/project.pbxproj`

### Step 1: Create Xcode project structure

Run:
```bash
cd ios/FinanceTracker
# If project doesn't exist, create via Xcode or use xcodegen
xcodegen generate 2>/dev/null || echo "Configure manually in Xcode"
```

### Step 2: Configure project settings

In Xcode:
- Set iOS Deployment Target to 16.0
- Enable SwiftUI (App lifecycle)
- Bundle Identifier: `com.financetracker.app`
- Disable "Core Data" checkbox (we'll add manually)

### Step 3: Create main App file

Create `ios/FinanceTracker/FinanceTracker/FinanceTrackerApp.swift`:

```swift
import SwiftUI

@main
struct FinanceTrackerApp: App {
    let persistenceController = CoreDataStack.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(\.managedObjectContext, persistenceController.container.viewContext)
        }
    }
}
```

### Step 4: Create placeholder ContentView

Create `ios/FinanceTracker/FinanceTracker/ContentView.swift`:

```swift
import SwiftUI

struct ContentView: View {
    var body: some View {
        Text("Finance Tracker")
            .padding()
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
```

### Step 5: Build to verify setup

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' clean build
```

---

## Task 2: Project Folder Structure and Groups

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Models/`
- Create: `ios/FinanceTracker/FinanceTracker/ViewModels/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Components/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Dashboard/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Analytics/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Transactions/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Auth/`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Settings/`
- Create: `ios/FinanceTracker/FinanceTracker/Services/`
- Create: `ios/FinanceTracker/FinanceTracker/Persistence/`

### Step 1: Create folder structure

Run:
```bash
cd ios/FinanceTracker/FinanceTracker
mkdir -p Models ViewModels Views/{Dashboard,Analytics,Transactions,Auth,Settings,Components} Services Persistence Resources
```

### Step 2: Add placeholder files to each directory

Run:
```bash
touch Models/.gitkeep ViewModels/.gitkeep
touch Views/{Dashboard,Analytics,Transactions,Auth,Settings,Components}/.gitkeep
touch Services/.gitkeep Persistence/.gitkeep Resources/.gitkeep
```

---

## Task 3: Core Data Models

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Persistence/CoreDataStack.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Persistence/Models/TransactionEntity+CoreDataClass.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Persistence/Models/TransactionEntity+CoreDataProperties.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Persistence/FinanceTracker.xcdatamodeld/`

### Step 1: Create CoreDataStack

Create `ios/FinanceTracker/FinanceTracker/Persistence/CoreDataStack.swift`:

```swift
import CoreData
import Foundation

class CoreDataStack {
    static let shared = CoreDataStack()

    lazy var container: NSPersistentContainer = {
        let container = NSPersistentContainer(name: "FinanceTracker")
        container.loadPersistentStores { _, error in
            if let error = error {
                fatalError("Core Data store failed to load: \(error)")
            }
        }
        return container
    }()

    var viewContext: NSManagedObjectContext {
        container.viewContext
    }

    func save() {
        let context = viewContext
        if context.hasChanges {
            do {
                try context.save()
            } catch {
                print("Failed to save Core Data: \(error)")
            }
        }
    }

    private init() {}
}
```

### Step 2: Create Core Data model file

Create `ios/FinanceTracker/FinanceTracker/Persistence/FinanceTracker.xcdatamodeld/FinanceTracker.xcdatamodel/contents`:

```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<model type="com.apple.IDECoreDataModeler.DataModel" documentVersion="1.0" lastSavedToolsVersion="22225" systemVersion="23A344" minimumToolsVersion="Automatic" sourceLanguage="Swift" userDefinedModelVersionIdentifier="">
    <entity name="TransactionEntity" representedClassName="TransactionEntity" syncable="YES">
        <attribute name="amount" optional="NO" attributeType="Double" defaultValueString="0.0" usesScalarValueType="YES"/>
        <attribute name="category" optional="YES" attributeType="String"/>
        <attribute name="date" optional="NO" attributeType="Date" usesScalarValueType="NO"/>
        <attribute name="id" optional="NO" attributeType="UUID" usesScalarValueType="NO"/>
        <attribute name="merchant" optional="YES" attributeType="String"/>
        <attribute name="remoteID" optional="YES" attributeType="String"/>
        <attribute name="source" optional="YES" attributeType="String"/>
        <attribute name="type" optional="NO" attributeType="String"/>
    </entity>
</model>
```

### Step 3: Create TransactionEntity classes

Create `ios/FinanceTracker/FinanceTracker/Persistence/Models/TransactionEntity+CoreDataProperties.swift`:

```swift
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
    @NSManaged public var source: String?
    @NSManaged public var type: String
}

extension TransactionEntity: Identifiable {}
```

Create `ios/FinanceTracker/FinanceTracker/Persistence/Models/TransactionEntity+CoreDataClass.swift`:

```swift
import Foundation
import CoreData

public class TransactionEntity: NSManagedObject {}
```

### Step 4: Build to verify Core Data setup

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 4: Swift Models (Domain Layer)

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Models/Transaction.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Models/TransactionType.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Models/Analytics.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Models/User.swift`

### Step 1: Create TransactionType enum

Create `ios/FinanceTracker/FinanceTracker/Models/TransactionType.swift`:

```swift
import Foundation

enum TransactionType: String, Codable, CaseIterable {
    case income = "income"
    case expense = "expense"
    case transfer = "transfer"

    var displayName: String {
        switch self {
        case .income: return "Income"
        case .expense: return "Expense"
        case .transfer: return "Transfer"
        }
    }
}
```

### Step 2: Create Transaction model

Create `ios/FinanceTracker/FinanceTracker/Models/Transaction.swift`:

```swift
import Foundation

struct Transaction: Identifiable, Codable, Equatable {
    let id: UUID
    var amount: Double
    var type: TransactionType
    var merchant: String?
    var category: String?
    var source: String?
    var date: Date
    var remoteID: String?

    init(
        id: UUID = UUID(),
        amount: Double,
        type: TransactionType,
        merchant: String? = nil,
        category: String? = nil,
        source: String? = nil,
        date: Date = Date(),
        remoteID: String? = nil
    ) {
        self.id = id
        self.amount = amount
        self.type = type
        self.merchant = merchant
        self.category = category
        self.source = source
        self.date = date
        self.remoteID = remoteID
    }
}
```

### Step 3: Create Analytics models

Create `ios/FinanceTracker/FinanceTracker/Models/Analytics.swift`:

```swift
import Foundation

struct Analytics: Codable, Equatable {
    let totalIncome: Double
    let totalExpenses: Double
    let balance: Double
    let categoryBreakdown: [CategorySummary]
    let sourceBreakdown: [SourceSummary]
    let period: TimePeriod

    var netSavings: Double {
        totalIncome - totalExpenses
    }
}

struct CategorySummary: Identifiable, Codable, Equatable {
    let id: UUID
    let category: String
    let amount: Double
    let transactionCount: Int
    let percentage: Double

    init(category: String, amount: Double, transactionCount: Int, totalAmount: Double) {
        self.id = UUID()
        self.category = category
        self.amount = amount
        self.transactionCount = transactionCount
        self.percentage = totalAmount > 0 ? (amount / totalAmount) * 100 : 0
    }
}

struct SourceSummary: Identifiable, Codable, Equatable {
    let id: UUID
    let source: String
    let amount: Double
    let transactionCount: Int

    init(source: String, amount: Double, transactionCount: Int) {
        self.id = UUID()
        self.source = source
        self.amount = amount
        self.transactionCount = transactionCount
    }
}

enum TimePeriod: String, Codable, CaseIterable {
    case week = "week"
    case month = "month"
    case threeMonths = "three_months"
    case year = "year"
    case all = "all"

    var displayName: String {
        switch self {
        case .week: return "This Week"
        case .month: return "This Month"
        case .threeMonths: return "3 Months"
        case .year: return "This Year"
        case .all: return "All Time"
        }
    }

    var days: Int? {
        switch self {
        case .week: return 7
        case .month: return 30
        case .threeMonths: return 90
        case .year: return 365
        case .all: return nil
        }
    }
}
```

### Step 4: Create User model

Create `ios/FinanceTracker/FinanceTracker/Models/User.swift`:

```swift
import Foundation

struct User: Codable, Equatable {
    let id: UUID
    var email: String
    var name: String?
    var apiKey: String?
    var isBiometricEnabled: Bool

    init(id: UUID = UUID(), email: String, name: String? = nil, apiKey: String? = nil, isBiometricEnabled: Bool = false) {
        self.id = id
        self.email = email
        self.name = name
        self.apiKey = apiKey
        self.isBiometricEnabled = isBiometricEnabled
    }
}
```

---

## Task 5: Keychain Service (Security)

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Services/KeychainManager.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Services/KeychainManagerTests.swift`

### Step 1: Write KeychainManager tests

Create `ios/FinanceTracker/FinanceTrackerTests/Services/KeychainManagerTests.swift`:

```swift
import XCTest
@testable import FinanceTracker

class KeychainManagerTests: XCTestCase {
    var sut: KeychainManager!
    let testKey = "test_api_key"
    let testValue = "test_secret_value_12345"

    override func setUp() {
        super.setUp()
        sut = KeychainManager()
        sut.delete(key: testKey)
    }

    override func tearDown() {
        sut.delete(key: testKey)
        super.tearDown()
    }

    func testStoreAndRetrieveApiKey() throws {
        let storeResult = sut.store(key: testKey, value: testValue)
        XCTAssertTrue(storeResult)

        let retrievedValue = sut.retrieve(key: testKey)
        XCTAssertEqual(retrievedValue, testValue)
    }

    func testRetrieveNonExistentKeyReturnsNil() {
        let value = sut.retrieve(key: "nonexistent_key")
        XCTAssertNil(value)
    }

    func testDeleteKey() {
        sut.store(key: testKey, value: testValue)
        XCTAssertNotNil(sut.retrieve(key: testKey))

        let deleteResult = sut.delete(key: testKey)
        XCTAssertTrue(deleteResult)

        XCTAssertNil(sut.retrieve(key: testKey))
    }

    func testUpdateExistingKey() {
        let initialValue = "initial_value"
        sut.store(key: testKey, value: initialValue)

        let newValue = "updated_value"
        sut.store(key: testKey, value: newValue)

        let retrievedValue = sut.retrieve(key: testKey)
        XCTAssertEqual(retrievedValue, newValue)
    }
}
```

### Step 2: Run tests to verify they fail

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/KeychainManagerTests
```

### Step 3: Implement KeychainManager

Create `ios/FinanceTracker/FinanceTracker/Services/KeychainManager.swift`:

```swift
import Foundation
import Security

class KeychainManager {
    private let service = "com.financetracker.app"

    func store(key: String, value: String) -> Bool {
        guard let data = value.data(using: .utf8) else { return false }

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data
        ]

        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)
        return status == errSecSuccess
    }

    func retrieve(key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let value = String(data: data, encoding: .utf8) else {
            return nil
        }

        return value
    }

    func delete(key: String) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]

        let status = SecItemDelete(query as CFDictionary)
        return status == errSecSuccess || status == errSecItemNotFound
    }
}
```

### Step 4: Run tests to verify they pass

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/KeychainManagerTests
```

---

## Task 6: API Service (Networking Layer)

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Services/APIService.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Services/APIEndpoint.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Models/APIResponse.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Services/APIServiceTests.swift`

### Step 1: Create APIEndpoint enum

Create `ios/FinanceTracker/FinanceTracker/Services/APIEndpoint.swift`:

```swift
import Foundation

enum APIEndpoint {
    case login(email: String, password: String)
    case register(email: String, password: String)
    case createTransaction(Transaction)
    case getTransactions(page: Int, limit: Int)
    case getAnalytics(period: TimePeriod)
    case getSummary

    var path: String {
        switch self {
        case .login: return "/api/v1/auth/login"
        case .register: return "/api/v1/auth/register"
        case .createTransaction: return "/api/v1/transactions"
        case .getTransactions: return "/api/v1/transactions"
        case .getAnalytics: return "/api/v1/analytics"
        case .getSummary: return "/api/v1/summary"
        }
    }

    var method: String {
        switch self {
        case .login, .register, .createTransaction:
            return "POST"
        case .getTransactions, .getAnalytics, .getSummary:
            return "GET"
        }
    }
}
```

### Step 2: Create API response models

Create `ios/FinanceTracker/FinanceTracker/Models/APIResponse.swift`:

```swift
import Foundation

struct LoginResponse: Codable {
    let apiKey: String
    let user: User
}

struct TransactionsResponse: Codable {
    let transactions: [Transaction]
    let page: Int
    let limit: Int
    let total: Int
}

struct CreateTransactionResponse: Codable {
    let transaction: Transaction
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
```

### Step 3: Write APIService tests

Create `ios/FinanceTracker/FinanceTrackerTests/Services/APIServiceTests.swift`:

```swift
import XCTest
@testable import FinanceTracker

class MockURLSession: URLSessionProtocol {
    var data: Data?
    var response: URLResponse?
    var error: Error?

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        if let error = error {
            throw error
        }
        guard let data = data, let response = response else {
            throw URLError(.badServerResponse)
        }
        return (data, response)
    }
}

protocol URLSessionProtocol {
    func data(for request: URLRequest) async throws -> (Data, URLResponse)
}

extension URLSession: URLSessionProtocol {}

class APIServiceTests: XCTestCase {
    var sut: APIService!
    var mockSession: MockURLSession!

    override func setUp() {
        super.setUp()
        mockSession = MockURLSession()
        sut = APIService(session: mockSession, baseURL: "https://api.test.com")
    }

    func testLoginSuccess() async throws {
        let expectedUser = User(email: "test@example.com", apiKey: "test_api_key")
        let loginResponse = LoginResponse(apiKey: "test_api_key", user: expectedUser)
        mockSession.data = try JSONEncoder().encode(loginResponse)
        mockSession.response = HTTPURLResponse(url: URL(string: "https://api.test.com")!, statusCode: 200, httpVersion: nil, headerFields: nil)!

        let result = try await sut.login(email: "test@example.com", password: "password123")

        XCTAssertEqual(result.email, expectedUser.email)
        XCTAssertEqual(result.apiKey, expectedUser.apiKey)
    }

    func testLoginFailure() async {
        mockSession.error = URLError(.notConnectedToInternet)

        do {
            _ = try await sut.login(email: "test@example.com", password: "password123")
            XCTFail("Expected error to be thrown")
        } catch {
            XCTAssertNotNil(error)
        }
    }
}
```

### Step 4: Run tests to verify they fail

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/APIServiceTests 2>&1 | head -50
```

### Step 5: Implement APIService

Create `ios/FinanceTracker/FinanceTracker/Services/APIService.swift`:

```swift
import Foundation

class APIService {
    private let session: URLSessionProtocol
    private let baseURL: String
    private let keychainManager: KeychainManager

    var apiKey: String? {
        keychainManager.retrieve(key: "finance_tracker_api_key")
    }

    init(session: URLSessionProtocol = URLSession.shared, baseURL: String, keychainManager: KeychainManager = KeychainManager()) {
        self.session = session
        self.baseURL = baseURL
        self.keychainManager = keychainManager
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
        var request = createAuthenticatedRequest(for: .createTransaction(transaction))
        request.httpBody = try JSONEncoder().encode(transaction)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let (data, response) = try await session.data(for: request)
        let createResponse: CreateTransactionResponse = try handleResponse(data, response: response)
        return createResponse.transaction
    }

    func getTransactions(page: Int = 1, limit: Int = 20) async throws -> [Transaction] {
        let request = createAuthenticatedRequest(for: .getTransactions(page: page, limit: limit))
        let (data, response) = try await session.data(for: request)
        let result: TransactionsResponse = try handleResponse(data, response: response)
        return result.transactions
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
```

### Step 6: Run tests to verify they pass

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/APIServiceTests
```

---

## Task 7: AuthManager (Authentication Service)

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Services/AuthManager.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Services/AuthManagerTests.swift`

### Step 1: Write AuthManager tests

Create `ios/FinanceTracker/FinanceTrackerTests/Services/AuthManagerTests.swift`:

```swift
import XCTest
@testable import FinanceTracker

class MockAPIServiceForAuth: APIServiceProtocol {
    var mockUser: User?
    var shouldThrowError = false

    func login(email: String, password: String) async throws -> User {
        if shouldThrowError {
            throw APIError.unauthorized
        }
        return mockUser ?? User(email: email, apiKey: "mock_api_key")
    }

    func register(email: String, password: String) async throws -> User {
        if shouldThrowError {
            throw APIError.clientError("Registration failed")
        }
        return User(email: email, apiKey: "mock_api_key")
    }
}

protocol APIServiceProtocol {
    func login(email: String, password: String) async throws -> User
    func register(email: String, password: String) async throws -> User
}

class AuthManagerTests: XCTestCase {
    var sut: AuthManager!
    var mockAPI: MockAPIServiceForAuth!
    var mockKeychain: KeychainManager!

    override func setUp() {
        super.setUp()
        mockAPI = MockAPIServiceForAuth()
        mockKeychain = KeychainManager()
        mockKeychain.delete(key: "finance_tracker_api_key")
        sut = AuthManager(apiService: mockAPI, keychainManager: mockKeychain)
    }

    override func tearDown() {
        mockKeychain.delete(key: "finance_tracker_api_key")
        super.tearDown()
    }

    func testLoginSuccess() async throws {
        let email = "test@example.com"
        let password = "password123"

        try await sut.login(email: email, password: password)

        let storedKey = mockKeychain.retrieve(key: "finance_tracker_api_key")
        XCTAssertEqual(storedKey, "mock_api_key")
        XCTAssertTrue(sut.isAuthenticated)
    }

    func testLoginFailure() async {
        mockAPI.shouldThrowError = true

        do {
            try await sut.login(email: "test@example.com", password: "wrong")
            XCTFail("Expected error")
        } catch {
            // Expected
        }

        XCTAssertFalse(sut.isAuthenticated)
    }

    func testLogout() async throws {
        try await sut.login(email: "test@example.com", password: "password123")
        XCTAssertTrue(sut.isAuthenticated)

        sut.logout()

        let storedKey = mockKeychain.retrieve(key: "finance_tracker_api_key")
        XCTAssertNil(storedKey)
        XCTAssertFalse(sut.isAuthenticated)
    }
}
```

### Step 2: Run tests to verify they fail

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/AuthManagerTests 2>&1 | head -30
```

### Step 3: Implement AuthManager

Create `ios/FinanceTracker/FinanceTracker/Services/AuthManager.swift`:

```swift
import Foundation
import Combine

class AuthManager: ObservableObject {
    @Published var isAuthenticated = false
    @Published var currentUser: User?

    private let apiService: APIServiceProtocol
    private let keychainManager: KeychainManager

    init(apiService: APIServiceProtocol, keychainManager: KeychainManager = KeychainManager()) {
        self.apiService = apiService
        self.keychainManager = keychainManager
        checkAuthStatus()
    }

    func login(email: String, password: String) async throws {
        let user = try await apiService.login(email: email, password: password)

        if let apiKey = user.apiKey {
            keychainManager.store(key: "finance_tracker_api_key", value: apiKey)
        }

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
        }
    }

    func register(email: String, password: String) async throws {
        let user = try await apiService.register(email: email, password: password)

        if let apiKey = user.apiKey {
            keychainManager.store(key: "finance_tracker_api_key", value: apiKey)
        }

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
        }
    }

    func logout() {
        keychainManager.delete(key: "finance_tracker_api_key")
        currentUser = nil
        isAuthenticated = false
    }

    private func checkAuthStatus() {
        if keychainManager.retrieve(key: "finance_tracker_api_key") != nil {
            isAuthenticated = true
        }
    }
}
```

### Step 4: Run tests to verify they pass

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/AuthManagerTests
```

---

## Task 8: Transaction Parser (Notification Parsing)

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Services/TransactionParser.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Resources/NotificationPatterns.plist`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`

### Step 1: Create NotificationPatterns.plist

Create `ios/FinanceTracker/FinanceTracker/Resources/NotificationPatterns.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Patterns</key>
    <dict>
        <key>Amount</key>
        <array>
            <string>(?:Rp|IDR|USD|\$)?\s*([\d,]+\.?\d*)</string>
        </array>
        <key>Merchant</key>
        <array>
            <string>(?:at|from|to)\s+([A-Z][A-Za-z\s]+?)(?:\s+on|\s*$|\.)</string>
        </array>
        <key>IncomeKeywords</key>
        <array>
            <string>received</string>
            <string>credit</string>
            <string>deposit</string>
            <string>transfer from</string>
            <string>incoming</string>
        </array>
        <key>ExpenseKeywords</key>
        <array>
            <string>spent</string>
            <string>payment</string>
            <string>purchase</string>
            <string>debit</string>
            <string>transfer to</string>
        </array>
    </dict>
    <key>KnownSources</key>
    <array>
        <string>BCA</string>
        <string>Mandiri</string>
        <string>BNI</string>
        <string>BRI</string>
        <string>Gopay</string>
        <string>OVO</string>
        <string>Dana</string>
        <string>Jenius</string>
    </array>
</dict>
</plist>
```

### Step 2: Write TransactionParser tests

Create `ios/FinanceTracker/FinanceTrackerTests/Services/TransactionParserTests.swift`:

```swift
import XCTest
@testable import FinanceTracker

class TransactionParserTests: XCTestCase {
    var sut: TransactionParser!

    override func setUp() {
        super.setUp()
        sut = TransactionParser()
    }

    func testParseExpenseTransaction() throws {
        let text = "You spent Rp 50.000 at Coffee Shop on Jan 21"
        let result = sut.parse(notificationText: text, source: "BCA")

        XCTAssertEqual(result?.amount, 50000.0)
        XCTAssertEqual(result?.type, .expense)
        XCTAssertEqual(result?.merchant, "Coffee Shop")
    }

    func testParseIncomeTransaction() throws {
        let text = "You received Rp 1.500.000 from John Doe"
        let result = sut.parse(notificationText: text, source: "Mandiri")

        XCTAssertEqual(result?.amount, 1500000.0)
        XCTAssertEqual(result?.type, .income)
        XCTAssertEqual(result?.merchant, "John Doe")
    }

    func testParseWithDecimalAmount() throws {
        let text = "Payment of $15.99 at Amazon"
        let result = sut.parse(notificationText: text, source: "Gopay")

        XCTAssertEqual(result?.amount, 15.99)
    }

    func testUnparseableTextReturnsNil() {
        let text = "Hello world, this is not a transaction"
        let result = sut.parse(notificationText: text, source: "Unknown")

        XCTAssertNil(result)
    }
}
```

### Step 3: Run tests to verify they fail

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests 2>&1 | head -30
```

### Step 4: Implement TransactionParser

Create `ios/FinanceTracker/FinanceTracker/Services/TransactionParser.swift`:

```swift
import Foundation

class TransactionParser {
    private var patterns: [String: Any] = [:]

    init() {
        loadPatterns()
    }

    func parse(notificationText: String, source: String) -> Transaction? {
        guard let amount = extractAmount(from: notificationText) else {
            return nil
        }

        let type = determineType(from: notificationText)
        let merchant = extractMerchant(from: notificationText)

        return Transaction(
            amount: amount,
            type: type,
            merchant: merchant,
            source: source,
            date: Date(),
            remoteID: nil
        )
    }

    private func loadPatterns() {
        guard let path = Bundle.main.path(forResource: "NotificationPatterns", ofType: "plist"),
              let plist = NSDictionary(contentsOfFile: path),
              let patternsDict = plist["Patterns"] as? [String: Any] else {
            return
        }
        patterns = patternsDict
    }

    private func extractAmount(from text: String) -> Double? {
        let pattern = "(?:Rp|IDR|USD|\\$)?\\s*([\\d,]+\\.?\\d*)"

        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
              let range = Range(match.range(at: 1), in: text) else {
            return nil
        }

        let amountString = String(text[range])
            .replacingOccurrences(of: ",", with: "")
        return Double(amountString)
    }

    private func determineType(from text: String) -> TransactionType {
        let lowercaseText = text.lowercased()

        if let incomeKeywords = patterns["IncomeKeywords"] as? [String] {
            for keyword in incomeKeywords where lowercaseText.contains(keyword.lowercased()) {
                return .income
            }
        }

        if let expenseKeywords = patterns["ExpenseKeywords"] as? [String] {
            for keyword in expenseKeywords where lowercaseText.contains(keyword.lowercased()) {
                return .expense
            }
        }

        return .expense
    }

    private func extractMerchant(from text: String) -> String? {
        let pattern = "(?:at|from|to)\\s+([A-Z][A-Za-z\\s]+?)(?:\\s+on|\\s*$|\\.)"

        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
              let range = Range(match.range(at: 1), in: text) else {
            return nil
        }

        return String(text[range]).trimmingCharacters(in: .whitespaces)
    }
}
```

### Step 5: Run tests to verify they pass

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/TransactionParserTests
```

---

## Task 9: NotificationManager (Background Notifications)

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Services/NotificationManager.swift`
- Create: `ios/FinanceTracker/FinanceTracker/AppDelegate.swift`

### Step 1: Create NotificationManager

Create `ios/FinanceTracker/FinanceTracker/Services/NotificationManager.swift`:

```swift
import Foundation
import UserNotifications

class NotificationManager: NSObject, ObservableObject {
    private let transactionParser: TransactionParser
    private let apiService: APIService
    private let coreDataStack: CoreDataStack

    var onNewTransaction: ((Transaction) -> Void)?

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

        guard let transaction = transactionParser.parse(notificationText: body, source: source) else {
            print("Failed to parse transaction from notification")
            return
        }

        saveTransactionToCache(transaction)

        Task {
            do {
                let savedTransaction = try await apiService.createTransaction(transaction)
                await MainActor.run {
                    onNewTransaction?(savedTransaction)
                }
            } catch {
                print("Failed to sync transaction to API: \(error)")
            }
        }
    }

    private func saveTransactionToCache(_ transaction: Transaction) {
        let entity = TransactionEntity(context: coreDataStack.viewContext)
        entity.id = transaction.id
        entity.amount = transaction.amount
        entity.type = transaction.type.rawValue
        entity.merchant = transaction.merchant
        entity.category = transaction.category
        entity.source = transaction.source
        entity.date = transaction.date
        entity.remoteID = transaction.remoteID

        coreDataStack.save()
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
```

### Step 2: Create Config file

Create `ios/FinanceTracker/FinanceTracker/Resources/Config.swift`:

```swift
import Foundation

enum Config {
    static let baseURL = "https://api.financetracker.app"
    #if DEBUG
    static let baseURL = "http://localhost:8080"
    #endif
}
```

### Step 3: Create AppDelegate

Create `ios/FinanceTracker/FinanceTracker/AppDelegate.swift`:

```swift
import UIKit
import UserNotifications

class AppDelegate: NSObject, UIApplicationDelegate {
    let notificationManager = NotificationManager()

    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        UNUserNotificationCenter.current().delegate = notificationManager
        application.registerForRemoteNotifications()
        return true
    }

    func application(
        _ application: UIApplication,
        didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
    ) {
        let tokenParts = deviceToken.map { data in String(format: "%02.2hhx", data) }
        let token = tokenParts.joined()
        print("Device Token: \(token)")
    }

    func application(
        _ application: UIApplication,
        didFailToRegisterForRemoteNotificationsWithError error: Error
    ) {
        print("Failed to register for remote notifications: \(error)")
    }
}
```

### Step 4: Update FinanceTrackerApp to use AppDelegate

Modify `ios/FinanceTracker/FinanceTracker/FinanceTrackerApp.swift`:

```swift
import SwiftUI

@main
struct FinanceTrackerApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    let persistenceController = CoreDataStack.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(\.managedObjectContext, persistenceController.container.viewContext)
        }
    }
}
```

### Step 5: Build to verify compilation

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 10: DashboardViewModel

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/ViewModels/DashboardViewModel.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/ViewModels/DashboardViewModelTests.swift`

### Step 1: Write DashboardViewModel tests

Create `ios/FinanceTracker/FinanceTrackerTests/ViewModels/DashboardViewModelTests.swift`:

```swift
import XCTest
@testable import FinanceTracker

class MockAPIServiceForDashboard: APIServiceProtocol {
    var mockTransactions: [Transaction] = []
    var mockSummary: SummaryResponse?
    var shouldThrowError = false

    func getTransactions(page: Int, limit: Int) async throws -> [Transaction] {
        if shouldThrowError { throw APIError.serverError }
        return Array(mockTransactions.prefix(limit))
    }

    func getAnalytics(period: TimePeriod) async throws -> Analytics {
        if shouldThrowError { throw APIError.serverError }
        return Analytics(
            totalIncome: 5000,
            totalExpenses: 3000,
            balance: 2000,
            categoryBreakdown: [],
            sourceBreakdown: [],
            period: period
        )
    }

    func getSummary() async throws -> SummaryResponse {
        if shouldThrowError { throw APIError.serverError }
        return mockSummary ?? SummaryResponse(balance: 2000, totalIncome: 5000, totalExpenses: 3000)
    }
}

class DashboardViewModelTests: XCTestCase {
    var sut: DashboardViewModel!
    var mockAPI: MockAPIServiceForDashboard!

    override func setUp() {
        super.setUp()
        mockAPI = MockAPIServiceForDashboard()
        mockAPI.mockTransactions = [
            Transaction(amount: 100, type: .income, merchant: "Salary", source: "BCA", date: Date()),
            Transaction(amount: 50, type: .expense, merchant: "Coffee", source: "Gopay", date: Date())
        ]
        mockAPI.mockSummary = SummaryResponse(balance: 2000, totalIncome: 5000, totalExpenses: 3000)
        sut = DashboardViewModel(apiService: mockAPI)
    }

    func testLoadData() async throws {
        await sut.loadData()

        XCTAssertEqual(sut.recentTransactions.count, 2)
        XCTAssertEqual(sut.balance, 2000)
        XCTAssertEqual(sut.totalIncome, 5000)
        XCTAssertEqual(sut.totalExpenses, 3000)
    }

    func testInitialState() {
        XCTAssertTrue(sut.isLoading)
        XCTAssertNil(sut.errorMessage)
    }

    func testErrorHandling() async {
        mockAPI.shouldThrowError = true
        await sut.loadData()

        XCTAssertNotNil(sut.errorMessage)
        XCTAssertFalse(sut.isLoading)
    }
}
```

### Step 2: Run tests to verify they fail

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/DashboardViewModelTests 2>&1 | head -30
```

### Step 3: Implement DashboardViewModel

Create `ios/FinanceTracker/FinanceTracker/ViewModels/DashboardViewModel.swift`:

```swift
import Foundation
import Combine

@MainActor
class DashboardViewModel: ObservableObject {
    @Published var recentTransactions: [Transaction] = []
    @Published var balance: Double = 0
    @Published var totalIncome: Double = 0
    @Published var totalExpenses: Double = 0
    @Published var isLoading = true
    @Published var errorMessage: String?

    private let apiService: APIServiceProtocol
    private var cancellables = Set<AnyCancellable>()

    init(apiService: APIServiceProtocol) {
        self.apiService = apiService
    }

    func loadData() async {
        isLoading = true
        errorMessage = nil

        do {
            let transactions = try await apiService.getTransactions(page: 1, limit: 5)
            self.recentTransactions = transactions

            let summary = try await apiService.getSummary()
            self.balance = summary.balance
            self.totalIncome = summary.totalIncome
            self.totalExpenses = summary.totalExpenses

            isLoading = false
        } catch {
            isLoading = false
            errorMessage = error.localizedDescription
        }
    }

    func refresh() async {
        await loadData()
    }
}
```

### Step 4: Run tests to verify they pass

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/DashboardViewModelTests
```

---

## Task 11: Design System - Colors and Typography

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Resources/ColorPalette.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Resources/Typography.swift`

### Step 1: Create ColorPalette

Create `ios/FinanceTracker/FinanceTracker/Resources/ColorPalette.swift`:

```swift
import SwiftUI

enum ColorPalette {
    // Primary Gradient Colors
    static let primaryIndigo = Color(red: 0.39, green: 0.4, blue: 0.95)
    static let primaryViolet = Color(red: 0.55, green: 0.36, blue: 0.96)

    // Status Colors
    static let success = Color(red: 0.06, green: 0.73, blue: 0.51)
    static let danger = Color(red: 0.94, green: 0.27, blue: 0.27)
    static let warning = Color(red: 1.0, green: 0.73, blue: 0.0)

    // Background Colors
    static let background = Color(red: 0.98, green: 0.98, blue: 0.98)
    static let cardBackground = Color.white

    // Text Colors
    static let textPrimary = Color(red: 0.07, green: 0.09, blue: 0.15)
    static let textSecondary = Color(red: 0.42, green: 0.45, blue: 0.44)

    // Gradient
    static let primaryGradient = LinearGradient(
        colors: [primaryIndigo, primaryViolet],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    // Semantic colors
    static let income = success
    static let expense = danger
}
```

### Step 2: Create Typography

Create `ios/FinanceTracker/FinanceTracker/Resources/Typography.swift`:

```swift
import SwiftUI

enum Typography {
    // Headings
    static let largeTitle = Font.system(size: 28, weight: .bold)
    static let title1 = Font.system(size: 24, weight: .bold)
    static let title2 = Font.system(size: 20, weight: .bold)
    static let title3 = Font.system(size: 18, weight: .semibold)

    // Body
    static let body = Font.system(size: 17, weight: .regular)
    static let bodyMedium = Font.system(size: 17, weight: .medium)
    static let bodyBold = Font.system(size: 17, weight: .bold)

    // Subhead
    static let subheadline = Font.system(size: 15, weight: .regular)
    static let subheadlineMedium = Font.system(size: 15, weight: .medium)

    // Caption
    static let caption = Font.system(size: 12, weight: .regular)
    static let captionMedium = Font.system(size: 12, weight: .medium)

    // Currency (monospaced for alignment)
    static let currency = Font.system(size: 17, weight: .semibold).monospaced()
    static let currencyLarge = Font.system(size: 24, weight: .bold).monospaced()
}
```

---

## Task 12: Reusable UI Components

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Views/Components/GradientButton.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Components/StatCard.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Components/TransactionCell.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Components/ChartCard.swift`

### Step 1: Create GradientButton

Create `ios/FinanceTracker/FinanceTracker/Views/Components/GradientButton.swift`:

```swift
import SwiftUI

struct GradientButton: View {
    let title: String
    let action: () -> Void
    var isDisabled: Bool = false

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(Typography.bodyMedium)
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .padding()
                .background(
                    ColorPalette.primaryGradient
                )
                .cornerRadius(12)
        }
        .disabled(isDisabled)
        .opacity(isDisabled ? 0.5 : 1.0)
    }
}

struct GradientButton_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 16) {
            GradientButton(title: "Get Started", action: {})
            GradientButton(title: "Disabled", action: {}, isDisabled: true)
        }
        .padding()
    }
}
```

### Step 2: Create StatCard

Create `ios/FinanceTracker/FinanceTracker/Views/Components/StatCard.swift`:

```swift
import SwiftUI

struct StatCard: View {
    let icon: String
    let title: String
    let value: String
    let trend: String?
    let isPositive: Bool

    init(
        icon: String,
        title: String,
        value: String,
        trend: String? = nil,
        isPositive: Bool = true
    ) {
        self.icon = icon
        self.title = title
        self.value = value
        self.trend = trend
        self.isPositive = isPositive
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: icon)
                    .font(.title3)
                    .foregroundColor(.white)
                    .frame(width: 40, height: 40)
                    .background(ColorPalette.primaryGradient)
                    .cornerRadius(10)

                Spacer()

                if let trend = trend {
                    HStack(spacing: 4) {
                        Image(systemName: isPositive ? "arrow.up.right" : "arrow.down.right")
                        Text(trend)
                    }
                    .font(Typography.caption)
                    .foregroundColor(isPositive ? ColorPalette.income : ColorPalette.expense)
                }
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(Typography.caption)
                    .foregroundColor(ColorPalette.textSecondary)

                Text(value)
                    .font(Typography.title2)
                    .foregroundColor(ColorPalette.textPrimary)
            }
        }
        .padding(16)
        .background(ColorPalette.cardBackground)
        .cornerRadius(16)
        .shadow(color: .black.opacity(0.05), radius: 10, x: 0, y: 4)
    }
}

struct StatCard_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 16) {
            StatCard(
                icon: "dollarsign.circle.fill",
                title: "Balance",
                value: "$2,450.00",
                trend: "+12%",
                isPositive: true
            )
            StatCard(
                icon: "arrow.down.circle.fill",
                title: "Income",
                value: "$5,000.00",
                trend: "+8%",
                isPositive: true
            )
            StatCard(
                icon: "arrow.up.circle.fill",
                title: "Expenses",
                value: "$2,550.00",
                trend: "-3%",
                isPositive: true
            )
        }
        .padding()
        .background(ColorPalette.background)
    }
}
```

### Step 3: Create TransactionCell

Create `ios/FinanceTracker/FinanceTracker/Views/Components/TransactionCell.swift`:

```swift
import SwiftUI

struct TransactionCell: View {
    let transaction: Transaction

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(backgroundColor)
                    .frame(width: 44, height: 44)

                Image(systemName: transaction.type == .income ? "arrow.down" : "arrow.up")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(transaction.type == .income ? ColorPalette.income : ColorPalette.expense)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(merchantName)
                    .font(Typography.bodyMedium)
                    .foregroundColor(ColorPalette.textPrimary)

                Text(categoryAndDate)
                    .font(Typography.caption)
                    .foregroundColor(ColorPalette.textSecondary)
            }

            Spacer()

            Text(amountText)
                .font(Typography.bodyBold.monospaced())
                .foregroundColor(transaction.type == .income ? ColorPalette.income : ColorPalette.expense)
        }
        .padding(.vertical, 8)
    }

    private var merchantName: String {
        transaction.merchant ?? "Unknown"
    }

    private var categoryAndDate: String {
        var parts: [String] = []
        if let category = transaction.category {
            parts.append(category)
        }
        parts.append(formatDate(transaction.date))
        return parts.joined(separator: " • ")
    }

    private var amountText: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        let prefix = transaction.type == .income ? "+" : "-"
        return prefix + (formatter.string(from: NSNumber(value: transaction.amount)) ?? "")
    }

    private var backgroundColor: Color {
        transaction.type == .income
            ? ColorPalette.success.opacity(0.15)
            : ColorPalette.danger.opacity(0.15)
    }

    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        return formatter.string(from: date)
    }
}

struct TransactionCell_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 0) {
            TransactionCell(transaction: Transaction(
                amount: 1500,
                type: .income,
                merchant: "Salary",
                category: "Income",
                source: "BCA",
                date: Date()
            ))
            TransactionCell(transaction: Transaction(
                amount: 50,
                type: .expense,
                merchant: "Coffee Shop",
                category: "Food",
                source: "Gopay",
                date: Date()
            ))
        }
        .padding()
        .background(ColorPalette.background)
    }
}
```

### Step 4: Create ChartCard

Create `ios/FinanceTracker/FinanceTracker/Views/Components/ChartCard.swift`:

```swift
import SwiftUI
import Charts

struct ChartCard: View {
    let title: String
    let content: AnyView

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text(title)
                .font(Typography.title3)
                .foregroundColor(ColorPalette.textPrimary)

            content
        }
        .padding(16)
        .background(ColorPalette.cardBackground)
        .cornerRadius(16)
        .shadow(color: .black.opacity(0.05), radius: 10, x: 0, y: 4)
    }
}

struct ChartCard_Previews: PreviewProvider {
    static var previews: some View {
        ChartCard(title: "Income vs Expenses") {
            AnyView(
                Text("Chart content here")
                    .frame(height: 200)
            )
        }
        .padding()
        .background(ColorPalette.background)
    }
}
```

### Step 5: Build to verify components

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 13: Dashboard View

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Views/Dashboard/DashboardView.swift`

### Step 1: Create DashboardView

Create `ios/FinanceTracker/FinanceTracker/Views/Dashboard/DashboardView.swift`:

```swift
import SwiftUI

struct DashboardView: View {
    @StateObject private var viewModel: DashboardViewModel
    @State private var showingAddTransaction = false

    init(viewModel: DashboardViewModel) {
        _viewModel = StateObject(wrappedValue: viewModel)
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 20) {
                headerSection
                statsSection
                recentTransactionsSection
            }
            .padding()
        }
        .background(ColorPalette.background)
        .refreshable {
            await viewModel.refresh()
        }
        .overlay {
            if viewModel.isLoading && viewModel.recentTransactions.isEmpty {
                ProgressView()
            }
        }
        .sheet(isPresented: $showingAddTransaction) {
            Text("Add Transaction View")
        }
    }

    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text("Good \(timeOfDay)")
                    .font(Typography.subheadline)
                    .foregroundColor(ColorPalette.textSecondary)

                Text("Dashboard")
                    .font(Typography.largeTitle)
                    .foregroundColor(ColorPalette.textPrimary)
            }

            Spacer()

            Button {
                // TODO: Open notifications
            } label: {
                Image(systemName: "bell.badge")
                    .font(.title3)
                    .foregroundColor(ColorPalette.textPrimary)
            }
        }
    }

    private var statsSection: some View {
        VStack(spacing: 12) {
            StatCard(
                icon: "dollarsign.circle.fill",
                title: "Balance",
                value: formatCurrency(viewModel.balance),
                trend: nil,
                isPositive: true
            )

            HStack(spacing: 12) {
                StatCard(
                    icon: "arrow.down.circle.fill",
                    title: "Income",
                    value: formatCurrency(viewModel.totalIncome),
                    trend: nil,
                    isPositive: true
                )

                StatCard(
                    icon: "arrow.up.circle.fill",
                    title: "Expenses",
                    value: formatCurrency(viewModel.totalExpenses),
                    trend: nil,
                    isPositive: true
                )
            }
        }
    }

    private var recentTransactionsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Recent Transactions")
                    .font(Typography.title3)
                    .foregroundColor(ColorPalette.textPrimary)

                Spacer()

                Button("See All") {
                    // TODO: Navigate to transactions list
                }
                .font(Typography.subheadlineMedium)
                .foregroundColor(ColorPalette.primaryIndigo)
            }

            if viewModel.recentTransactions.isEmpty {
                Text("No transactions yet")
                    .font(Typography.body)
                    .foregroundColor(ColorPalette.textSecondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.vertical, 20)
            } else {
                VStack(spacing: 0) {
                    ForEach(viewModel.recentTransactions) { transaction in
                        TransactionCell(transaction: transaction)

                        if transaction.id != viewModel.recentTransactions.last?.id {
                            Divider()
                                .padding(.leading, 68)
                        }
                    }
                }
                .background(ColorPalette.cardBackground)
                .cornerRadius(16)
                .shadow(color: .black.opacity(0.05), radius: 10, x: 0, y: 4)
            }
        }
    }

    private var timeOfDay: String {
        let hour = Calendar.current.component(.hour, from: Date())
        switch hour {
        case 0..<12: return "Morning"
        case 12..<18: return "Afternoon"
        default: return "Evening"
        }
    }

    private func formatCurrency(_ value: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        return formatter.string(from: NSNumber(value: value)) ?? "$0.00"
    }
}

struct DashboardView_Previews: PreviewProvider {
    static var previews: some View {
        let mockAPI = MockAPIServiceForDashboard()
        mockAPI.mockTransactions = [
            Transaction(amount: 1500, type: .income, merchant: "Salary", source: "BCA"),
            Transaction(amount: 50, type: .expense, merchant: "Coffee", source: "Gopay")
        ]
        mockAPI.mockSummary = SummaryResponse(balance: 2000, totalIncome: 5000, totalExpenses: 3000)

        let viewModel = DashboardViewModel(apiService: mockAPI)

        return DashboardView(viewModel: viewModel)
    }
}
```

### Step 2: Build to verify view compiles

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 14: Analytics View

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/ViewModels/AnalyticsViewModel.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Analytics/AnalyticsView.swift`

### Step 1: Create AnalyticsViewModel

Create `ios/FinanceTracker/FinanceTracker/ViewModels/AnalyticsViewModel.swift`:

```swift
import Foundation
import Combine
import Charts

@MainActor
class AnalyticsViewModel: ObservableObject {
    @Published var analytics: Analytics?
    @Published var selectedPeriod: TimePeriod = .month
    @Published var isLoading = false
    @Published var errorMessage: String?

    private let apiService: APIServiceProtocol

    init(apiService: APIServiceProtocol) {
        self.apiService = apiService
    }

    func loadAnalytics() async {
        isLoading = true
        errorMessage = nil

        do {
            let data = try await apiService.getAnalytics(period: selectedPeriod)
            self.analytics = data
            isLoading = false
        } catch {
            isLoading = false
            errorMessage = error.localizedDescription
        }
    }

    func changePeriod(_ period: TimePeriod) async {
        selectedPeriod = period
        await loadAnalytics()
    }
}
```

### Step 2: Create AnalyticsView

Create `ios/FinanceTracker/FinanceTracker/Views/Analytics/AnalyticsView.swift`:

```swift
import SwiftUI
import Charts

struct AnalyticsView: View {
    @StateObject private var viewModel: AnalyticsViewModel

    init(viewModel: AnalyticsViewModel) {
        _viewModel = StateObject(wrappedValue: viewModel)
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 20) {
                periodSelector

                if let analytics = viewModel.analytics {
                    summarySection(analytics: analytics)

                    if !analytics.categoryBreakdown.isEmpty {
                        categoryChartSection(analytics: analytics)
                    }

                    if !analytics.sourceBreakdown.isEmpty {
                        sourceRankingSection(analytics: analytics)
                    }
                }
            }
            .padding()
        }
        .background(ColorPalette.background)
        .task {
            await viewModel.loadAnalytics()
        }
    }

    private var periodSelector: some View {
        Picker("Period", selection: $viewModel.selectedPeriod) {
            ForEach(TimePeriod.allCases, id: \.self) { period in
                Text(period.displayName).tag(period)
            }
        }
        .pickerStyle(.segmented)
        .onChange(of: viewModel.selectedPeriod) { newPeriod in
            Task {
                await viewModel.changePeriod(newPeriod)
            }
        }
    }

    private func summarySection(analytics: Analytics) -> some View {
        HStack(spacing: 12) {
            StatCard(
                icon: "arrow.down.circle.fill",
                title: "Income",
                value: formatCurrency(analytics.totalIncome),
                trend: nil,
                isPositive: true
            )

            StatCard(
                icon: "arrow.up.circle.fill",
                title: "Expenses",
                value: formatCurrency(analytics.totalExpenses),
                trend: nil,
                isPositive: true
            )
        }
    }

    private func categoryChartSection(analytics: Analytics) -> some View {
        ChartCard(title: "Spending by Category") {
            AnyView(
                VStack(alignment: .leading, spacing: 16) {
                    Chart(analytics.categoryBreakdown) { item in
                        SectorMark(
                            angle: .value("Amount", item.amount),
                            innerRadius: .ratio(0.5),
                            angularInset: 2
                        )
                        .foregroundStyle(by: .value("Category", item.category))
                        .cornerRadius(4)
                    }
                    .frame(height: 200)
                    .chartLegend(position: .bottom, alignment: .leading)

                    ForEach(analytics.categoryBreakdown.prefix(5)) { item in
                        HStack {
                            Circle()
                                .fill(colorForCategory(item.category))
                                .frame(width: 12, height: 12)

                            Text(item.category)
                                .font(Typography.subheadline)
                                .foregroundColor(ColorPalette.textPrimary)

                            Spacer()

                            VStack(alignment: .trailing, spacing: 2) {
                                Text(formatCurrency(item.amount))
                                    .font(Typography.bodyMedium)
                                    .foregroundColor(ColorPalette.textPrimary)

                                Text("\(Int(item.percentage))%")
                                    .font(Typography.caption)
                                    .foregroundColor(ColorPalette.textSecondary)
                            }
                        }
                    }
                }
            )
        }
    }

    private func sourceRankingSection(analytics: Analytics) -> some View {
        ChartCard(title: "Sources") {
            AnyView(
                VStack(spacing: 12) {
                    ForEach(analytics.sourceBreakdown.sorted(by: { $0.amount > $1.amount })) { item in
                        HStack {
                            Text(item.source)
                                .font(Typography.body)
                                .foregroundColor(ColorPalette.textPrimary)

                            Spacer()

                            Text(formatCurrency(item.amount))
                                .font(Typography.bodyMedium.monospaced())
                                .foregroundColor(ColorPalette.textPrimary)

                            Text("(\(item.transactionCount))")
                                .font(Typography.caption)
                                .foregroundColor(ColorPalette.textSecondary)
                    }
                }
            )
        }
    }

    private func formatCurrency(_ value: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        return formatter.string(from: NSNumber(value: value)) ?? "$0.00"
    }

    private func colorForCategory(_ category: String) -> Color {
        let colors: [Color] = [
            .blue, .purple, .pink, .orange, .green,
            .yellow, .cyan, .indigo, .mint, .teal
        ]
        let index = abs(category.hashValue) % colors.count
        return colors[index]
    }
}

struct AnalyticsView_Previews: PreviewProvider {
    static var previews: some View {
        let mockAPI = MockAPIServiceForDashboard()
        let viewModel = AnalyticsViewModel(apiService: mockAPI)

        return AnalyticsView(viewModel: viewModel)
    }
}
```

### Step 3: Build to verify views compile

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 15: Authentication Views

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/ViewModels/AuthViewModel.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Auth/LoginView.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Auth/RegisterView.swift`

### Step 1: Create AuthViewModel

Create `ios/FinanceTracker/FinanceTracker/ViewModels/AuthViewModel.swift`:

```swift
import Foundation
import Combine

@MainActor
class AuthViewModel: ObservableObject {
    @Published var email = ""
    @Published var password = ""
    @Published var confirmPassword = ""
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var isAuthenticated = false

    private let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    func login() async {
        isLoading = true
        errorMessage = nil

        do {
            try await authManager.login(email: email, password: password)
            isAuthenticated = true
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    func register() async {
        isLoading = true
        errorMessage = nil

        guard password == confirmPassword else {
            errorMessage = "Passwords do not match"
            isLoading = false
            return
        }

        do {
            try await authManager.register(email: email, password: password)
            isAuthenticated = true
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    var isValid: Bool {
        !email.isEmpty && !password.isEmpty && email.contains("@")
    }

    var isValidForRegistration: Bool {
        isValid && !confirmPassword.isEmpty && password == confirmPassword
    }
}
```

### Step 2: Create LoginView

Create `ios/FinanceTracker/FinanceTracker/Views/Auth/LoginView.swift`:

```swift
import SwiftUI

struct LoginView: View {
    @StateObject private var viewModel: AuthViewModel
    @Environment(\.dismiss) var dismiss
    @State private var showingRegister = false

    init(authManager: AuthManager) {
        _viewModel = StateObject(wrappedValue: AuthViewModel(authManager: authManager))
    }

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "chart.line.uptrend.xyaxis.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(ColorPalette.primaryGradient)

            VStack(spacing: 8) {
                Text("Finance Tracker")
                    .font(Typography.largeTitle)
                    .foregroundColor(ColorPalette.textPrimary)

                Text("Track your finances with ease")
                    .font(Typography.body)
                    .foregroundColor(ColorPalette.textSecondary)
            }

            Spacer()

            VStack(spacing: 16) {
                TextField("Email", text: $viewModel.email)
                    .textFieldStyle(.roundedBorder)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)

                SecureField("Password", text: $viewModel.password)
                    .textFieldStyle(.roundedBorder)

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(Typography.caption)
                        .foregroundColor(ColorPalette.danger)
                }

                GradientButton(
                    title: "Sign In",
                    action: {
                        Task {
                            await viewModel.login()
                            if viewModel.isAuthenticated {
                                dismiss()
                            }
                        }
                    },
                    isDisabled: !viewModel.isValid || viewModel.isLoading
                )
            }
            .padding()

            HStack(spacing: 4) {
                Text("Don't have an account?")
                    .font(Typography.subheadline)
                    .foregroundColor(ColorPalette.textSecondary)

                Button("Sign Up") {
                    showingRegister = true
                }
                .font(Typography.subheadlineMedium)
                .foregroundColor(ColorPalette.primaryIndigo)
            }

            Spacer()
        }
        .padding()
        .sheet(isPresented: $showingRegister) {
            RegisterView(authManager: viewModel.authManager)
        }
    }
}

struct LoginView_Previews: PreviewProvider {
    static var previews: some View {
        let mockAPI = MockAPIServiceForAuth()
        let authManager = AuthManager(apiService: mockAPI)
        return LoginView(authManager: authManager)
    }
}
```

### Step 3: Create RegisterView

Create `ios/FinanceTracker/FinanceTracker/Views/Auth/RegisterView.swift`:

```swift
import SwiftUI

struct RegisterView: View {
    @StateObject private var viewModel: AuthViewModel
    @Environment(\.dismiss) var dismiss

    init(authManager: AuthManager) {
        _viewModel = StateObject(wrappedValue: AuthViewModel(authManager: authManager))
    }

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "chart.line.uptrend.xyaxis.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(ColorPalette.primaryGradient)

            VStack(spacing: 8) {
                Text("Create Account")
                    .font(Typography.largeTitle)
                    .foregroundColor(ColorPalette.textPrimary)

                Text("Start tracking your finances")
                    .font(Typography.body)
                    .foregroundColor(ColorPalette.textSecondary)
            }

            Spacer()

            VStack(spacing: 16) {
                TextField("Email", text: $viewModel.email)
                    .textFieldStyle(.roundedBorder)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)

                SecureField("Password", text: $viewModel.password)
                    .textFieldStyle(.roundedBorder)

                SecureField("Confirm Password", text: $viewModel.confirmPassword)
                    .textFieldStyle(.roundedBorder)

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(Typography.caption)
                        .foregroundColor(ColorPalette.danger)
                }

                GradientButton(
                    title: "Sign Up",
                    action: {
                        Task {
                            await viewModel.register()
                            if viewModel.isAuthenticated {
                                dismiss()
                            }
                        }
                    },
                    isDisabled: !viewModel.isValidForRegistration || viewModel.isLoading
                )
            }
            .padding()

            Spacer()
        }
        .padding()
        .overlay(alignment: .topLeading) {
            Button {
                dismiss()
            } label: {
                Image(systemName: "xmark.circle.fill")
                    .font(.title3)
                    .foregroundColor(.gray)
            }
            .padding()
        }
    }
}

struct RegisterView_Previews: PreviewProvider {
    static var previews: some View {
        let mockAPI = MockAPIServiceForAuth()
        let authManager = AuthManager(apiService: mockAPI)
        return RegisterView(authManager: authManager)
    }
}
```

### Step 4: Build to verify auth views compile

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 16: Main Tab Navigation and App Entry Point

**Files:**
- Modify: `ios/FinanceTracker/FinanceTracker/ContentView.swift`

### Step 1: Create Main TabView

Modify `ios/FinanceTracker/FinanceTracker/ContentView.swift`:

```swift
import SwiftUI

struct ContentView: View {
    @StateObject private var authManager: AuthManager
    @State private var selectedTab = 0
    @State private var showingLogin = false

    init() {
        let mockAPI = APIService(baseURL: Config.baseURL)
        _authManager = StateObject(wrappedValue: AuthManager(apiService: mockAPI))
    }

    var body: some View {
        Group {
            if authManager.isAuthenticated {
                mainTabView
            } else {
                loginView
            }
        }
        .onAppear {
            showingLogin = !authManager.isAuthenticated
        }
        .sheet(isPresented: $showingLogin) {
            LoginView(authManager: authManager)
        }
    }

    private var mainTabView: some View {
        TabView(selection: $selectedTab) {
            DashboardView(viewModel: DashboardViewModel(apiService: APIService(baseURL: Config.baseURL)))
                .tabItem {
                    Label("Dashboard", systemImage: selectedTab == 0 ? "chart.bar.fill" : "chart.bar")
                }
                .tag(0)

            AnalyticsView(viewModel: AnalyticsViewModel(apiService: APIService(baseURL: Config.baseURL)))
                .tabItem {
                    Label("Analytics", systemImage: selectedTab == 1 ? "chart.pie.fill" : "chart.pie")
                }
                .tag(1)

            Text("Transactions View")
                .tabItem {
                    Label("Transactions", systemImage: selectedTab == 2 ? "list.bullet.rectangle.fill" : "list.bullet.rectangle")
                }
                .tag(2)

            Text("Settings View")
                .tabItem {
                    Label("Settings", systemImage: selectedTab == 3 ? "gearshape.fill" : "gearshape")
                }
                .tag(3)
        }
        .accentColor(ColorPalette.primaryIndigo)
        .environmentObject(authManager)
    }

    private var loginView: some View {
        VStack {
            if showingLogin {
                LoginView(authManager: authManager)
            } else {
                ProgressView()
            }
        }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
```

### Step 2: Build to verify app compiles

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 17: Transaction List View

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/ViewModels/TransactionsViewModel.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Transactions/TransactionsView.swift`
- Create: `ios/FinanceTracker/FinanceTracker/Views/Transactions/TransactionDetailView.swift`

### Step 1: Create TransactionsViewModel

Create `ios/FinanceTracker/FinanceTracker/ViewModels/TransactionsViewModel.swift`:

```swift
import Foundation
import Combine

@MainActor
class TransactionsViewModel: ObservableObject {
    @Published var transactions: [Transaction] = []
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var selectedType: TransactionType?
    @Published var selectedSource: String?
    @Published var searchText: String = ""

    private let apiService: APIServiceProtocol
    private var currentPage = 1
    private let pageSize = 20
    private var hasMorePages = true

    init(apiService: APIServiceProtocol) {
        self.apiService = apiService
    }

    func loadTransactions() async {
        isLoading = true
        errorMessage = nil

        do {
            let newTransactions = try await apiService.getTransactions(page: currentPage, limit: pageSize)

            if currentPage == 1 {
                transactions = newTransactions
            } else {
                transactions.append(contentsOf: newTransactions)
            }

            hasMorePages = newTransactions.count == pageSize
            isLoading = false
        } catch {
            isLoading = false
            errorMessage = error.localizedDescription
        }
    }

    func loadMore() async {
        guard !isLoading && hasMorePages else { return }
        currentPage += 1
        await loadTransactions()
    }

    func refresh() async {
        currentPage = 1
        await loadTransactions()
    }

    var filteredTransactions: [Transaction] {
        transactions.filter { transaction in
            if let type = selectedType, transaction.type != type {
                return false
            }
            if let source = selectedSource, transaction.source != source {
                return false
            }
            if !searchText.isEmpty {
                let search = searchText.lowercased()
                return transaction.merchant?.lowercased().contains(search) == true
                    || transaction.category?.lowercased().contains(search) == true
            }
            return true
        }
    }

    var uniqueSources: [String] {
        Set(transactions.compactMap { $0.source })
            .sorted()
    }
}
```

### Step 2: Create TransactionsView

Create `ios/FinanceTracker/FinanceTracker/Views/Transactions/TransactionsView.swift`:

```swift
import SwiftUI

struct TransactionsView: View {
    @StateObject private var viewModel: TransactionsViewModel
    @State private var selectedTransaction: Transaction?

    init(apiService: APIServiceProtocol) {
        _viewModel = StateObject(wrappedValue: TransactionsViewModel(apiService: apiService))
    }

    var body: some View {
        VStack(spacing: 0) {
            filterSection

            if viewModel.filteredTransactions.isEmpty && !viewModel.isLoading {
                emptyState
            } else {
                List {
                    ForEach(viewModel.filteredTransactions) { transaction in
                        Button {
                            selectedTransaction = transaction
                        } label: {
                            TransactionCell(transaction: transaction)
                                .listRowInsets(EdgeInsets(top: 4, leading: 16, bottom: 4, trailing: 16))
                                .listRowSeparator(.hidden)
                                .listRowBackground(Color.clear)
                        }

                        if transaction.id == viewModel.filteredTransactions.last?.id && viewModel.hasMorePages {
                            Color.clear
                                .onAppear {
                                    Task {
                                        await viewModel.loadMore()
                                    }
                                }
                        }
                    }
                }
                .listStyle(.plain)
                .background(ColorPalette.background)
            }
        }
        .background(ColorPalette.background)
        .refreshable {
            await viewModel.refresh()
        }
        .sheet(item: $selectedTransaction) { transaction in
            TransactionDetailView(transaction: transaction)
        }
        .searchable(text: $viewModel.searchText, prompt: "Search transactions...")
        .task {
            if viewModel.transactions.isEmpty {
                await viewModel.loadTransactions()
            }
        }
    }

    private var filterSection: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                FilterChip(
                    title: "All",
                    isSelected: viewModel.selectedType == nil
                ) {
                    viewModel.selectedType = nil
                }

                ForEach(TransactionType.allCases, id: \.self) { type in
                    FilterChip(
                        title: type.displayName,
                        isSelected: viewModel.selectedType == type
                    ) {
                        viewModel.selectedType = type
                    }
                }

                Spacer()
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
        }
        .background(ColorPalette.cardBackground)
    }

    private var emptyState: some View {
        VStack(spacing: 16) {
            Image(systemName: "tray")
                .font(.system(size: 60))
                .foregroundColor(ColorPalette.textSecondary)

            Text("No transactions found")
                .font(Typography.body)
                .foregroundColor(ColorPalette.textSecondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct FilterChip: View {
    let title: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(Typography.subheadlineMedium)
                .foregroundColor(isSelected ? .white : ColorPalette.textPrimary)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(isSelected ? ColorPalette.primaryGradient : Color.gray.opacity(0.1))
                .cornerRadius(20)
        }
    }
}

struct TransactionsView_Previews: PreviewProvider {
    static var previews: some View {
        let mockAPI = MockAPIServiceForDashboard()
        return TransactionsView(apiService: mockAPI)
    }
}
```

### Step 3: Create TransactionDetailView

Create `ios/FinanceTracker/FinanceTracker/Views/Transactions/TransactionDetailView.swift`:

```swift
import SwiftUI

struct TransactionDetailView: View {
    let transaction: Transaction
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationView {
            VStack(spacing: 24) {
                Spacer()

                VStack(spacing: 8) {
                    Image(systemName: transaction.type == .income ? "arrow.down.circle.fill" : "arrow.up.circle.fill")
                        .font(.system(size: 60))
                        .foregroundColor(transaction.type == .income ? ColorPalette.income : ColorPalette.expense)

                    Text(amountText)
                        .font(Typography.largeTitle.monospaced())
                        .foregroundColor(transaction.type == .income ? ColorPalette.income : ColorPalette.expense)

                    Text(transaction.type.displayName)
                        .font(Typography.subheadline)
                        .foregroundColor(ColorPalette.textSecondary)
                }

                Divider()

                VStack(spacing: 16) {
                    DetailRow(label: "Merchant", value: transaction.merchant ?? "Unknown")
                    DetailRow(label: "Category", value: transaction.category ?? "Uncategorized")
                    DetailRow(label: "Source", value: transaction.source ?? "Unknown")
                    DetailRow(label: "Date", value: formatDate(transaction.date))
                    DetailRow(label: "Transaction ID", value: transaction.id.uuidString)
                }
                .padding()

                Spacer()
            }
            .padding()
            .navigationTitle("Transaction Details")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }

    private var amountText: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        let prefix = transaction.type == .income ? "+" : "-"
        return prefix + (formatter.string(from: NSNumber(value: transaction.amount)) ?? "")
    }

    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .long
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

struct DetailRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .font(Typography.subheadline)
                .foregroundColor(ColorPalette.textSecondary)

            Spacer()

            Text(value)
                .font(Typography.body)
                .foregroundColor(ColorPalette.textPrimary)
        }
    }
}

struct TransactionDetailView_Previews: PreviewProvider {
    static var previews: some View {
        TransactionDetailView(
            transaction: Transaction(
                amount: 1500,
                type: .income,
                merchant: "Salary",
                category: "Income",
                source: "BCA",
                date: Date()
            )
        )
    }
}
```

### Step 4: Update ContentView to use TransactionsView

Modify `ios/FinanceTracker/FinanceTracker/ContentView.swift` - Update the Transactions tab:

```swift
// In mainTabView, replace the Transactions tab with:
TransactionsView(viewModel: TransactionsViewModel(apiService: APIService(baseURL: Config.baseURL)))
    .tabItem {
        Label("Transactions", systemImage: selectedTab == 2 ? "list.bullet.rectangle.fill" : "list.bullet.rectangle")
    }
    .tag(2)
```

### Step 5: Build to verify views compile

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 18: Settings View

**Files:**
- Create: `ios/FinanceTracker/FinanceTracker/Views/Settings/SettingsView.swift`

### Step 1: Create SettingsView

Create `ios/FinanceTracker/FinanceTracker/Views/Settings/SettingsView.swift`:

```swift
import SwiftUI
import LocalAuthentication

struct SettingsView: View {
    @EnvironmentObject var authManager: AuthManager
    @State private var isBiometricEnabled = false
    @State private var showingBiometricError = false
    @State private var biometricError: Error?

    var body: some View {
        NavigationView {
            List {
                Section {
                    HStack(spacing: 16) {
                        Circle()
                            .fill(ColorPalette.primaryGradient)
                            .frame(width: 60, height: 60)

                        VStack(alignment: .leading, spacing: 4) {
                            Text(authManager.currentUser?.name ?? "User")
                                .font(Typography.bodyBold)
                                .foregroundColor(ColorPalette.textPrimary)

                            Text(authManager.currentUser?.email ?? "")
                                .font(Typography.subheadline)
                                .foregroundColor(ColorPalette.textSecondary)
                        }

                        Spacer()
                    }
                    .padding(.vertical, 8)
                }

                Section("Security") {
                    Toggle("Biometric Authentication", isOn: $isBiometricEnabled)
                        .onChange(of: isBiometricEnabled) { newValue in
                            enableBiometric(newValue)
                        }
                }

                Section("Notifications") {
                    HStack {
                        Text("Notification Access")
                            .font(Typography.body)
                            .foregroundColor(ColorPalette.textPrimary)

                        Spacer()

                        Image(systemName: "checkmark.circle.fill")
                            .foregroundColor(ColorPalette.success)
                    }

                    HStack {
                        Text("Connected Sources")
                            .font(Typography.body)
                            .foregroundColor(ColorPalette.textPrimary)

                        Spacer()

                        Text("3 sources")
                            .font(Typography.subheadline)
                            .foregroundColor(ColorPalette.textSecondary)
                    }
                }

                Section("About") {
                    HStack {
                        Text("Version")
                            .font(Typography.body)
                            .foregroundColor(ColorPalette.textPrimary)

                        Spacer()

                        Text("1.0.0")
                            .font(Typography.subheadline)
                            .foregroundColor(ColorPalette.textSecondary)
                    }
                }

                Section {
                    Button {
                        authManager.logout()
                    } label: {
                        HStack {
                            Spacer()
                            Text("Log Out")
                                .font(Typography.bodyMedium)
                                .foregroundColor(ColorPalette.danger)
                            Spacer()
                        }
                    }
                }
            }
            .navigationTitle("Settings")
            .alert("Biometric Error", isPresented: $showingBiometricError) {
                Button("OK", role: .cancel) {}
            } message: {
                if let error = biometricError {
                    Text(error.localizedDescription)
                }
            }
        }
    }

    private func enableBiometric(_ enable: Bool) {
        guard enable else {
            authManager.currentUser?.isBiometricEnabled = false
            return
        }

        let context = LAContext()
        var error: NSError?

        if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
            context.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, localizedReason: "Enable biometric authentication for Finance Tracker") { success, error in
                DispatchQueue.main.async {
                    if success {
                        authManager.currentUser?.isBiometricEnabled = true
                    } else {
                        isBiometricEnabled = false
                        if let error = error {
                            biometricError = error
                            showingBiometricError = true
                        }
                    }
                }
            }
        } else {
            isBiometricEnabled = false
            biometricError = error
            showingBiometricError = true
        }
    }
}

struct SettingsView_Previews: PreviewProvider {
    static var previews: some View {
        let mockAPI = MockAPIServiceForAuth()
        let authManager = AuthManager(apiService: mockAPI)
        return SettingsView()
            .environmentObject(authManager)
    }
}
```

### Step 2: Update ContentView to use SettingsView

Modify `ios/FinanceTracker/FinanceTracker/ContentView.swift` - Update the Settings tab:

```swift
// In mainTabView, replace the Settings tab with:
SettingsView()
    .tabItem {
        Label("Settings", systemImage: selectedTab == 3 ? "gearshape.fill" : "gearshape")
    }
    .tag(3)
```

### Step 3: Build to verify settings view compiles

Run:
```bash
cd ios/FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

---

## Task 19: Integration Tests

**Files:**
- Create: `ios/FinanceTracker/FinanceTrackerTests/Integration/AuthenticationFlowTests.swift`
- Create: `ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift`

### Step 1: Create AuthenticationFlowTests

Create `ios/FinanceTracker/FinanceTrackerTests/Integration/AuthenticationFlowTests.swift`:

```swift
import XCTest
@testable import FinanceTracker

class AuthenticationFlowTests: XCTestCase {
    var authManager: AuthManager!
    var mockAPI: MockAPIServiceForAuth!
    var keychainManager: KeychainManager!

    override func setUp() {
        super.setUp()
        mockAPI = MockAPIServiceForAuth()
        keychainManager = KeychainManager()
        keychainManager.delete(key: "finance_tracker_api_key")
        authManager = AuthManager(apiService: mockAPI, keychainManager: keychainManager)
    }

    override func tearDown() {
        keychainManager.delete(key: "finance_tracker_api_key")
        super.tearDown()
    }

    func testCompleteLoginFlow() async throws {
        let email = "test@example.com"
        let password = "password123"
        mockAPI.mockUser = User(email: email, apiKey: "test_api_key")

        try await authManager.login(email: email, password: password)

        XCTAssertTrue(authManager.isAuthenticated)

        let storedKey = keychainManager.retrieve(key: "finance_tracker_api_key")
        XCTAssertEqual(storedKey, "test_api_key")

        XCTAssertEqual(authManager.currentUser?.email, email)
    }

    func testCompleteLogoutFlow() async throws {
        mockAPI.mockUser = User(email: "test@example.com", apiKey: "test_api_key")
        try await authManager.login(email: "test@example.com", password: "password123")
        XCTAssertTrue(authManager.isAuthenticated)

        authManager.logout()

        XCTAssertFalse(authManager.isAuthenticated)

        let storedKey = keychainManager.retrieve(key: "finance_tracker_api_key")
        XCTAssertNil(storedKey)
    }

    func testRegistrationFlow() async throws {
        let email = "newuser@example.com"
        let password = "password123"
        mockAPI.mockUser = User(email: email, apiKey: "new_api_key")

        try await authManager.register(email: email, password: password)

        XCTAssertTrue(authManager.isAuthenticated)

        let storedKey = keychainManager.retrieve(key: "finance_tracker_api_key")
        XCTAssertEqual(storedKey, "new_api_key")
    }
}
```

### Step 2: Create NotificationFlowTests

Create `ios/FinanceTracker/FinanceTrackerTests/Integration/NotificationFlowTests.swift`:

```swift
import XCTest
import UserNotifications
@testable import FinanceTracker

class NotificationFlowTests: XCTestCase {
    var parser: TransactionParser!
    var mockAPI: MockAPIServiceForNotification!
    var notificationManager: NotificationManager!

    override func setUp() {
        super.setUp()
        parser = TransactionParser()
        mockAPI = MockAPIServiceForNotification()
        notificationManager = NotificationManager(
            transactionParser: parser,
            apiService: mockAPI,
            coreDataStack: CoreDataStack(inMemory: true)
        )
    }

    func testNotificationToTransactionFlow() async throws {
        let notificationText = "You spent Rp 50.000 at Coffee Shop"
        let source = "BCA"

        guard let transaction = parser.parse(notificationText: notificationText, source: source) else {
            XCTFail("Failed to parse transaction")
            return
        }

        XCTAssertEqual(transaction.amount, 50000.0)
        XCTAssertEqual(transaction.type, .expense)
        XCTAssertEqual(transaction.merchant, "Coffee Shop")
        XCTAssertEqual(transaction.source, "BCA")
    }

    func testIncomeNotificationParsing() async throws {
        let notificationText = "You received Rp 1.500.000 from John Doe"
        let source = "Mandiri"

        guard let transaction = parser.parse(notificationText: notificationText, source: source) else {
            XCTFail("Failed to parse transaction")
            return
        }

        XCTAssertEqual(transaction.type, .income)
        XCTAssertEqual(transaction.amount, 1500000.0)
        XCTAssertEqual(transaction.merchant, "John Doe")
    }
}

class MockAPIServiceForNotification: APIServiceProtocol {
    var capturedTransaction: Transaction?

    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        capturedTransaction = transaction
        return transaction
    }
}
```

### Step 3: Run integration tests

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' -only-testing:FinanceTrackerTests/IntegrationTests
```

---

## Task 20: Final Build and Test Verification

**Files:**
- Verify all files compile

### Step 1: Clean build

Run:
```bash
cd ios/FinanceTracker
xcodebuild clean -scheme FinanceTracker
xcodebuild -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15' build
```

### Step 2: Run all tests

Run:
```bash
cd ios/FinanceTracker
xcodebuild test -scheme FinanceTracker -destination 'platform=iOS Simulator,name=iPhone 15'
```

### Step 3: Verify Info.plist has necessary permissions

Check `ios/FinanceTracker/FinanceTracker/Info.plist` contains:

```xml
<key>NSUserNotificationsUsageDescription</key>
<string>We need notification access to automatically record your transactions from bank alerts.</string>
<key>NSFaceIDUsageDescription</key>
<string>Use Face ID to securely access your finance data.</string>
```

If missing, add them.

---

## Summary

This implementation plan creates a complete iOS Finance Tracker application with:

1. **Project Setup**: Xcode project with proper folder structure
2. **Core Data**: Persistent storage for offline transaction caching
3. **Domain Models**: Transaction, User, Analytics, TimePeriod
4. **Services**: KeychainManager, APIService, AuthManager, TransactionParser, NotificationManager
5. **ViewModels**: Dashboard, Analytics, Transactions, Auth
6. **Views**: Dashboard, Analytics, Transactions, Settings, Login, Register
7. **Components**: GradientButton, StatCard, TransactionCell, ChartCard
8. **Design System**: ColorPalette and Typography
9. **Tests**: Unit tests for services and view models, integration tests

**Total Tasks**: 20
**Estimated File Count**: 40+ files
**Test Coverage**: Core services, ViewModels, and integration flows

---

## Notes for Execution

1. **API Configuration**: Update `Config.swift` with actual backend URL
2. **Pattern File**: The `NotificationPatterns.plist` may need adjustment based on actual bank notification formats
3. **Dependency Injection**: The plan uses protocol-based DI for testability
4. **iOS Permissions**: Ensure notification and biometric permissions are properly configured in Info.plist
5. **Swift Charts**: Requires iOS 16.0+ deployment target
