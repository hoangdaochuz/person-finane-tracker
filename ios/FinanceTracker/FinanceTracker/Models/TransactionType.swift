//
//  TransactionType.swift
//  FinanceTracker
//
//  Created by Claude on 2025-01-20.
//

import Foundation

enum TransactionType: String, CaseIterable, Codable {
    case income = "income"
    case expense = "expense"
    case transfer = "transfer"

    var displayName: String {
        switch self {
        case .income:
            return "Income"
        case .expense:
            return "Expense"
        case .transfer:
            return "Transfer"
        }
    }
}