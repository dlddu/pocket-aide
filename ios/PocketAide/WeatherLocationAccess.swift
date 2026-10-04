import CoreLocation
import PocketAideAPI

@MainActor
enum WeatherLocationAccess {
    private static var running = false

    static func start() {
        guard !running else { return }
        running = true
        Task {
            await refresh()
            running = false
        }
    }

    nonisolated private static func refresh() async {
        let store = WeatherLocationStore()
        let session = CLServiceSession(authorization: .whenInUse)
        defer { session.invalidate() }
        let location = await withTaskGroup(of: CLLocation?.self) { group in
            group.addTask { await firstLocation(store: store) }
            group.addTask {
                try? await Task.sleep(for: .seconds(20))
                return nil
            }
            let first = await group.next() ?? nil
            group.cancelAll()
            return first
        }
        guard let location else { return }
        let place = try? await CLGeocoder().reverseGeocodeLocation(location).first?.locality
        let coordinate = WeatherCoordinate(
            latitude: location.coordinate.latitude,
            longitude: location.coordinate.longitude,
            placeName: place
        )
        guard coordinate != store.load() else { return }
        store.save(coordinate)
        WidgetRefresher.reloadAll()
    }

    nonisolated private static func firstLocation(store: WeatherLocationStore) async -> CLLocation? {
        do {
            for try await update in CLLocationUpdate.liveUpdates() {
                if update.authorizationDenied || update.authorizationDeniedGlobally {
                    if store.load() != nil {
                        store.clear()
                        WidgetRefresher.reloadAll()
                    }
                    return nil
                }
                if let location = update.location {
                    return location
                }
            }
        } catch {
            return nil
        }
        return nil
    }
}
