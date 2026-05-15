import AVFoundation
import Foundation

@MainActor
final class TTSPlayer: ObservableObject {
    private let synthesizer = AVSpeechSynthesizer()

    func speak(_ text: String, language: String = "ko-KR") {
        if synthesizer.isSpeaking {
            synthesizer.stopSpeaking(at: .immediate)
        }
        let utterance = AVSpeechUtterance(string: text)
        utterance.voice = AVSpeechSynthesisVoice(language: language)
        utterance.rate = AVSpeechUtteranceDefaultSpeechRate * 0.95
        synthesizer.speak(utterance)
    }
}
