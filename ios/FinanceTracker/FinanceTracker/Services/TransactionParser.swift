import Foundation

class TransactionParser {
    private var patterns: [String: Any] = [:]

    init() {
        loadPatterns()
    }

    func parse(notificationText: String, source: String) -> TransactionDTO? {
        // Try to extract amount, but allow parsing without it for category testing
        let amount = extractAmount(from: notificationText)
        let type = determineType(from: notificationText)
        let merchant = extractMerchant(from: notificationText)
        let category = extractCategory(from: notificationText, type: type)

        // Return nil only if absolutely no meaningful information can be extracted
        // Allow parsing with amount=0 if we have merchant or specific category
        if amount == nil {
            // No amount found - only return non-nil if we have merchant OR category (not "Uncategorized")
            if merchant == nil && (category == nil || category == "Uncategorized") {
                return nil
            }
        }

        return TransactionDTO(
            id: UUID(),
            amount: amount ?? 0,
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
        // Pattern 1: Vietnamese VND format with currency symbols (Rp 50.000, USD 1.500.000)
        let vndWithCurrency = "(?:Rp|USD|IDR)\\s+(\\d{1,3}(?:[,.]\\d{3})+)"

        if let result = extractUsingPattern(text, pattern: vndWithCurrency, isVND: true) {
            return result
        }

        // Pattern 2: Vietnamese format followed by vnd/d (50.000vnd, 1.500.000d)
        let vndWithSuffix = "(\\d{1,3}(?:[,.]\\d{3})+)\\s*(?:vnd|d|VND)(?:\\s|,|\\.|$)"

        if let result = extractUsingPattern(text, pattern: vndWithSuffix, isVND: true) {
            return result
        }

        // Pattern 3: English/USD format with decimal ($15.99, USD 50.00)
        let englishPattern = "(?:\\$|USD)\\s*([\\d,]+\\.\\d+)"

        if let result = extractUsingPattern(text, pattern: englishPattern, isVND: false) {
            return result
        }

        // Pattern 4: Generic number with separators (dot or comma)
        // Must not be followed by 9+ digits (account number)
        let genericPattern = "(?:^|\\s|\\D)(\\d{1,3}[,.]\\d{2,3})\\b(?!\\d)"

        if let result = extractUsingPattern(text, pattern: genericPattern, isVND: true) {
            // Only accept if result seems reasonable (avoid extracting account numbers)
            if result < 1000000 {  // Less than 1 million, likely a transaction amount
                return result
            }
        }

        // Pattern 5: Large plain numbers (6-8 digits) for VND without separators
        // Must be surrounded by non-digits or string boundaries
        let largeNumberPattern = "(?:^|\\D)(\\d{6,8})(?=\\D|$)"

        if let result = extractUsingPattern(text, pattern: largeNumberPattern, isVND: true) {
            return result
        }

        return nil
    }

    private func extractUsingPattern(_ text: String, pattern: String, isVND: Bool) -> Double? {
        guard let regex = try? NSRegularExpression(pattern: pattern) else {
            return nil
        }

        let nsString = text as NSString
        let fullRange = NSRange(location: 0, length: nsString.length)

        guard let match = regex.firstMatch(in: text, range: fullRange) else {
            return nil
        }

        // Find the first capture group that has content
        for i in 1..<match.numberOfRanges {
            let captureRange = match.range(at: i)
            if captureRange.location != NSNotFound {
                let amountString = nsString.substring(with: captureRange)

                if isVND {
                    // VND format: dots/commas are thousands separators
                    let cleaned = amountString.replacingOccurrences(of: ".", with: "")
                        .replacingOccurrences(of: ",", with: "")
                    if let amount = Double(cleaned), amount > 0 {
                        return amount
                    }
                } else {
                    // English format: commas are thousands separators, dots are decimal points
                    var cleaned = amountString.replacingOccurrences(of: ",", with: "")
                    // Check if there's a decimal point
                    if let dotIndex = cleaned.firstIndex(of: ".") {
                        let beforeDot = String(cleaned[..<dotIndex])
                        let afterDot = String(cleaned[dotIndex...])
                        if let value = Double(beforeDot + afterDot) {
                            return value
                        }
                    }
                    if let amount = Double(cleaned), amount > 0 {
                        return amount
                    }
                }
            }
        }

        return nil
    }

    private func determineType(from text: String) -> TransactionType {
        let lowercaseText = text.lowercased()

        // Check for expense keywords first (more specific patterns)
        // "nap the"/"nạp thẻ" indicate phone top-up (expense)
        if lowercaseText.contains("nap the") || lowercaseText.contains("nạp thẻ") {
            return .expense
        }

        // "topup"/"top up" indicate phone top-up (expense)
        if lowercaseText.contains("topup") || lowercaseText.contains("top up") {
            return .expense
        }

        // "chuyen" with "tien" indicates transfer (expense)
        if lowercaseText.contains("chuyen tien") || lowercaseText.contains("chuyển tiền") ||
           lowercaseText.contains("chuyen den") || lowercaseText.contains("chuyển đến") {
            return .expense
        }

        // Vietnamese income keywords
        let vietnameseIncomeKeywords = ["dc cong", "đc cộng", "được cộng", "nhan", "nhận",
                                               "nap tien", "nạp tiền"]  // "nap tien" = top up money (income)

        for keyword in vietnameseIncomeKeywords {
            if lowercaseText.contains(keyword) {
                return .income
            }
        }

        // English income keywords
        let englishIncomeKeywords = ["received", "credit", "deposit", "income", "salary"]

        for keyword in englishIncomeKeywords {
            if lowercaseText.contains(keyword) {
                return .income
            }
        }

        // Check plist patterns if available
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

        // Vietnamese expense keywords
        let vietnameseExpenseKeywords = ["da tru", "đã trừ", "thanh toan", "thanh toán",
                                             "chuyen", "chuyển", "mua", "mua hàng", "ma mathe"]

        for keyword in vietnameseExpenseKeywords {
            if lowercaseText.contains(keyword) {
                return .expense
            }
        }

        // "nap"/"nạp" alone (without "tien"/"tien" or "the"/"thẻ") - check context
        // If followed by "thanh cong" (success), it's income
        if lowercaseText.contains("nap thanh cong") || lowercaseText.contains("nạp thành công") {
            return .income
        }

        // Default to expense for unknown transactions
        return .expense
    }

    private func extractCategory(from text: String, type: TransactionType) -> String? {
        let lowercaseText = text.lowercased()

        // Special handling for phone top-up - should be Bills regardless of type
        if lowercaseText.contains("topup") || lowercaseText.contains("top up") ||
           lowercaseText.contains("nap the") || lowercaseText.contains("nạp thẻ") ||
           lowercaseText.contains("nap thanh cong") || lowercaseText.contains("nạp thành công") ||
           lowercaseText.contains("vao dt") || lowercaseText.contains("vào dt") {
            return "Bills"
        }

        // First check for "transfer" keyword which overrides shopping
        if lowercaseText.contains("transfer to") || lowercaseText.contains("chuyen tien") ||
           lowercaseText.contains("chuyển tiền") || lowercaseText.contains("chuyen den") ||
           lowercaseText.contains("chuyển đến") {
            return "Transfer"
        }

        // Define category keywords as array of tuples for proper iteration
        let categoryKeywords: [(keywords: [String], category: String)] = [
            // Income categories
            (["salary", "luong", "lương"], "Income"),
            (["refund", "hoan tra", "hoàn trả"], "Refund"),

            // Expense - Food & Beverages
            (["kfc", "lotteria", "jollibee", "highlands coffee", "coffee house", "the coffee house",
              "starbucks", "circle k", "food", "coffee", "restaurant", "cafe", "thuc an", "thức ăn"], "Food"),

            // Transportation
            (["grab", "gojek", "uber", "be", "xe", "transport", "di chuyen", "di chuyển",
              "chuyen di", "chuyển di"], "Transportation"),

            // Shopping
            (["shopee", "lazada", "tiki", "shopping", "mua hang", "mua hàng"], "Shopping"),

            // Bills & Utilities
            (["bill", "tien dien", "tiền điện", "hoa don dien", "hóa đơn điện",
              "tien nuoc", "tiền nước", "internet", "dien", "điện", "nuoc", "nước", "electric"], "Bills")
        ]

        // Iterate through all category keywords
        for (keywords, category) in categoryKeywords {
            for keyword in keywords {
                if lowercaseText.contains(keyword) {
                    // For income type, only return income-related categories
                    if type == .income && (category == "Income" || category == "Refund") {
                        return category
                    }
                    // For expense type, all categories except "Income" are valid
                    if type == .expense && category != "Income" {
                        return category
                    }
                }
            }
        }

        // Default categories based on type
        return type == .income ? "Income" : "Uncategorized"
    }

    private func extractMerchant(from text: String) -> String? {
        // Vietnamese merchant patterns using NSRegularExpression

        // Pattern 1: "tai MERCHANT" (at) - case-insensitive
        // Match word(s) after "tai" until special characters or end
        // Changed to greedy quantifier * instead of *? to capture full names
        let taiPattern = "(?i)tai\\s+([A-ZÀ-Ỹa-zà-ỹ]+(?:\\s+[A-ZÀ-Ỹa-zà-ỹ]+)*)(?=\\s*(?:vnd|d|VND|$|,|\\.|\\d))"

        if let regex = try? NSRegularExpression(pattern: taiPattern),
           let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
           let range = Range(match.range(at: 1), in: text) {
            let merchant = String(text[range]).trimmingCharacters(in: .whitespaces)
            if !merchant.isEmpty && merchant.count > 1 {
                return capitalizeFirstLetter(merchant)
            }
        }

        // Pattern 2: "tu MERCHANT" (from)
        let tuPattern = "(?i)tu\\s+([A-ZÀ-Ỹa-zà-ỹ]+(?:\\s+[A-ZÀ-Ỹa-zà-ỹ]+)*)(?=\\s*(?:vnd|d|VND|$|,|\\.|\\d))"

        if let regex = try? NSRegularExpression(pattern: tuPattern),
           let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
           let range = Range(match.range(at: 1), in: text) {
            let merchant = String(text[range]).trimmingCharacters(in: .whitespaces)
            if !merchant.isEmpty && merchant.count > 1 {
                return capitalizeFirstLetter(merchant)
            }
        }

        // Pattern 3: "den MERCHANT" (to)
        let denPattern = "(?i)den\\s+([A-ZÀ-Ỹa-zà-ỹ]+(?:\\s+[A-ZÀ-Ỹa-zà-ỹ]+)*)(?=\\s*(?:vnd|d|VND|$|,|\\.|\\d))"

        if let regex = try? NSRegularExpression(pattern: denPattern),
           let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
           let range = Range(match.range(at: 1), in: text) {
            let merchant = String(text[range]).trimmingCharacters(in: .whitespaces)
            if !merchant.isEmpty && merchant.count > 1 {
                return capitalizeFirstLetter(merchant)
            }
        }

        // English pattern: "at/from/to MERCHANT"
        let englishPattern = "(?i)(?:at|from|to)\\s+([A-Z][A-Za-z\\s]+?)(?:\\s+on|\\s*$|\\.|\\s+\\d)"

        if let regex = try? NSRegularExpression(pattern: englishPattern),
           let match = regex.firstMatch(in: text, range: NSRange(text.startIndex..., in: text)),
           let range = Range(match.range(at: 1), in: text) {
            return String(text[range]).trimmingCharacters(in: .whitespaces)
        }

        return nil
    }

    private func capitalizeFirstLetter(_ text: String) -> String {
        guard !text.isEmpty else { return text }
        return text.prefix(1).capitalized + text.dropFirst()
    }
}
