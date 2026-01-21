import Foundation

struct User: Codable, Equatable {
    let id: UUID
    var email: String
    var name: String?
    var apiKey: String?
    var isBiometricEnabled: Bool

    init(id: UUID = UUID(), email: String, name: String? = nil, apiKey: String? = nil, isBiometricEnabled: Bool = false) {
        self.id = id
        self.email = email
        self.name = name
        self.apiKey = apiKey
        self.isBiometricEnabled = isBiometricEnabled
    }
}
