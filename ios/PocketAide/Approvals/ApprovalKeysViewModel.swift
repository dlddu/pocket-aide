import Foundation
import PocketAideAPI

@MainActor
final class ApprovalKeysViewModel: ObservableObject {
    @Published private(set) var keys: [ApprovalCallerKey] = []
    @Published private(set) var isLoading = false
    @Published private(set) var hasLoaded = false
    @Published private(set) var isIssuing = false
    @Published private(set) var issued: IssuedApprovalKey?
    @Published private(set) var isRegistering = false
    @Published private(set) var registerOutcome: ApprovalServiceAccountOutcome?
    @Published private(set) var autoResponse: ApprovalAutoResponse?
    @Published var modeNotice: String?
    @Published var errorMessage: String?
    @Published var issueNotice: String?
    @Published var registerNotice: String?

    private let api: APIClient?

    init(api: APIClient?) {
        self.api = api
    }

    func load() async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            keys = try await api.listApprovalKeys()
            autoResponse = try await api.approvalSettings().autoResponse
            errorMessage = nil
            hasLoaded = true
        } catch {
            errorMessage = String(describing: error)
        }
    }

    func issue(name raw: String) async {
        guard let api, !isIssuing, let name = ApprovalKeyFormatting.trimmedName(raw) else { return }
        isIssuing = true
        defer { isIssuing = false }
        do {
            switch try await api.issueApprovalKey(name: name) {
            case .issued(let key):
                issued = key
                issueNotice = nil
                await load()
            case .duplicateName:
                issueNotice = ApprovalKeyFormatting.duplicateNameNotice
            }
        } catch {
            issueNotice = String(describing: error)
        }
    }

    @discardableResult
    func registerServiceAccount(name rawName: String, namespace: String, account: String) async -> Bool {
        guard let api, !isRegistering, let name = ApprovalKeyFormatting.trimmedName(rawName) else { return false }
        guard let serviceAccount = ApprovalKeyFormatting.serviceAccount(namespace: namespace, name: account) else {
            registerOutcome = .invalidServiceAccount
            registerNotice = ApprovalKeyFormatting.invalidServiceAccountNotice
            return false
        }
        isRegistering = true
        defer { isRegistering = false }
        do {
            let outcome = try await api.registerApprovalServiceAccount(name: name, serviceAccount: serviceAccount)
            registerOutcome = outcome
            switch outcome {
            case .registered:
                registerNotice = nil
                await load()
                return true
            case .duplicateName:
                registerNotice = ApprovalKeyFormatting.serviceAccountDuplicateNameNotice
            case .duplicateServiceAccount:
                registerNotice = ApprovalKeyFormatting.duplicateServiceAccountNotice
            case .invalidServiceAccount:
                registerNotice = ApprovalKeyFormatting.invalidServiceAccountNotice
            }
        } catch {
            registerOutcome = nil
            registerNotice = String(describing: error)
        }
        return false
    }

    func setAutoResponse(_ mode: ApprovalAutoResponse) async {
        guard let api, mode != autoResponse else { return }
        do {
            autoResponse = try await api.updateApprovalAutoResponse(mode).autoResponse
            modeNotice = nil
        } catch {
            modeNotice = String(describing: error)
        }
    }

    func revoke(_ key: ApprovalCallerKey) async {
        guard let api else { return }
        do {
            _ = try await api.revokeApprovalKey(id: key.id)
        } catch {
            errorMessage = String(describing: error)
        }
        await load()
    }

    func finishIssuing() {
        issued = nil
        issueNotice = nil
    }

    func finishRegistering() {
        registerOutcome = nil
        registerNotice = nil
    }
}
