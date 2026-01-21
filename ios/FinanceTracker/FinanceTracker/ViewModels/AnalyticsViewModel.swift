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