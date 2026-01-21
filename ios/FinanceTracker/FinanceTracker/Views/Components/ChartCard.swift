import SwiftUI
import Charts

struct ChartCard: View {
    let title: String
    let content: AnyView

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text(title)
                .font(Typography.title3)
                .foregroundColor(ColorPalette.textPrimary)

            content
        }
        .padding(16)
        .background(ColorPalette.cardBackground)
        .cornerRadius(16)
        .shadow(color: .black.opacity(0.05), radius: 10, x: 0, y: 4)
    }
}

struct ChartCard_Previews: PreviewProvider {
    static var previews: some View {
        ChartCard(
            title: "Income vs Expenses",
            content: AnyView(
                Text("Chart content here")
                    .frame(height: 200)
            )
        )
        .padding()
        .background(ColorPalette.background)
    }
}