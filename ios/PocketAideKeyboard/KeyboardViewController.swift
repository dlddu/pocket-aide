import UIKit

public final class KeyboardViewController: UIInputViewController {
    private let helloButton: UIButton = {
        var cfg = UIButton.Configuration.filled()
        cfg.title = "Insert hello from pocket-aide"
        cfg.cornerStyle = .large
        return UIButton(configuration: cfg)
    }()

    private let nextKeyboardButton: UIButton = {
        var cfg = UIButton.Configuration.tinted()
        cfg.title = "next ⇄"
        return UIButton(configuration: cfg)
    }()

    public override func viewDidLoad() {
        super.viewDidLoad()
        view.backgroundColor = .systemGroupedBackground

        helloButton.translatesAutoresizingMaskIntoConstraints = false
        nextKeyboardButton.translatesAutoresizingMaskIntoConstraints = false

        helloButton.addAction(UIAction { [weak self] _ in
            self?.textDocumentProxy.insertText("Hello from pocket-aide")
        }, for: .touchUpInside)

        nextKeyboardButton.addAction(UIAction { [weak self] _ in
            self?.advanceToNextInputMode()
        }, for: .touchUpInside)

        view.addSubview(helloButton)
        view.addSubview(nextKeyboardButton)

        NSLayoutConstraint.activate([
            helloButton.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            helloButton.topAnchor.constraint(equalTo: view.safeAreaLayoutGuide.topAnchor, constant: 16),
            nextKeyboardButton.leadingAnchor.constraint(equalTo: view.leadingAnchor, constant: 16),
            nextKeyboardButton.bottomAnchor.constraint(equalTo: view.safeAreaLayoutGuide.bottomAnchor, constant: -16),
        ])
    }
}
