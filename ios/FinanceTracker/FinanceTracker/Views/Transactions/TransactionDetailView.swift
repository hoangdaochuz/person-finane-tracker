import SwiftUI

struct TransactionDetailView: View {
    let transaction: Transaction
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationView {
            VStack(spacing: 24) {
                Spacer()

                VStack(spacing: 8) {
                    Image(systemName: TransactionType(rawValue: transaction.type) == .income ? "arrow.down.circle.fill" : "arrow.up.circle.fill")
                        .font(.system(size: 60))
                        .foregroundColor(TransactionType(rawValue: transaction.type) == .income ? ColorPalette.income : ColorPalette.expense)

                    Text(amountText)
                        .font(Typography.largeTitle.monospaced())
                        .foregroundColor(TransactionType(rawValue: transaction.type) == .income ? ColorPalette.income : ColorPalette.expense)

                    Text((TransactionType(rawValue: transaction.type) ?? .expense).displayName)
                        .font(Typography.subheadline)
                        .foregroundColor(ColorPalette.textSecondary)
                }

                Divider()

                VStack(spacing: 16) {
                    DetailRow(label: "Merchant", value: transaction.merchant ?? "Unknown")
                    DetailRow(label: "Category", value: transaction.category ?? "Uncategorized")
                    DetailRow(label: "Source", value: transaction.source ?? "Unknown")
                    DetailRow(label: "Date", value: formatDate(transaction.date))
                    DetailRow(label: "Transaction ID", value: transaction.id.uuidString)
                }
                .padding()

                Spacer()
            }
            .padding()
            .navigationTitle("Transaction Details")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }

    private var amountText: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "USD"
        let prefix = TransactionType(rawValue: transaction.type) == .income ? "+" : "-"
        return prefix + (formatter.string(from: NSNumber(value: transaction.amount)) ?? "")
    }

    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .long
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

struct DetailRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .font(Typography.subheadline)
                .foregroundColor(ColorPalette.textSecondary)

            Spacer()

            Text(value)
                .font(Typography.body)
                .foregroundColor(ColorPalette.textPrimary)
        }
    }
}

struct TransactionDetailView_Previews: PreviewProvider {
    static var previews: some View {
        TransactionDetailView(
            transaction: Transaction.preview(
                amount: 1500,
                type: .income,
                merchant: "Salary",
                category: "Income",
                source: "BCA",
                date: Date()
            )
        )
    }
}