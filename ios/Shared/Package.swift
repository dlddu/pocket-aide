// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "PocketAideShared",
    defaultLocalization: "en",
    platforms: [.iOS(.v26)],
    products: [
        .library(name: "DesignSystem", targets: ["DesignSystem"]),
        .library(name: "PocketAideAPI", targets: ["PocketAideAPI"]),
        .library(name: "PocketAideAuth", targets: ["PocketAideAuth"]),
        .library(name: "PocketAideStorage", targets: ["PocketAideStorage"]),
    ],
    targets: [
        .target(
            name: "DesignSystem",
            path: "Sources/DesignSystem",
            resources: [.process("Resources")]
        ),
        .target(
            name: "PocketAideStorage",
            path: "Sources/PocketAideStorage"
        ),
        .target(
            name: "PocketAideAPI",
            dependencies: ["PocketAideStorage"],
            path: "Sources/PocketAideAPI"
        ),
        .target(
            name: "PocketAideAuth",
            dependencies: ["PocketAideAPI", "PocketAideStorage"],
            path: "Sources/PocketAideAuth"
        ),
    ]
)
