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