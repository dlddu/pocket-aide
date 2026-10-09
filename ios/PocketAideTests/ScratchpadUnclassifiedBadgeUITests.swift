// 검증 시나리오: test-scratchpad.md#시나리오 7
import UIKit
import Vision
import XCTest

final class ScratchpadUnclassifiedBadgeUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
        UITestAuth.ensureSignedIn(self)
    }

    func testHeaderCountAndTabBadgeFollowAddMoveDeleteAndRelaunch() {
        let token = TodoUI.uniqueToken()
        let kept = "남겨 둘 메모 \(token)"
        let deleted = "지울 메모 \(token)"
        let moved = "개인으로 옮길 메모 \(token)"
        let app = TodoUI.launch()

        var scratchpad = ScratchpadScreen.open(in: app)
        scratchpad.add(kept)
        scratchpad.add(deleted)
        let counts = UnclassifiedCounts(app: app)
        guard let base = counts.header else {
            return XCTFail("The 분류되지 않은 메모 count should be readable in the 임시공간 header")
        }
        XCTAssertGreaterThanOrEqual(base, 2, "The header should count at least the two memos just added")
        counts.assertBoth(base, "With the two memos added")

        scratchpad.add(moved)
        counts.assertBoth(base + 1, "Right after adding a memo")

        scratchpad.move(moved, to: "personal")
        XCTAssertTrue(TodoUI.waitToDisappear(scratchpad.memo(moved)), "The memo moved to 개인 should leave 임시공간")
        counts.assertBoth(base, "Right after moving a memo to 개인")

        scratchpad.swipeDelete(deleted)
        counts.assertBoth(base - 1, "Right after deleting a memo")

        TodoUI.relaunch(app)
        scratchpad = ScratchpadScreen.open(in: app)
        XCTAssertTrue(scratchpad.memo(kept).waitForExistence(timeout: 15), "The remaining memo should still be listed")
        counts.assertBoth(base - 1, "After relaunching the app")

        scratchpad.remove([kept])
    }
}

private struct UnclassifiedCounts {
    let app: XCUIApplication

    var header: Int? {
        app.staticTexts
            .matching(NSPredicate(format: "identifier IN %@ AND label MATCHES %@", ["scratchpad.badge", "scratchpad.screen"], "[0-9]+"))
            .allElementsBoundByIndex
            .min { $0.frame.minY < $1.frame.minY }
            .flatMap { Int($0.label) }
    }

    var tab: XCUIElement {
        app.tabBars.firstMatch.buttons
            .matching(NSPredicate(format: "label BEGINSWITH %@", ScratchpadScreen.tabLabel))
            .firstMatch
    }

    func assertBoth(_ expected: Int, _ step: String, file: StaticString = #filePath, line: UInt = #line) {
        var reading = TabBadgeReading.unreadable
        let matched = RoutinesScreen.eventually(timeout: 15) {
            reading = TabBadgeReader.read(tab, in: app)
            return header == expected && reading.count == expected
        }
        if !matched {
            let shot = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
            shot.lifetime = .keepAlways
            let badge = XCTAttachment(string: "\(reading)")
            badge.lifetime = .keepAlways
            XCTContext.runActivity(named: "Counts disagreed: \(step)") {
                $0.add(shot)
                $0.add(badge)
            }
        }
        XCTAssertTrue(
            matched,
            "\(step): the header count and the 임시공간 tab badge should both read \(expected) but read \(header.map(String.init) ?? "-") / \(reading)",
            file: file,
            line: line
        )
    }
}

enum TabBadgeReading: CustomStringConvertible {
    case none
    case number(Int)
    case unreadable

    var count: Int? {
        switch self {
        case .none: return 0
        case .number(let value): return value
        case .unreadable: return nil
        }
    }

    var description: String {
        switch self {
        case .none: return "no badge"
        case .number(let value): return "badge \(value)"
        case .unreadable: return "badge unreadable"
        }
    }
}

enum TabBadgeReader {
    static func read(_ tab: XCUIElement, in app: XCUIApplication) -> TabBadgeReading {
        guard tab.exists, app.frame.width > 0, let screen = XCUIScreen.main.screenshot().image.cgImage else {
            return .unreadable
        }
        let scale = CGFloat(screen.width) / app.frame.width
        let area = tab.frame.insetBy(dx: -6, dy: -10)
        let pixelArea = CGRect(x: area.minX * scale, y: area.minY * scale, width: area.width * scale, height: area.height * scale)
            .integral
            .intersection(CGRect(x: 0, y: 0, width: screen.width, height: screen.height))
        guard !pixelArea.isEmpty, let crop = screen.cropping(to: pixelArea), let bitmap = Bitmap(crop) else {
            return .unreadable
        }
        let spans = bitmap.redSpans()
        guard spans.reduce(0, { $0 + $1.red }) >= Int(scale * scale * 20) else {
            return .none
        }
        guard let glyphs = GlyphMask(bitmap.glyphPoints(in: spans)) else {
            return .unreadable
        }
        return (matchShape(glyphs) ?? recognizeText(glyphs)).map(TabBadgeReading.number) ?? .unreadable
    }

    private static let templates: [Int: GlyphMask] = Dictionary(
        uniqueKeysWithValues: (1...99).compactMap { value in TabBadgeReader.rendered(String(value)).map { (value, $0) } }
    )

    private static func rendered(_ text: String) -> GlyphMask? {
        let attributes: [NSAttributedString.Key: Any] = [
            .font: UIFont.systemFont(ofSize: 64, weight: .medium),
            .foregroundColor: UIColor.black,
        ]
        let size = (text as NSString).size(withAttributes: attributes)
        let canvas = CGSize(width: ceil(size.width) + 8, height: ceil(size.height) + 8)
        let format = UIGraphicsImageRendererFormat()
        format.scale = 1
        format.opaque = true
        let image = UIGraphicsImageRenderer(size: canvas, format: format).image { context in
            UIColor.white.setFill()
            context.fill(CGRect(origin: .zero, size: canvas))
            (text as NSString).draw(at: CGPoint(x: 4, y: 4), withAttributes: attributes)
        }
        guard let cgImage = image.cgImage, let bitmap = Bitmap(cgImage) else { return nil }
        return GlyphMask(bitmap.darkPoints())
    }

    private static func matchShape(_ glyphs: GlyphMask) -> Int? {
        let scores = templates
            .map { (value: $0.key, score: glyphs.similarity(to: $0.value)) }
            .sorted { $0.score > $1.score }
        guard let best = scores.first else { return nil }
        let runnerUp = scores.dropFirst().first?.score ?? 0
        return best.score >= 0.5 && best.score - runnerUp >= 0.15 ? best.value : nil
    }

    private static func recognizeText(_ glyphs: GlyphMask) -> Int? {
        guard let image = glyphs.image() else { return nil }
        for level in [VNRequestTextRecognitionLevel.accurate, .fast] {
            let request = VNRecognizeTextRequest()
            request.recognitionLevel = level
            request.usesLanguageCorrection = false
            request.recognitionLanguages = ["en-US"]
            try? VNImageRequestHandler(cgImage: image, options: [:]).perform([request])
            let text = (request.results ?? []).compactMap { $0.topCandidates(1).first?.string }.joined()
            if let value = Int(String(text.compactMap(digit))) {
                return value
            }
        }
        return nil
    }

    private static let lookalikes: [Character: Character] = [
        "I": "1", "l": "1", "|": "1", "i": "1", "!": "1",
        "O": "0", "o": "0", "D": "0", "Q": "0",
        "S": "5", "s": "5", "Z": "2", "z": "2", "B": "8", "g": "9", "q": "9",
    ]

    private static func digit(_ character: Character) -> Character? {
        character.isASCII && character.isNumber ? character : lookalikes[character]
    }
}

private struct Span {
    let y: Int
    let left: Int
    let right: Int
    let red: Int
}

private struct Point {
    let x: Int
    let y: Int
}

private struct GlyphMask {
    let width: Int
    let height: Int
    let ink: [Bool]

    init?(_ points: [Point]) {
        guard let minX = points.map(\.x).min(), let maxX = points.map(\.x).max(),
              let minY = points.map(\.y).min(), let maxY = points.map(\.y).max() else { return nil }
        let w = maxX - minX + 1
        let h = maxY - minY + 1
        var cells = [Bool](repeating: false, count: w * h)
        for point in points {
            cells[(point.y - minY) * w + point.x - minX] = true
        }
        width = w
        height = h
        ink = cells
    }

    func similarity(to other: GlyphMask) -> Double {
        var overlap = 0
        var union = 0
        for y in 0..<height {
            for x in 0..<width {
                let mine = ink[y * width + x]
                let theirs = other.ink[(y * other.height / height) * other.width + x * other.width / width]
                if mine && theirs { overlap += 1 }
                if mine || theirs { union += 1 }
            }
        }
        return union == 0 ? 0 : Double(overlap) / Double(union)
    }

    func image(magnify: Int = 4, margin: Int = 6) -> CGImage? {
        let w = (width + margin * 2) * magnify
        let h = (height + margin * 2) * magnify
        var out = [UInt8](repeating: 255, count: w * h)
        for y in 0..<height {
            for x in 0..<width where ink[y * width + x] {
                let ox = (x + margin) * magnify
                for row in ((y + margin) * magnify)..<((y + margin + 1) * magnify) {
                    out.replaceSubrange((row * w + ox)..<(row * w + ox + magnify), with: repeatElement(0, count: magnify))
                }
            }
        }
        guard let provider = CGDataProvider(data: Data(out) as CFData) else { return nil }
        return CGImage(
            width: w,
            height: h,
            bitsPerComponent: 8,
            bitsPerPixel: 8,
            bytesPerRow: w,
            space: CGColorSpaceCreateDeviceGray(),
            bitmapInfo: CGBitmapInfo(rawValue: CGImageAlphaInfo.none.rawValue),
            provider: provider,
            decode: nil,
            shouldInterpolate: false,
            intent: .defaultIntent
        )
    }
}

private struct Bitmap {
    let width: Int
    let height: Int
    let bytes: [UInt8]

    init?(_ image: CGImage) {
        let w = image.width
        let h = image.height
        var buffer = [UInt8](repeating: 0, count: w * h * 4)
        let drawn = buffer.withUnsafeMutableBytes { raw -> Bool in
            guard let context = CGContext(
                data: raw.baseAddress,
                width: w,
                height: h,
                bitsPerComponent: 8,
                bytesPerRow: w * 4,
                space: CGColorSpaceCreateDeviceRGB(),
                bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
            ) else { return false }
            context.draw(image, in: CGRect(x: 0, y: 0, width: w, height: h))
            return true
        }
        guard drawn else { return nil }
        width = w
        height = h
        bytes = buffer
    }

    private func channel(_ x: Int, _ y: Int, _ offset: Int) -> Int {
        Int(bytes[(y * width + x) * 4 + offset])
    }

    private func isBadgeRed(_ x: Int, _ y: Int) -> Bool {
        channel(x, y, 0) >= 200 && channel(x, y, 1) <= 110 && channel(x, y, 2) <= 110
    }

    private func isGlyph(_ x: Int, _ y: Int) -> Bool {
        channel(x, y, 0) >= 200 && channel(x, y, 1) >= 170 && channel(x, y, 2) >= 170
    }

    func redSpans() -> [Span] {
        var spans: [Span] = []
        for y in 0..<height {
            let red = (0..<width).filter { isBadgeRed($0, y) }
            if let left = red.first, let right = red.last {
                spans.append(Span(y: y, left: left, right: right, red: red.count))
            }
        }
        return spans
    }

    func glyphPoints(in spans: [Span]) -> [Point] {
        spans.flatMap { span in
            (span.left...span.right).filter { isGlyph($0, span.y) }.map { Point(x: $0, y: span.y) }
        }
    }

    func darkPoints() -> [Point] {
        (0..<height).flatMap { y in
            (0..<width).filter { channel($0, y, 0) < 128 }.map { Point(x: $0, y: y) }
        }
    }
}
