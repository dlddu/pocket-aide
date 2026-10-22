import DesignSystem
import PocketAideAPI
import SwiftUI

struct SessionCreateSheet: View {
    let config: AgentSessionConfig?
    let isCreating: Bool
    let error: AgentSessionError?
    let onCreate: (String, AgentSessionWorkload, String?) -> Void
    let onCancel: () -> Void

    @State private var name = ""
    @State private var workload: AgentSessionWorkload = .claudeCode
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
                if workload.takesModel {
                    field(label: "모델") { modelInput }
                }
                Text(workload.takesModel
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

    private var workloadPicker: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("타입")
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.7))
            Picker("타입", selection: $workload) {
                ForEach(AgentSessionWorkload.allCases, id: \.self) { kind in
                    Text(kind.label).tag(kind)
                }
            }
            .pickerStyle(.segmented)
            .accessibilityIdentifier("sessions.create.workload.picker")
            Text(workload.summary)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.6))
                .accessibilityIdentifier("sessions.create.workload.summary")
        }
    }

    @ViewBuilder
    private var modelInput: some View {
        if let config, !config.models.isEmpty {
            Picker("모델", selection: Binding(
                get: { pickedModel ?? defaultPick(config) },
                set: { pickedModel = $0 }
            )) {
                ForEach(config.models, id: \.self) { model in
                    Text(model).tag(model)
                }
            }
            .pickerStyle(.menu)
            .tint(DesignTokens.Color.ink(area))
            .accessibilityIdentifier("sessions.create.model.picker")
        } else {
            TextField(modelPlaceholder, text: $typedModel)
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
            .disabled(trimmedName.isEmpty || isCreating)
            .opacity(trimmedName.isEmpty ? 0.5 : 1)
            .accessibilityIdentifier("sessions.create.save.button")

            Button("취소", action: onCancel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))
                .padding(.vertical, 8)
                .accessibilityIdentifier("sessions.create.cancel.button")
        }
    }

    private var modelPlaceholder: String {
        guard let model = config?.defaultModel, !model.isEmpty else { return "platform-default" }
        return model
    }

    private var trimmedName: String {
        name.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func defaultPick(_ config: AgentSessionConfig) -> String {
        config.models.contains(config.defaultModel) ? config.defaultModel : config.models[0]
    }

    private var chosenModel: String? {
        guard workload.takesModel else { return nil }
        if let config, !config.models.isEmpty {
            return pickedModel ?? defaultPick(config)
        }
        return typedModel
    }

    private func createFailureMessage(_ error: AgentSessionError) -> String {
        switch error {
        case .unknown, .disabled:
            return "세션을 띄우지 못했어요. 세션 플랫폼에 「\(workload.label)」 타입 설정(이미지·자격증명)이 있는지 확인해 주세요."
        default:
            return error.message
        }
    }

    private func handleCreate() {
        guard !trimmedName.isEmpty, !isCreating else { return }
        onCreate(trimmedName, workload, chosenModel)
    }
}
