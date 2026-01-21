import SwiftUI

enum ColorPalette {
    // Primary Gradient Colors
    static let primaryIndigo = Color(red: 0.39, green: 0.4, blue: 0.95)
    static let primaryViolet = Color(red: 0.55, green: 0.36, blue: 0.96)

    // Status Colors
    static let success = Color(red: 0.06, green: 0.73, blue: 0.51)
    static let danger = Color(red: 0.94, green: 0.27, blue: 0.27)
    static let warning = Color(red: 1.0, green: 0.73, blue: 0.0)

    // Background Colors
    static let background = Color(red: 0.98, green: 0.98, blue: 0.98)
    static let cardBackground = Color.white

    // Text Colors
    static let textPrimary = Color(red: 0.07, green: 0.09, blue: 0.15)
    static let textSecondary = Color(red: 0.42, green: 0.45, blue: 0.44)

    // Gradient
    static let primaryGradient = LinearGradient(
        colors: [primaryIndigo, primaryViolet],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    // Semantic colors
    static let income = success
    static let expense = danger
}