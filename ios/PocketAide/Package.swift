// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "PocketAide",
    platforms: [
        .macOS(.v14),
        .iOS(.v17),
    ],
    products: [
        .library(
            name: "PocketAide",
            targets: ["PocketAide"]
        ),
    ],
    targets: [
        .target(
            name: "PocketAide",
            path: "Sources/PocketAide",
            swiftSettings: [
                .enableUpcomingFeature("BareSlashRegexLiterals"),
            ]
        ),
        .testTarget(
            name: "PocketAideTests",
            dependencies: ["PocketAide"],
            path: "Tests/PocketAideTests"
        ),
    ]
)
