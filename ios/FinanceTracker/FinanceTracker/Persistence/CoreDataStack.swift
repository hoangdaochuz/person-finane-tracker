import CoreData
import Foundation

enum CoreDataError: LocalizedError {
    case storeLoadingFailed(Error)
    case modelNotFound(String)
    case saveFailed(Error)

    var errorDescription: String? {
        switch self {
        case .storeLoadingFailed(let error):
            return "Failed to load Core Data store: \(error.localizedDescription)"
        case .modelNotFound(let name):
            return "Core Data model '\(name)' not found"
        case .saveFailed(let error):
            return "Failed to save Core Data context: \(error.localizedDescription)"
        }
    }
}

class CoreDataStack {
    static let shared = CoreDataStack()

    private(set) var container: NSPersistentContainer?
    private var initializationError: CoreDataError?

    var viewContext: NSManagedObjectContext {
        guard let context = container?.viewContext else {
            // Return a temporary context for previews/testing when store isn't loaded
            let context = NSManagedObjectContext(concurrencyType: .mainQueueConcurrencyType)
            return context
        }
        return context
    }

    var isInitialized: Bool {
        container != nil
    }

    private init() {
        loadPersistentStore()
    }

    private func loadPersistentStore() {
        guard let modelURL = Bundle.main.url(forResource: "FinanceTracker", withExtension: "momd") ??
                              Bundle.main.url(forResource: "FinanceTracker", withExtension: "mom") else {
            initializationError = .modelNotFound("FinanceTracker")
            print("Core Data model not found")
            return
        }

        guard let managedObjectModel = NSManagedObjectModel(contentsOf: modelURL) else {
            initializationError = .modelNotFound("FinanceTracker")
            print("Failed to create NSManagedObjectModel")
            return
        }

        let container = NSPersistentContainer(name: "FinanceTracker", managedObjectModel: managedObjectModel)

        container.loadPersistentStores { _, error in
            if let error = error {
                self.initializationError = .storeLoadingFailed(error)
                print("Core Data store failed to load: \(error)")
                // Graceful degradation - app continues without persistence
            } else {
                self.container = container
            }
        }
    }

    func save() throws {
        guard let container else {
            throw CoreDataError.storeLoadingFailed(
                NSError(domain: "CoreDataStack", code: -1, userInfo: [NSLocalizedDescriptionKey: "Container not initialized"])
            )
        }

        let context = container.viewContext
        if context.hasChanges {
            do {
                try context.save()
            } catch {
                throw CoreDataError.saveFailed(error)
            }
        }
    }

    func saveSilently() {
        do {
            try save()
        } catch {
            print("Failed to save Core Data silently: \(error)")
        }
    }
}
