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