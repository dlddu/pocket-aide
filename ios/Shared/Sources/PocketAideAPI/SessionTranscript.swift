import Foundation

public struct AgentSessionTurn: Equatable, Sendable, Identifiable {
    public enum Status: Equatable, Sendable {
        case running
        case ok
        case error
        case timeout
        case canceled
        case other(String)

        init(marker value: String) {
            switch value {
            case "ok": self = .ok
            case "error": self = .error
            case "timeout": self = .timeout
            case "canceled": self = .canceled
            default: self = .other(value)
            }
        }

        public var label: String {
            switch self {
            case .running: return "진행 중"
            case .ok: return "완료"
            case .error: return "실패"
            case .timeout: return "시간 초과"
            case .canceled: return "취소됨"
            case .other(let raw): return raw
            }
        }

        public var isFinished: Bool { self != .running }
    }

    public let seq: Int
    public internal(set) var input: String
    public internal(set) var hasInput: Bool
    public internal(set) var output: String
    public internal(set) var status: Status
    public internal(set) var startedAt: Date?
    public internal(set) var endedAt: Date?

    public var id: Int { seq }

    public init(
        seq: Int,
        input: String,
        hasInput: Bool = true,
        output: String = "",
        status: Status = .running,
        startedAt: Date? = nil,
        endedAt: Date? = nil
    ) {
        self.seq = seq
        self.input = input
        self.hasInput = hasInput
        self.output = output
        self.status = status
        self.startedAt = startedAt
        self.endedAt = endedAt
    }

    public var inputPreview: String { Self.firstLine(input) }

    public var outputPreview: String { Self.lastLine(output) }

    static func firstLine(_ text: String) -> String {
        text.split(whereSeparator: \.isNewline).first { !$0.allSatisfy(\.isWhitespace) }.map { String($0) } ?? ""
    }

    static func lastLine(_ text: String) -> String {
        text.split(whereSeparator: \.isNewline).last { !$0.allSatisfy(\.isWhitespace) }.map { String($0) } ?? ""
    }
}

public struct AgentSessionTranscript: Equatable, Sendable {
    public enum Segment: Equatable, Sendable {
        case text(String)
        case turn(AgentSessionTurn)
    }

    public static let sentinel: Character = "\u{1E}"
    static let lead = "session-platform "
    static let inputLead = "\u{1E}session-platform input "

    public private(set) var segments: [Segment] = []

    public init(segments: [Segment] = []) {
        self.segments = segments
    }

    public var turns: [AgentSessionTurn] {
        segments.compactMap { segment -> AgentSessionTurn? in
            if case .turn(let turn) = segment { return turn }
            return nil
        }
    }

    public var hasTurns: Bool { !turns.isEmpty }

    public var lastSeq: Int? { turns.last?.seq }

    public static func parse(_ raw: String) -> AgentSessionTranscript {
        var transcript = AgentSessionTranscript()
        guard raw.contains(sentinel) else {
            transcript.appendText(raw)
            return transcript
        }
        let parts = raw.split(separator: sentinel, omittingEmptySubsequences: false)
        transcript.appendText(String(parts[0]))
        var openTurn: Int?
        for part in parts.dropFirst() {
            guard let newline = part.firstIndex(of: "\n") else { continue }
            let header = part[..<newline]
            let body = String(part[part.index(after: newline)...])
            guard header.hasPrefix(lead) else {
                transcript.appendBody(body, to: openTurn)
                continue
            }
            let fields = header.dropFirst(lead.count).split(separator: " ")
            let kind = fields.first.map { String($0) } ?? ""
            var values: [String: String] = [:]
            for field in fields.dropFirst() {
                let pair = field.split(separator: "=", maxSplits: 1)
                if pair.count == 2 {
                    values[String(pair[0])] = String(pair[1])
                }
            }
            guard let seq = Int(values["seq"] ?? "") else {
                transcript.appendBody(body, to: openTurn)
                continue
            }
            switch kind {
            case "input":
                let prompt = body.hasSuffix("\n") ? String(body.dropLast()) : body
                transcript.segments.append(.turn(AgentSessionTurn(seq: seq, input: prompt, startedAt: date(values["at"]))))
                openTurn = nil
            case "output":
                transcript.ensureTurn(seq)
                openTurn = seq
                transcript.appendBody(body, to: seq)
            case "end":
                transcript.ensureTurn(seq)
                transcript.updateTurn(seq) { turn in
                    turn.status = Status(marker: values["status"] ?? "")
                    turn.endedAt = date(values["at"])
                    turn.output = trimmingTrailingNewlines(turn.output)
                }
                openTurn = nil
                transcript.appendText(body)
            default:
                transcript.appendBody(body, to: openTurn)
            }
        }
        return transcript
    }

    public static func keepingWholeTurns(_ tail: String) -> String {
        guard let start = tail.range(of: inputLead)?.lowerBound, start != tail.startIndex else { return tail }
        return String(tail[start...])
    }

    private typealias Status = AgentSessionTurn.Status

    private mutating func appendText(_ text: String) {
        guard !text.isEmpty else { return }
        if case .text(let previous)? = segments.last {
            segments[segments.count - 1] = .text(previous + text)
        } else {
            segments.append(.text(text))
        }
    }

    private mutating func appendBody(_ body: String, to seq: Int?) {
        guard let seq else {
            appendText(body)
            return
        }
        updateTurn(seq) { $0.output += body }
    }

    private mutating func ensureTurn(_ seq: Int) {
        if turnIndex(seq) == nil {
            segments.append(.turn(AgentSessionTurn(seq: seq, input: "", hasInput: false)))
        }
    }

    private mutating func updateTurn(_ seq: Int, _ change: (inout AgentSessionTurn) -> Void) {
        guard let index = turnIndex(seq), case .turn(var turn) = segments[index] else { return }
        change(&turn)
        segments[index] = .turn(turn)
    }

    private func turnIndex(_ seq: Int) -> Int? {
        segments.lastIndex { segment -> Bool in
            if case .turn(let turn) = segment { return turn.seq == seq }
            return false
        }
    }

    private static func date(_ raw: String?) -> Date? {
        guard let raw else { return nil }
        return ISO8601DateFormatter().date(from: raw)
    }

    private static func trimmingTrailingNewlines(_ text: String) -> String {
        var trimmed = Substring(text)
        while trimmed.last?.isNewline == true {
            trimmed = trimmed.dropLast()
        }
        return String(trimmed)
    }
}

public extension AgentSessionFormatting {
    static func turnTiming(_ turn: AgentSessionTurn, timeZone: TimeZone = .current) -> String? {
        guard let started = turn.startedAt else { return nil }
        let clock = DateFormatter()
        clock.locale = Locale(identifier: "ko_KR")
        clock.timeZone = timeZone
        clock.dateFormat = "HH:mm"
        var parts = [clock.string(from: started)]
        if let ended = turn.endedAt {
            parts.append(duration(ended.timeIntervalSince(started)))
        }
        return parts.joined(separator: " · ")
    }

    static func duration(_ interval: TimeInterval) -> String {
        let seconds = max(0, Int(interval.rounded()))
        if seconds < 60 { return "\(seconds)초" }
        if seconds < 3600 {
            let rest = seconds % 60
            return rest == 0 ? "\(seconds / 60)분" : "\(seconds / 60)분 \(rest)초"
        }
        let minutes = (seconds % 3600) / 60
        return minutes == 0 ? "\(seconds / 3600)시간" : "\(seconds / 3600)시간 \(minutes)분"
    }
}
