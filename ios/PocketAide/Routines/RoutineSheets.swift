import DesignSystem
import PocketAideAPI
import SwiftUI

struct RoutineAddSheet: View {
    let startDay: String
    let onSave: (RoutineDraft) -> Void
    let onCancel: () -> Void

    @State private var name = ""
    @State private var cadence: RoutineCadence = .daily
    @State private var weekdays = 0
    @State private var monthDay = 1
    @State private var steps: [String] = [""]

    var body: some View {
        Sheet(area: .routines, onClose: onCancel) {
            ScrollView {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                    Text("새 루틴")
                        .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.routines))
                        .padding(.top, 8)
                        .accessibilityIdentifier("routines.sheet.title")
                    Card(area: .routines, padding: .medium) {
                        TextField("루틴 이름 — 예: 아침 루틴", text: $name)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                            .foregroundStyle(DesignTokens.Color.ink(.routines))
                            .accessibilityIdentifier("routines.sheet.name.field")
                    }
                    cadenceSection
                    stepsSection
                    actions
                        .padding(.top, 8)
                        .padding(.bottom, 32)
                }
                .padding(.horizontal, 24)
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxHeight: 560)
        }
    }

    private var cadenceSection: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            label("반복 주기")
            Picker("반복 주기", selection: $cadence) {
                ForEach(RoutineCadence.allCases, id: \.self) { option in
                    Text(option.displayName).tag(option)
                }
            }
            .pickerStyle(.segmented)
            .accessibilityIdentifier("routines.sheet.cadence.picker")
            switch cadence {
            case .daily:
                EmptyView()
            case .weekdays, .weekly:
                weekdayPicker
            case .monthly:
                Stepper("매월 \(monthDay)일", value: $monthDay, in: 1...31)
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                    .foregroundStyle(DesignTokens.Color.ink(.routines))
                    .accessibilityIdentifier("routines.sheet.monthday.stepper")
            }
        }
        .onChange(of: cadence) { _, newValue in
            if newValue == .weekly, weekdays.nonzeroBitCount != 1 {
                weekdays = 0
            }
        }
    }

    private var weekdayPicker: some View {
        HStack(spacing: 6) {
            ForEach(RoutineWeekdays.displayOrder, id: \.self) { index in
                let selected = RoutineWeekdays.contains(weekdays, index)
                Button {
                    toggleWeekday(index)
                } label: {
                    Text(RoutineWeekdays.symbols[index])
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold))
                        .frame(width: 34, height: 34)
                        .background(Circle().fill(selected ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.card(.routines)))
                        .overlay(Circle().stroke(DesignTokens.Color.rule(.routines), lineWidth: 1))
                        .foregroundStyle(selected ? DesignTokens.Color.surface(.routines) : DesignTokens.Color.ink(.routines))
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("routines.sheet.weekday.\(index)")
            }
        }
    }

    private var stepsSection: some View {
        VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
            label("단계")
            ForEach(steps.indices, id: \.self) { index in
                HStack(spacing: DesignTokens.Spacing.sm) {
                    Text("\(index + 1)")
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
                        .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                        .frame(width: 18)
                    TextField("단계 이름", text: $steps[index])
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                        .foregroundStyle(DesignTokens.Color.ink(.routines))
                        .accessibilityIdentifier("routines.sheet.step.\(index)")
                }
                .padding(.horizontal, DesignTokens.Spacing.md)
                .padding(.vertical, 10)
                .background(DesignTokens.Color.card(.routines))
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            }
            Button {
                steps.append("")
            } label: {
                Label("단계 추가", systemImage: "plus")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))
                    .foregroundStyle(DesignTokens.Color.accent(.routines))
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("routines.sheet.step.add.button")
        }
    }

    private func label(_ text: String) -> some View {
        Text(text)
            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold))
            .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.7))
    }

    private var actions: some View {
        VStack(spacing: 4) {
            Button(action: handleSave) {
                Text("저장")
                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(Capsule().fill(DesignTokens.Color.ink(.routines)))
                    .foregroundStyle(DesignTokens.Color.surface(.routines))
            }
            .buttonStyle(.plain)
            .disabled(!canSave)
            .opacity(canSave ? 1 : 0.5)
            .accessibilityIdentifier("routines.sheet.save.button")

            Button("취소", action: onCancel)
                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                .padding(.vertical, 8)
                .accessibilityIdentifier("routines.sheet.cancel.button")
        }
    }

    private var trimmedName: String {
        name.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private var canSave: Bool {
        guard !trimmedName.isEmpty else { return false }
        switch cadence {
        case .daily, .monthly:
            return true
        case .weekdays:
            return weekdays != 0
        case .weekly:
            return weekdays.nonzeroBitCount == 1
        }
    }

    private func toggleWeekday(_ index: Int) {
        let bit = RoutineWeekdays.bit(index)
        if cadence == .weekly {
            weekdays = bit
        } else {
            weekdays ^= bit
        }
    }

    private func handleSave() {
        guard canSave else { return }
        let cleaned = steps
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
        onSave(RoutineDraft(
            name: trimmedName,
            cadence: cadence,
            weekdays: cadence == .daily || cadence == .monthly ? 0 : weekdays,
            monthDay: cadence == .monthly ? monthDay : 0,
            startDay: startDay,
            steps: cleaned
        ))
    }
}

struct RoutineStepAddSheet: View {
    let routine: Routine
    let onSave: (String) -> Void
    let onDeleteStep: (RoutineStep) -> Void
    let onCancel: () -> Void

    @State private var title = ""

    var body: some View {
        Sheet(area: .routines, onClose: onCancel) {
            VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                Text("\(routine.name) · 단계")
                    .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                    .foregroundStyle(DesignTokens.Color.ink(.routines))
                    .padding(.top, 8)
                    .accessibilityIdentifier("routines.steps.sheet.title")
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {
                    ForEach(routine.steps) { step in
                        HStack {
                            Text(step.title)
                                .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))
                                .foregroundStyle(DesignTokens.Color.ink(.routines))
                            Spacer(minLength: 0)
                            Button {
                                onDeleteStep(step)
                            } label: {
                                Image(systemName: "minus.circle")
                                    .foregroundStyle(DesignTokens.Color.destructive(.routines))
                            }
                            .buttonStyle(.plain)
                            .accessibilityIdentifier("routines.steps.sheet.delete.\(step.id)")
                        }
                    }
                }
                Card(area: .routines, padding: .medium) {
                    TextField("새 단계", text: $title)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))
                        .foregroundStyle(DesignTokens.Color.ink(.routines))
                        .accessibilityIdentifier("routines.steps.sheet.field")
                }
                VStack(spacing: 4) {
                    Button(action: handleSave) {
                        Text("단계 추가")
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(Capsule().fill(DesignTokens.Color.ink(.routines)))
                            .foregroundStyle(DesignTokens.Color.surface(.routines))
                    }
                    .buttonStyle(.plain)
                    .disabled(trimmedTitle.isEmpty)
                    .opacity(trimmedTitle.isEmpty ? 0.5 : 1)
                    .accessibilityIdentifier("routines.steps.sheet.save.button")

                    Button("닫기", action: onCancel)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                        .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                        .padding(.vertical, 8)
                        .accessibilityIdentifier("routines.steps.sheet.cancel.button")
                }
                .padding(.top, 8)
                .padding(.bottom, 32)
            }
            .padding(.horizontal, 24)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private var trimmedTitle: String {
        title.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func handleSave() {
        guard !trimmedTitle.isEmpty else { return }
        onSave(trimmedTitle)
    }
}

struct RoutineHistorySheet: View {
    let routine: Routine
    let days: [RoutineHistoryDay]
    let onClose: () -> Void

    private let columns = Array(repeating: GridItem(.flexible(), spacing: 4), count: 10)

    var body: some View {
        Sheet(area: .routines, onClose: onClose) {
            ScrollView {
                VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("\(routine.name) · \(days.count)일 이력")
                            .font(DesignTokens.Typography.font(size: 18, weight: .bold))
                            .foregroundStyle(DesignTokens.Color.ink(.routines))
                            .accessibilityIdentifier("routines.history.title")
                        Text(summaryText)
                            .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))
                            .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                            .accessibilityIdentifier("routines.history.summary")
                    }
                    .padding(.top, 8)
                    LazyVGrid(columns: columns, spacing: 4) {
                        ForEach(days) { day in
                            RoundedRectangle(cornerRadius: 4, style: .continuous)
                                .fill(fill(for: day))
                                .overlay(
                                    RoundedRectangle(cornerRadius: 4, style: .continuous)
                                        .stroke(DesignTokens.Color.rule(.routines), lineWidth: day.scheduled ? 0 : 1)
                                )
                                .aspectRatio(1, contentMode: .fit)
                                .accessibilityIdentifier("routines.history.cell.\(day.day)")
                                .accessibilityValue(day.statusLabel)
                        }
                    }
                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(days.reversed()) { day in
                            HStack {
                                Text(RoutineDayFormat.title(of: day.day))
                                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm, weight: .semibold))
                                    .foregroundStyle(DesignTokens.Color.ink(.routines))
                                Spacer(minLength: 0)
                                Text(day.statusLabel)
                                    .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                                    .foregroundStyle(day.completed ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.ink(.routines).opacity(0.55))
                            }
                            .accessibilityElement(children: .combine)
                            .accessibilityIdentifier("routines.history.row.\(day.day)")
                        }
                    }
                    Button("닫기", action: onClose)
                        .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))
                        .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55))
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 8)
                        .padding(.bottom, 24)
                        .accessibilityIdentifier("routines.history.close.button")
                }
                .padding(.horizontal, 24)
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxHeight: 560)
        }
    }

    private var summaryText: String {
        let summary = RoutineHistorySummary(days)
        return "\(routine.scheduleSummary) · 예정 \(summary.scheduledDays)일 중 \(summary.completedDays)일 완료"
    }

    private func fill(for day: RoutineHistoryDay) -> Color {
        guard day.scheduled else { return .clear }
        if day.completed { return DesignTokens.Color.accent(.routines) }
        if day.done > 0 { return DesignTokens.Color.accent(.routines).opacity(0.4) }
        return DesignTokens.Color.soft(.routines)
    }
}
