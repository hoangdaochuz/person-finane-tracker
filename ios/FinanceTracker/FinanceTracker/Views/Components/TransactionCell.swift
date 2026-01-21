import SwiftUI

struct TransactionCell: View {
    let transaction: Transaction

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(backgroundColor)
                    .frame(width: 44, height: 44)

                Image(systemName: transaction.transactionType == .income ? "arrow.down" : "arrow.up")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(transaction.transactionType == .income ? ColorPalette.income : ColorPalette.expense)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(merchantName)
                    .font(Typography.bodyMedium)
                    .foregroundColor(ColorPalette.textPrimary)

                Text(categoryAndDate)
                    .font(Typography.caption)
                    .foregroundColor(ColorPalette.textSecondary)
            }

            Spacer()

            Text(amountText)
                .font(Typography.bodyBold.monospaced())
                .foregroundColor(transaction.transactionType == .income ? ColorPalette.income : ColorPalette.expense)
        }
        .padding(.vertical, 8)
    }

    private var merchantName: String {
        transaction.merchant ?? "Unknown"
    }

    private var categoryAndDate: String {
        var parts: [String] = []
        if let category = transaction.category {
            parts.append(category)
        }
        parts.append(formatDate(transaction.date))
        return parts.joined(separator: " • ")
    }

    private var amountText: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        let prefix = transaction.transactionType == .income ? "+" : "-"
        return prefix + (formatter.string(from: NSNumber(value: transaction.amount)) ?? "")
    }

    private var backgroundColor: Color {
        transaction.transactionType == .income
            ? ColorPalette.success.opacity(0.15)
            : ColorPalette.danger.opacity(0.15)
    }

    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        return formatter.string(from: date)
    }
}

struct TransactionCell_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 0) {
            TransactionCell(transaction: Transaction.preview(
                amount: 1500,
                type: .income,
                merchant: "Salary",
                category: "Income",
                source: "BCA",
                date: Date()
            ))
            TransactionCell(transaction: Transaction.preview(
                amount: 50,
                type: .expense,
                merchant: "Coffee Shop",
                category: "Food",
                source: "Gopay",
                date: Date()
            ))
        }
        .padding()
        .background(ColorPalette.background)
    }
}