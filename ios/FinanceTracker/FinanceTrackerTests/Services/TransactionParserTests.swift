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