import Foundation
import PocketAideAPI

@MainActor
final class WeatherViewModel: ObservableObject {
    enum State: Equatable {
        case loading
        case needsLocation
        case error
        case loaded(WeatherForecastSnapshot)
    }

    @Published private(set) var state: State = .loading
    @Published private(set) var placeName: String?
    private var shownLocation: WeatherCoordinate?

    func load(useCache: Bool) async {
        guard let location = WeatherLocationStore().load() else {
            shownLocation = nil
            placeName = nil
            state = .needsLocation
            return
        }
        shownLocation = location
        placeName = location.placeName
        if useCache, let cached = WeatherForecastCache().load(for: location, now: Date()) {
            state = .loaded(cached)
        }
        do {
            state = .loaded(try await WeatherClient.fetchForecast(location))
        } catch is CancellationError {
            return
        } catch let error as URLError where error.code == .cancelled {
            return
        } catch {
            state = .error
        }
    }

    func locationDidRefresh() async {
        guard WeatherLocationStore().load() != shownLocation else { return }
        await load(useCache: true)
    }
}
