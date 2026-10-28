import Foundation

public enum ApprovalNotificationAction {
    public static let categoryIdentifier = "APPROVAL_REQUEST"
    public static let approveIdentifier = "APPROVAL_APPROVE"
    public static let rejectIdentifier = "APPROVAL_REJECT"

    public static func title(_ decision: ApprovalDecision) -> String {
        decision == .approve ? "승인" : "거절"
    }

    public static func decision(forActionIdentifier identifier: String) -> ApprovalDecision? {
        switch identifier {
        case approveIdentifier: return .approve
        case rejectIdentifier: return .reject
        default: return nil
        }
    }

    public static func resultNotice(_ result: Result<ApprovalDecisionOutcome, Error>) -> String {
        switch result {
        case .success(.decided(let item)):
            return "\(ApprovalFormatting.statusLabel(item.status)) — \(ApprovalFormatting.summary(of: item.context))"
        case .success(.conflict(let item)):
            return ApprovalFormatting.conflictNotice(item)
        case .failure:
            return ApprovalFormatting.failedDecisionNotice
        }
    }
}
