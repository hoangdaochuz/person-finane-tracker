import Foundation

enum Config {
    #if DEBUG
    static let baseURL = "http://localhost:8080"
    #else
    static let baseURL = "https://api.financetracker.app"
    #endif
}