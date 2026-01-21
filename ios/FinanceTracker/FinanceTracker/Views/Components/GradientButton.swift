import SwiftUI

struct GradientButton: View {
    let title: String
    let action: () -> Void
    var isDisabled: Bool = false

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(Typography.bodyMedium)
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .padding()
                .background(
                    ColorPalette.primaryGradient
                )
                .cornerRadius(12)
        }
        .disabled(isDisabled)
        .opacity(isDisabled ? 0.5 : 1.0)
    }
}

struct GradientButton_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: 16) {
            GradientButton(title: "Get Started", action: {})
            GradientButton(title: "Disabled", action: {}, isDisabled: true)
        }
        .padding()
    }
}