import AppIntents

public struct ShowHelloIntent: AppIntent {
    public static var title: LocalizedStringResource = "Show pocket-aide hello"
    public static var description = IntentDescription("Returns a greeting from pocket-aide.")
    public static var openAppWhenRun: Bool = false

    public init() {}

    public func perform() async throws -> some IntentResult & ReturnsValue<String> {
        .result(value: "Hello from pocket-aide!")
    }
}

public struct PocketAideShortcuts: AppShortcutsProvider {
    public static var appShortcuts: [AppShortcut] {
        AppShortcut(
            intent: ShowHelloIntent(),
            phrases: [
                "Show hello from \(.applicationName)",
                "Say hi with \(.applicationName)",
            ],
            shortTitle: "Show hello",
            systemImageName: "hand.wave"
        )
    }
}
