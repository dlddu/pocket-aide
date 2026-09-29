import DesignSystem
import PocketAideAPI
import SwiftUI

struct TodoEditSheet: View {
    enum Mode: Equatable {
        case create
        case edit(TodoItem)
    }

    /// FilterPills needs a non-optional selection; `.unset` stands for "no priority".
    enum PriorityChoice: String, CaseIterable, Hashable {
        case unset, high, normal, low

        init(_ priority: TodoPriority?) {
            self = priority.flatMap { PriorityChoice(rawValue: $0.rawValue) } ?? .unset
        }

        var priority: TodoPriority? { TodoPriority(rawValue: rawValue) }

        var displayName: String { priority?.displayName ?? "없음" }
    }

    let area: TodoArea
    let mode: Mode
    let onSave: (TodoDraft) -> Void
    let onCancel: () -> Void
    let onDelete: (() -> Void)?

    @State private var title: String
    @State private var memo: String
    @State private var hasDueDate: Bool
    @State private var dueDate: Date
    @State private var priority: PriorityChoice
    private let done: Bool

    init(
        area: TodoArea,
        mode: Mode,
        onSave: @escaping (TodoDraft) -> Void,
        onCancel: @escaping () -> Void,
        onDelete: (() -> Void)? = nil
    ) {
        self.area = area
        self.mode = mode
        self.onSave = onSave
        self.onCancel = onCancel
        self.onDelete = onDelete
        let existing: TodoItem? = {
            if case .edit(let item) = mode { return item }
            return nil
        }()
        let parsedDue = existing?.dueDate.flatMap(TodoDueDate.date(from:))
        _title = State(initialValue: existing?.title ?? "")
        _memo = State(initialValue: existing?.memo ?? "")
        _hasDueDate = State(initialValue: parsedDue != nil)
        _dueDate = State(initialValue: parsedDue ?? Date())
        _priority = State(initialValue: PriorityChoice(existing?.priority))
        done = existing?.isDone ?? false
    }

    private var tone: DesignTokens.Area { area.designArea }

    var body: some View {
        Sheet(area: tone, onClose: onCancel) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                Text(isEditing ? "할 일 편집" : "새 할 일")
                    .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(tone))
                    .padding(.top, 8)
                    .accessibilityIdentifier("todo.sheet.title")
                fields
                prioritySection
                actions
                    .padding(.top, 8)
                    .padding(.bottom, 32)
            }
            .padding(.horizontal, 24)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var fields: some View {
        Card(area: tone, padding: .medium) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) {
                TextField("제목", text: $title)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.titleMd, weight: .semibold))
                    .accessibilityIdentifier("todo.sheet.title.field")
                TextField("메모", text: $memo, axis: .vertical)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .lineLimit(1...4)
                    .accessibilityIdentifier("todo.sheet.memo.field")
                Toggle("마감일", isOn: $hasDueDate.animation())
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .medium))
                    .tint(DesignTokens.Color.accent(tone))
                    .accessibilityIdentifier("todo.sheet.due.toggle")
                if hasDueDate {
                    DatePicker("날짜", selection: $dueDate, displayedComponents: .date)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                        .accessibilityIdentifier("todo.sheet.due.picker")
                }
            }
            .foregroundStyle(DesignTokens.Color.ink(tone))
        }
    }

    private var prioritySection: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            AreaLabel(area: tone, text: "우선순위 (선택)")
            FilterPills(area: tone, options: PriorityChoice.allCases, selection: $priority) { option in
                Text(option.displayName)
                    .font(DesignTokens.Typography.font(
                        size: DesignTokens.Typography.body,
                        weight: option == priority ? .bold : .medium
                    ))
            }
        }
    }

    private var actions: some View {
        VStack(spacing: 4) {
            Button(action: handleSave) {
                Text("저장")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(Capsule().fill(DesignTokens.Color.ink(tone)))
                    .foregroundStyle(DesignTokens.Color.surface(tone))
            }
            .buttonStyle(.plain)
            .disabled(trimmedTitle.isEmpty)
            .opacity(trimmedTitle.isEmpty ? 0.5 : 1)
            .accessibilityIdentifier("todo.sheet.save.button")

            if isEditing, let onDelete {
                Button(action: onDelete) {
                    Text("삭제")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .semibold))
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .overlay(Capsule().stroke(DesignTokens.Color.destructive(tone), lineWidth: 1.4))
                        .foregroundStyle(DesignTokens.Color.destructive(tone))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("todo.sheet.delete.button")
            }

            Button("취소", action: onCancel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(tone).opacity(0.55))
                .padding(.vertical, 8)
                .accessibilityIdentifier("todo.sheet.cancel.button")
        }
    }

    private var isEditing: Bool {
        if case .edit = mode { return true }
        return false
    }

    private var trimmedTitle: String {
        title.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func handleSave() {
        guard !trimmedTitle.isEmpty else { return }
        onSave(TodoDraft(
            title: trimmedTitle,
            memo: memo.trimmingCharacters(in: .whitespacesAndNewlines),
            dueDate: hasDueDate ? TodoDueDate.string(from: dueDate) : nil,
            priority: priority.priority,
            done: done
        ))
    }
}
