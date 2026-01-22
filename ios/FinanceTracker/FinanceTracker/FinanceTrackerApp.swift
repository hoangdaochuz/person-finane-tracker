//
//  FinanceTrackerApp.swift
//  FinanceTracker
//
//  Created by dev on 21/1/26.
//

import SwiftUI
import CoreData

@main
struct FinanceTrackerApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    let persistenceController = CoreDataStack.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(\.managedObjectContext, persistenceController.viewContext)
        }
    }
}
