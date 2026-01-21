import Foundation

class TransactionParser {
    private var patterns: [String: Any] = [:]

    init() {
        loadPatterns()
    }

    func parse(notificationText: String, source: String) -> TransactionDTO? {
        guard let amount = extractAmount(from: notificationText) else {
            return nil
        }

        let type = determineType(from: notificationText)
        let merchant = extractMerchant(from: notificationText)
        let category = extractCategory(from: notificationText, type: type)

        return TransactionDTO(
            id: UUID(),
            amount: amount,
            type: type.rawValue,
            merchant: merchant,
            category: category,
            source: source,
            date: Date(),
            remoteID: nil
        )
    }

    private func loadPatterns() {
        guard let path = Bundle.main.path(forResource: "NotificationPatterns", ofType: "plist"),
              let plist = NSDictionary(contentsOfFile: path),
              let patternsDict = plist["Patterns"] as? [String: Any] else {
            return
        }
        patterns = patternsDict
    }

    private func extractAmount(from text: String) -> Double? {
        let pattern = "(?:Rp|IDR|USD|\\$)?\\s*([\\d,]+\\.?\\d*)"

        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
              let range = Range(match.range(at: 1), in: text) else {
            return nil
        }

        let amountString = String(text[range])
            .replacingOccurrences(of: ",", with: "")
        return Double(amountString)
    }

    private func determineType(from text: String) -> TransactionType {
        let lowercaseText = text.lowercased()

        if let incomeKeywords = patterns["IncomeKeywords"] as? [String] {
            for keyword in incomeKeywords where lowercaseText.contains(keyword.lowercased()) {
                return .income
            }
        }

        if let expenseKeywords = patterns["ExpenseKeywords"] as? [String] {
            for keyword in expenseKeywords where lowercaseText.contains(keyword.lowercased()) {
                return .expense
            }
        }

        return .expense
    }

    private func extractCategory(from text: String, type: TransactionType) -> String? {
        let lowercaseText = text.lowercased()

        // Category keywords based on transaction type
        let categoryKeywords: [TransactionType: [String: String]] = [
            .income: ["salary": "Income", "transfer": "Transfer", "refund": "Refund"],
            .expense: [
                "food": "Food",
                "coffee": "Food",
                "restaurant": "Food",
                "grab": "Food",
                "gojek": "Transportation",
                "transport": "Transportation",
                "uber": "Transportation",
                "shopping": "Shopping",
                "transfer": "Transfer",
                "bill": "Bills",
                "electric": "Bills",
                "internet": "Bills"
            ]
        ]

        if let categories = categoryKeywords[type] {
            for (keyword, category) in categories where lowercaseText.contains(keyword) {
                return category
            }
        }

        return type == .income ? "Income" : "Uncategorized"
    }

    private func extractMerchant(from text: String) -> String? {
        let pattern = "(?:at|from|to)\\s+([A-Z][A-Za-z\\s]+?)(?:\\s+on|\\s*$|\\.)"

        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
              let range = Range(match.range(at: 1), in: text) else {
            return nil
        }

        return String(text[range]).trimmingCharacters(in: .whitespaces)
    }
}
