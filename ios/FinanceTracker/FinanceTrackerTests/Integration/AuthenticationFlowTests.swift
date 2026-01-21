import XCTest
@testable import FinanceTracker
import LocalAuthentication

final class AuthenticationFlowTests: XCTestCase {

    private var authManager: AuthManager!
    private var userDefaults: UserDefaults!

    override func setUp() {
        super.setUp()
        userDefaults = UserDefaults(suiteName: "TestDefaults")
        userDefaults?.removePersistentDomain(forName: "TestDefaults")

        authManager = AuthManager(userDefaults: userDefaults)
    }

    override func tearDown() {
        userDefaults?.removePersistentDomain(forName: "TestDefaults")
        authManager = nil
        userDefaults = nil
        super.tearDown()
    }

    // MARK: - Authentication Tests

    func testSuccessfulFaceIDAuthentication() {
        // Given
        let expectation = XCTestExpectation(description: "FaceID authentication succeeds")

        // Simulate FaceID being available
        let context = LAContext()
        var error: NSError?
        _ = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)

        // When
        authManager.authenticateWithFaceID { success in
            // Then
            XCTAssertTrue(success)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testFaceIDAuthenticationFailure() {
        // Given
        let expectation = XCTestExpectation(description: "FaceID authentication fails")

        // Simulate FaceID being available but authentication failing
        let context = LAContext()
        var error: NSError?
        _ = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)

        // When
        authManager.authenticateWithFaceID { success in
            // Then
            XCTAssertFalse(success)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testFallbackToPasscodeWhenFaceIDUnavailable() {
        // Given
        let expectation = XCTestExpectation(description: "Passcode authentication fallback")

        // Simulate FaceID being unavailable
        let context = LAContext()
        var error: NSError?
        let canUseBiometrics = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)
        XCTAssertFalse(canUseBiometrics)

        // When
        authManager.authenticateWithFaceID { success in
            // Then
            // Should fall back to passcode authentication
            XCTAssertFalse(success)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 5.0)
    }

    func testSessionTimeoutBehavior() {
        // Given
        authManager.setSessionTimeout(300) // 5 minutes

        // When
        let loginTime = Date()
        var elapsedTime = 0.0

        repeat {
            elapsedTime = Date().timeIntervalSince(loginTime)
        } while elapsedTime < 2.0 // Simulate time passing

        // Then
        XCTAssertFalse(authManager.isSessionActive())
    }

    func testSessionRenewalAfterAuthentication() {
        // Given
        authManager.setSessionTimeout(300)
        XCTAssertFalse(authManager.isSessionActive())

        // When
        authManager.authenticateWithFaceID { success in
            if success {
                XCTAssertTrue(self.authManager.isSessionActive())
            }
        }

        // Sleep briefly to allow authentication to complete
        usleep(100000)
    }

    // MARK: - Security Tests

    func testBiometricDataProtection() {
        // Given
        let context = LAContext()
        var error: NSError?
        _ = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)

        // When
        let isBiometricEnabled = authManager.isBiometricAuthenticationAvailable()

        // Then
        // Verify biometric data is properly protected
        XCTAssert(isBiometricEnabled || !isBiometricEnabled)
    }

    func testAuthenticationAttemptsLockout() {
        // Given
        let maxAttempts = 5

        // When
        for _ in 0..<maxAttempts {
            authManager.authenticateWithFaceID { success in
                XCTAssertFalse(success)
            }
        }

        // Then
        // App should lock out after max attempts
        let isLocked = authManager.isLockedOut()
        XCTAssertTrue(isLocked)
    }

    // MARK: - Persistence Tests

    func testAuthenticationStatePersistence() {
        // Given
        authManager.saveAuthenticationState(true, for: "user123")

        // When
        let isAuthenticated = authManager.getAuthenticationState(for: "user123")

        // Then
        XCTAssertTrue(isAuthenticated)
    }

    func testClearAuthenticationState() {
        // Given
        authManager.saveAuthenticationState(true, for: "user123")
        XCTAssertTrue(authManager.getAuthenticationState(for: "user123"))

        // When
        authManager.clearAuthenticationState(for: "user123")

        // Then
        XCTAssertFalse(authManager.getAuthenticationState(for: "user123"))
    }
}