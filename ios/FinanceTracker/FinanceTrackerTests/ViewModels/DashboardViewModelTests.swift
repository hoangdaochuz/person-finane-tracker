import XCTest
@testable import FinanceTracker

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