// WeatherView.swift
// PocketAide

import SwiftUI

struct WeatherView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Weather")
                    .font(.largeTitle)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Weather")
        }
        .accessibilityIdentifier("tab_weather_view")
    }
}
