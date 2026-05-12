import AVFoundation
import Foundation

final class TTSPlayer: ObservableObject {
    private let synthesizer = AVSpeechSynthesizer()

    func speak(_ text: String, languageCode: String = "ko-KR") {
        let utterance = AVSpeechUtterance(string: text)
        utterance.voice = AVSpeechSynthesisVoice(language: languageCode)
        utterance.rate = AVSpeechUtteranceDefaultSpeechRate
        synthesizer.stopSpeaking(at: .immediate)
        synthesizer.speak(utterance)
    }

    func stop() {
        synthesizer.stopSpeaking(at: .immediate)
    }
}
