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

                    if !analytics.categorySummaries.isEmpty {
                        categoryChartSection(analytics: analytics)
                    }

                    if !analytics.sourceSummaries.isEmpty {
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
        ChartCard(
            title: "Spending by Category",
            content: AnyView(
                VStack(alignment: .leading, spacing: 16) {
                    if #available(iOS 17.0, *) {
                        Chart(analytics.categorySummaries) { item in
                            SectorMark(
                                angle: .value("Amount", item.totalAmount),
                                innerRadius: .ratio(0.5),
                                angularInset: 2
                            )
                            .foregroundStyle(by: .value("Category", item.category))
                            .cornerRadius(4)
                        }
                        .frame(height: 200)
                        .chartLegend(position: .bottom, alignment: .leading)
                    } else {
                        // iOS 16 fallback: Bar chart instead of pie chart
                        Chart(analytics.categorySummaries) { item in
                            BarMark(
                                x: .value("Category", item.category),
                                y: .value("Amount", item.totalAmount)
                            )
                            .foregroundStyle(colorForCategory(item.category))
                        }
                        .frame(height: 200)
                    }

                    ForEach(analytics.categorySummaries.prefix(5)) { item in
                        HStack {
                            Circle()
                                .fill(colorForCategory(item.category))
                                .frame(width: 12, height: 12)

                            Text(item.category)
                                .font(Typography.subheadline)
                                .foregroundColor(ColorPalette.textPrimary)

                            Spacer()

                            VStack(alignment: .trailing, spacing: 2) {
                                Text(formatCurrency(item.totalAmount))
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
        )
    }

    private func sourceRankingSection(analytics: Analytics) -> some View {
        ChartCard(
            title: "Sources",
            content: AnyView(
                VStack(spacing: 12) {
                    ForEach(analytics.sourceSummaries.sorted(by: { $0.totalAmount > $1.totalAmount })) { item in
                        HStack {
                            Text(item.source)
                                .font(Typography.body)
                                .foregroundColor(ColorPalette.textPrimary)

                            Spacer()

                            Text(formatCurrency(item.totalAmount))
                                .font(Typography.bodyMedium.monospaced())
                                .foregroundColor(ColorPalette.textPrimary)

                            Text("(\(item.transactionCount))")
                                .font(Typography.caption)
                                .foregroundColor(ColorPalette.textSecondary)
                        }
                    }
                }
            )
        )
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
        // Create a simple mock API service
        class MockAPIService: APIServiceProtocol {
            func login(email: String, password: String) async throws -> User {
                User(email: email, isBiometricEnabled: false)
            }

            func register(email: String, password: String, name: String? = nil) async throws -> User {
                User(email: email, isBiometricEnabled: false)
            }

            func createTransaction(_ transaction: Transaction) async throws -> Transaction {
                transaction
            }

            func getTransactions(page: Int, limit: Int) async throws -> [Transaction] {
                []
            }

            func getAnalytics(period: TimePeriod) async throws -> Analytics {
                Analytics(
                    totalIncome: 5000,
                    totalExpenses: 3000,
                    netBalance: 2000,
                    totalTransactions: 10,
                    averageTransactionAmount: 500,
                    categorySummaries: [
                        CategorySummary(
                            category: "Food",
                            totalAmount: 1500,
                            transactionCount: 5,
                            type: .expense,
                            percentage: 50
                        )
                    ],
                    sourceSummaries: [
                        SourceSummary(
                            source: "Grab",
                            totalAmount: 1000,
                            transactionCount: 3,
                            type: .expense,
                            percentage: 33
                        )
                    ],
                    topCategories: [],
                    topSources: [],
                    timePeriod: .month,
                    dateRange: Date()...Date()
                )
            }

            func getSummary() async throws -> SummaryResponse {
                SummaryResponse(balance: 2000, totalIncome: 5000, totalExpenses: 3000)
            }
        }

        let viewModel = AnalyticsViewModel(apiService: MockAPIService())
        return AnalyticsView(viewModel: viewModel)
    }
}