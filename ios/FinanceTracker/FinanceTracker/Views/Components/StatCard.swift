import SwiftUI

struct StatCard: View {
    let icon: String
    let title: String
    let value: String
    let trend: String?
    let isPositive: Bool

    init(
        icon: String,
        title: String,
        value: String,
        trend: String? = nil,
        isPositive: Bool = true
    ) {
        self.icon = icon
        self.title = title
        self.value = value
        self.trend = trend
        self.isPositive = isPositive
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: icon)
                    .font(.title3)
                    .foregroundColor(.white)
                    .frame(width: 40, height: 40)
                    .background(ColorPalette.primaryGradient)
                    .cornerRadius(10)

                Spacer()

                if let trend = trend {
                    HStack(spacing: 4) {
                        Image(systemName: isPositive ? "arrow.up.right" : "arrow.down.right")
                        Text(trend)
                    }
                    .font(Typography.caption)
                    .foregroundColor(isPositive ? ColorPalette.income : ColorPalette.expense)
                }
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(Typography.caption)
                    .foregroundColor(ColorPalette.textSecondary)

                Text(value)
                    .font(Typography.title2)
                    .foregroundColor(ColorPalette.textPrimary)
            }
        }
        .padding(16)
        .background(ColorPalette.cardBackground)
        .cornerRadius(16)
        .shadow(color: .black.opacity(0.05), radius: 10, x: 0, y: 4)
    }
}

struct StatCard_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 16) {
            StatCard(
                icon: "dollarsign.circle.fill",
                title: "Balance",
                value: "$2,450.00",
                trend: "+12%",
                isPositive: true
            )
            StatCard(
                icon: "arrow.down.circle.fill",
                title: "Income",
                value: "$5,000.00",
                trend: "+8%",
                isPositive: true
            )
            StatCard(
                icon: "arrow.up.circle.fill",
                title: "Expenses",
                value: "$2,550.00",
                trend: "-3%",
                isPositive: true
            )
        }
        .padding()
        .background(ColorPalette.background)
    }
}