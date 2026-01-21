import SwiftUI

enum Typography {
    // Headings
    static let largeTitle = Font.system(size: 28, weight: .bold)
    static let title1 = Font.system(size: 24, weight: .bold)
    static let title2 = Font.system(size: 20, weight: .bold)
    static let title3 = Font.system(size: 18, weight: .semibold)

    // Body
    static let body = Font.system(size: 17, weight: .regular)
    static let bodyMedium = Font.system(size: 17, weight: .medium)
    static let bodyBold = Font.system(size: 17, weight: .bold)

    // Subhead
    static let subheadline = Font.system(size: 15, weight: .regular)
    static let subheadlineMedium = Font.system(size: 15, weight: .medium)

    // Caption
    static let caption = Font.system(size: 12, weight: .regular)
    static let captionMedium = Font.system(size: 12, weight: .medium)

    // Currency (monospaced for alignment)
    static let currency = Font.system(size: 17, weight: .semibold).monospaced()
    static let currencyLarge = Font.system(size: 24, weight: .bold).monospaced()
}