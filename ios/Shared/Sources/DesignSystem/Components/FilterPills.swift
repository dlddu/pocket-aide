import SwiftUI

public struct FilterPills<Option: Hashable, Label: View>: View {
    private let area: DesignTokens.Area
    private let options: [Option]
    @Binding private var selection: Option
    private let label: (Option) -> Label

    public init(
        area: DesignTokens.Area,
        options: [Option],
        selection: Binding<Option>,
        @ViewBuilder label: @escaping (Option) -> Label
    ) {
        self.area = area
        self.options = options
        self._selection = selection
        self.label = label
    }

    public var body: some View {
        HStack(spacing: DesignTokens.Spacing.sm) {
            ForEach(options, id: \.self) { option in
                let active = option == selection
                Button {
                    selection = option
                } label: {
                    label(option)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 10)
                        .background(
                            Capsule(style: .continuous)
                                .fill(active ? DesignTokens.Color.soft(area) : Color.clear)
                        )
                        .overlay(
                            Capsule(style: .continuous)
                                .stroke(
                                    active ? DesignTokens.Color.accent(area).opacity(0.4) : DesignTokens.Color.rule(area),
                                    lineWidth: active ? 1.5 : 1
                                )
                        )
                        .foregroundStyle(active ? DesignTokens.Color.ink(area) : DesignTokens.Color.ink(area).opacity(0.55))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("filter.pill.\(String(describing: option))")
                .accessibilityAddTraits(active ? .isSelected : [])
            }
        }
    }
}
