import SwiftUI

public struct DSFilterPills<Option: Hashable, OptionContent: View>: View {
    private let area: DesignTokens.Area
    private let options: [Option]
    @Binding private var selection: Option
    private let optionContent: (Option, Bool) -> OptionContent

    public init(
        area: DesignTokens.Area,
        options: [Option],
        selection: Binding<Option>,
        @ViewBuilder optionContent: @escaping (Option, Bool) -> OptionContent
    ) {
        self.area = area
        self.options = options
        self._selection = selection
        self.optionContent = optionContent
    }

    public var body: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            ForEach(options, id: \.self) { option in
                let isActive = option == selection
                Button {
                    selection = option
                } label: {
                    optionContent(option, isActive)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, DesignTokens.Spacing.md)
                        .background(
                            RoundedRectangle(cornerRadius: DesignTokens.Radius.chip)
                                .fill(
                                    isActive
                                        ? DesignTokens.Color.soft(area)
                                        : Color.clear
                                )
                        )
                        .overlay(
                            RoundedRectangle(cornerRadius: DesignTokens.Radius.chip)
                                .stroke(
                                    isActive
                                        ? DesignTokens.Color.accent(area).opacity(0.4)
                                        : DesignTokens.Color.rule(area),
                                    lineWidth: 1
                                )
                        )
                        .foregroundStyle(
                            isActive
                                ? DesignTokens.Color.ink(area)
                                : DesignTokens.Color.ink(area).opacity(0.65)
                        )
                }
                .buttonStyle(.plain)
            }
        }
    }
}
