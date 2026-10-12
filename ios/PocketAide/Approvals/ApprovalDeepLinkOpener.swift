import PocketAideAPI
import SwiftUI

struct ApprovalDeepLinkOpener: ViewModifier {
    @Binding var openedApprovalID: String?
    @Binding var path: [ApprovalRequest]
    @Binding var showsSettings: Bool
    @Binding var showsHistory: Bool
    let api: APIClient?

    func body(content: Content) -> some View {
        content.onChange(of: openedApprovalID, initial: true) { _, id in
            guard let id else { return }
            openedApprovalID = nil
            Task { await open(id) }
        }
    }

    private func open(_ id: String) async {
        guard let api, let item = try? await api.approval(id: id) else { return }
        showsSettings = false
        showsHistory = false
        path = [item]
    }
}
