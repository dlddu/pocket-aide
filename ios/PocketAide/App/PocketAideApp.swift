// PocketAideApp.swift
// PocketAide

import SwiftUI

@main
struct PocketAideApp: App {

    /// UI 테스트 실행 중 여부. `--uitesting` 인자가 있으면 `true`.
    private let isUITesting: Bool = CommandLine.arguments.contains("--uitesting")

    var body: some Scene {
        WindowGroup {
            if isUITesting {
                // 인증/온보딩을 건너뛰고 TabBar로 바로 이동
                MainTabView()
            } else {
                MainTabView()
            }
        }
    }
}
