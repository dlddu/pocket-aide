import DesignSystem
import PocketAideAPI
import SwiftUI

struct SessionCreateSheet: View {
    let config: AgentSessionConfig?
    let isCreating: Bool
    let error: AgentSessionError?
    let onCreate: (String, AgentSessionWorkloadType, String?) -> Void
    let onCancel: () -> Void

    @State private var name = ""
    @State private var pickedType: String?
    @State private var pickedModel: String?
    @State private var typedModel = ""

    private let area = SessionsView.area

    var body: some View {
        Sheet(area: area, onClose: onCancel) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                Text("새 세션")
                    .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(area))
                    .padding(.top, 8)
                    .accessibilityIdentifier("sessions.create.title")
                field(label: "이름") {
                    TextField("예: 테스트 실패 원인 찾기", text: $name)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                        .foregroundStyle(DesignTokens.Color.ink(area))
                        .accessibilityIdentifier("sessions.create.name.field")
                }
                workloadPicker
                if let catalog = workload?.model {
                    field(label: "모델") { modelInput(catalog) }
                }
                Text(workload?.takesModel == true
                     ? "타입과 모델은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요."
                     : "타입은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요.")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                    .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.6))
                    .fixedSize(horizontal: false, vertical: true)
                if let error {
                    Text(createFailureMessage(error))
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm))
                        .foregroundStyle(DesignTokens.StatusColor.failure)
                        .fixedSize(horizontal: false, vertical: true)
                        .accessibilityIdentifier("sessions.create.error")
                }
                actions
                    .padding(.top, 8)
                    .padding(.bottom, 32)
            }
            .padding(.horizontal, 24)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    @ViewBuilder
    private var workloadPicker: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("타입")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.7))
            if let types = config?.workloadTypes, !types.isEmpty {
                Card(area: area, padding: .medium) {
                    VStack(spacing: 0) {
                        ForEach(Array(types.enumerated()), id: \.element.id) { index, kind in
                            if index > 0 {
                                Rectangle()
                                    .fill(DesignTokens.Color.rule(area))
                                    .frame(height: 1)
                            }
                            typeRow(kind)
                        }
                    }
                }
            } else {
                Card(area: area, padding: .medium) {
                    HStack(spacing: DesignTokens.Spacing.sm) {
                        if config == nil {
                            ProgressView()
                        }
                        Text(config == nil ? "타입 목록을 불러오는 중…" : "세션 플랫폼이 제공하는 타입이 없어요.")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.6))
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                }
                .accessibilityIdentifier("sessions.create.workload.empty")
            }
        }
    }

    private func typeRow(_ kind: AgentSessionWorkloadType) -> some View {
        let selected = kind.id == workload?.id
        return Button {
            pickedType = kind.id
            pickedModel = nil
            typedModel = ""
        } label: {
            HStack(alignment: .top, spacing: DesignTokens.Spacing.md) {
                Image(systemName: selected ? "largecircle.fill.circle" : "circle")
                    .foregroundStyle(selected ? DesignTokens.Color.accent(area) : DesignTokens.Color.ink(area).opacity(0.4))
                VStack(alignment: .leading, spacing: 2) {
                    Text(kind.displayName)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                        .foregroundStyle(DesignTokens.Color.ink(area))
                    if let summary = kind.description, !summary.isEmpty {
                        Text(summary)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                            .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.6))
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }
                Spacer(minLength: 0)
            }
            .padding(.vertical, DesignTokens.Spacing.sm)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(selected ? .isSelected : [])
        .accessibilityIdentifier("sessions.create.workload.\(kind.id)")
    }

    @ViewBuilder
    private func modelInput(_ catalog: AgentSessionModelCatalog) -> some View {
        if let initial = catalog.initialPick {
            Picker("모델", selection: Binding(
                get: { pickedModel ?? initial },
                set: { pickedModel = $0 }
            )) {
                ForEach(catalog.models, id: \.self) { model in
                    Text(model).tag(model)
                }
            }
            .pickerStyle(.menu)
            .tint(DesignTokens.Color.ink(area))
            .accessibilityIdentifier("sessions.create.model.picker")
        } else {
            TextField(catalog.defaultModel.isEmpty ? "platform-default" : catalog.defaultModel, text: $typedModel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                .foregroundStyle(DesignTokens.Color.ink(area))
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .accessibilityIdentifier("sessions.create.model.field")
        }
    }

    private func field<Content: View>(label: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(label)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.7))
            Card(area: area, padding: .medium) {
                content()
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
    }

    private var actions: some View {
        VStack(spacing: 4) {
            Button(action: handleCreate) {
                HStack(spacing: DesignTokens.Spacing.sm) {
                    if isCreating {
                        ProgressView()
                            .tint(DesignTokens.Color.surface(area))
                    }
                    Text(isCreating ? "세션 띄우는 중…" : "만들기")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 14)
                .background(Capsule().fill(DesignTokens.Color.ink(area)))
                .foregroundStyle(DesignTokens.Color.surface(area))
            }
            .buttonStyle(.plain)
            .disabled(!canCreate)
            .opacity(canCreate ? 1 : 0.5)
            .accessibilityIdentifier("sessions.create.save.button")

            Button("취소", action: onCancel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .padding(.vertical, 8)
                .accessibilityIdentifier("sessions.create.cancel.button")
        }
    }

    private var workload: AgentSessionWorkloadType? {
        guard let types = config?.workloadTypes, !types.isEmpty else { return nil }
        return types.first { $0.id == pickedType } ?? types[0]
    }

    private var trimmedName: String {
        name.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private var canCreate: Bool {
        !trimmedName.isEmpty && workload != nil && !isCreating
    }

    private var chosenModel: String? {
        guard let catalog = workload?.model else { return nil }
        if let initial = catalog.initialPick {
            return pickedModel ?? initial
        }
        return typedModel
    }

    private func createFailureMessage(_ error: AgentSessionError) -> String {
        switch error {
        case .unknown, .disabled:
            let name = workload?.displayName ?? "선택한"
            return "세션을 띄우지 못했어요. 세션 플랫폼에 「\(name)」 타입 설정(이미지·자격증명)이 있는지 확인해 주세요."
        default:
            return error.message
        }
    }

    private func handleCreate() {
        guard canCreate, let workload else { return }
        onCreate(trimmedName, workload, chosenModel)
    }
}
