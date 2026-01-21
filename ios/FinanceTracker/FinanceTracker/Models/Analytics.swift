//
//  Analytics.swift
//  FinanceTracker
//
//  Created by Claude on 2025-01-20.
//

import Foundation

// MARK: - Time Period Enum
enum TimePeriod: String, CaseIterable, Codable {
    case threeMonths = "three_months"
    case week = "week"
    case month = "month"
    case year = "year"
    case all = "all"

    var displayName: String {
        switch self {
        case .threeMonths:
            return "Last 3 Months"
        case .week:
            return "This Week"
        case .month:
            return "This Month"
        case .year:
            return "This Year"
        case .all:
            return "All Time"
        }
    }
}

// MARK: - Category Summary
struct CategorySummary: Identifiable, Codable {
    let id = UUID()
    let category: String
    let totalAmount: Double
    let transactionCount: Int
    let type: TransactionType
    let percentage: Double

    // Computed property for average amount
    var averageAmount: Double {
        return transactionCount > 0 ? totalAmount / Double(transactionCount) : 0
    }

    // Computed property for formatted total
    var formattedTotal: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        return formatter.string(from: NSNumber(value: totalAmount)) ?? "$0.00"
    }
}

// MARK: - Source Summary
struct SourceSummary: Identifiable, Codable {
    let id = UUID()
    let source: String
    let totalAmount: Double
    let transactionCount: Int
    let type: TransactionType
    let percentage: Double

    // Computed property for average amount
    var averageAmount: Double {
        return transactionCount > 0 ? totalAmount / Double(transactionCount) : 0
    }

    // Computed property for formatted total
    var formattedTotal: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        return formatter.string(from: NSNumber(value: totalAmount)) ?? "$0.00"
    }
}

// MARK: - Analytics
struct Analytics: Codable {
    let totalIncome: Double
    let totalExpenses: Double
    let netBalance: Double
    let totalTransactions: Int
    let averageTransactionAmount: Double
    let categorySummaries: [CategorySummary]
    let sourceSummaries: [SourceSummary]
    let topCategories: [CategorySummary]
    let topSources: [SourceSummary]
    let timePeriod: TimePeriod
    let dateRange: ClosedRange<Date>

    // Computed properties for backward compatibility
    var categoryBreakdown: [CategorySummary] {
        return categorySummaries
    }

    var sourceBreakdown: [SourceSummary] {
        return sourceSummaries
    }

    // Computed property for formatted income
    var formattedIncome: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        return formatter.string(from: NSNumber(value: totalIncome)) ?? "$0.00"
    }

    // Computed property for formatted expenses
    var formattedExpenses: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        return formatter.string(from: NSNumber(value: totalExpenses)) ?? "$0.00"
    }

    // Computed property for formatted net balance
    var formattedNetBalance: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        return formatter.string(from: NSNumber(value: netBalance)) ?? "$0.00"
    }

    // Computed property for formatted average transaction
    var formattedAverageTransaction: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        return formatter.string(from: NSNumber(value: averageTransactionAmount)) ?? "$0.00"
    }

    // Computed property for savings rate
    var savingsRate: Double {
        let totalFlow = totalIncome + totalExpenses
        return totalFlow > 0 ? (netBalance / totalFlow) * 100 : 0
    }

    // Computed property for income to expense ratio
    var incomeExpenseRatio: Double {
        return totalExpenses > 0 ? totalIncome / totalExpenses : 0
    }
}