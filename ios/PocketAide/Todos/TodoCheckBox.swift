import DesignSystem
import PocketAideAPI
import SwiftUI

struct TodoCheckBox: View {
    let area: TodoArea
    let isDone: Bool

    private var tone: DesignTokens.Area { area.designArea }

    var body: some View {
        if area.checkIsRound {
            box(Circle())
        } else {
            box(Rectangle())
        }
    }

    private func box<S: InsettableShape>(_ shape: S) -> some View {
        Image(systemName: "checkmark")
            .font(.system(size: 12, weight: .bold))
            .foregroundStyle(.white)
            .opacity(isDone ? 1 : 0)
            .frame(width: area.checkSize, height: area.checkSize)
            .background(shape.fill(isDone ? DesignTokens.Color.accent(tone) : .clear))
            .overlay(shape.strokeBorder(DesignTokens.Color.accent(tone), lineWidth: 1.6))
            .contentShape(Rectangle())
    }
}
