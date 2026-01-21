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