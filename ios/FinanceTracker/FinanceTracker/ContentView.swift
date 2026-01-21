//
//  ContentView.swift
//  FinanceTracker
//
//  Created by dev on 21/1/26.
//

import SwiftUI

struct ContentView: View {
    @StateObject private var authManager: AuthManager
    @State private var selectedTab = 0
    @State private var isCheckingAuth = true

    init() {
        let mockAPI = APIService(baseURL: Config.baseURL)
        _authManager = StateObject(wrappedValue: AuthManager(apiService: mockAPI))
    }

    var body: some View {
        Group {
            if isCheckingAuth {
                // Show loading while checking authentication status
                loadingView
            } else if authManager.isAuthenticated {
                mainTabView
            } else {
                LoginView(authManager: authManager)
            }
        }
        .onAppear {
            // Small delay to ensure auth check completes
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                isCheckingAuth = false
            }
        }
    }

    private var loadingView: some View {
        VStack(spacing: 20) {
            ProgressView()
                .scaleEffect(1.5)

            Text("Loading...")
                .font(Typography.subheadline)
                .foregroundColor(ColorPalette.textSecondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(ColorPalette.background)
    }

    private var mainTabView: some View {
        TabView(selection: $selectedTab) {
            DashboardView(viewModel: DashboardViewModel(apiService: APIService(baseURL: Config.baseURL)))
                .tabItem {
                    Label("Dashboard", systemImage: selectedTab == 0 ? "chart.bar.fill" : "chart.bar")
                }
                .tag(0)

            AnalyticsView(viewModel: AnalyticsViewModel(apiService: APIService(baseURL: Config.baseURL)))
                .tabItem {
                    Label("Analytics", systemImage: selectedTab == 1 ? "chart.pie.fill" : "chart.pie")
                }
                .tag(1)

            TransactionsView(apiService: APIService(baseURL: Config.baseURL))
                .tabItem {
                    Label("Transactions", systemImage: selectedTab == 2 ? "list.bullet.rectangle.fill" : "list.bullet.rectangle")
                }
                .tag(2)

            SettingsView()
                .tabItem {
                    Label("Settings", systemImage: selectedTab == 3 ? "gearshape.fill" : "gearshape")
                }
                .tag(3)
        }
        .accentColor(ColorPalette.primaryIndigo)
        .environmentObject(authManager)
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
