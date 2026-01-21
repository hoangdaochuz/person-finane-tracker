# iOS App Design - Personal Finance Tracker

**Date:** 2026-01-21
**Status:** Design Approved

## Overview

A native iOS application for tracking financial transactions by reading bank/e-wallet notifications and displaying them in a modern fintech-style dashboard with analytics.

## Key Requirements

| Aspect | Decision |
|--------|----------|
| Data Source | Notification Reading (SMS/Push from banks) |
| Data Strategy | Backend-First (primary data on API, local cache only) |
| MVP Scope | Full Dashboard (capture + analytics visualization) |
| Authentication | User Login (email/password → API key) |
| Design Style | Modern Fintech (gradients, bold colors, card-based UI) |
| Minimum iOS | iOS 16.0 (Swift Charts requirement) |

---

## Architecture & Data Flow

### MVVM Architecture

```
Notification (Background)
    ↓
NotificationManager (UNUserNotificationCenter)
    ↓
TransactionParser (Regex/Pattern matching)
    ↓
APIService (URLSession → Backend)
    ↓
Dashboard ViewModels (Transform data)
    ↓
SwiftUI Views (Display)
```

### Key Components

| Component | Responsibility |
|-----------|---------------|
| `NotificationManager` | UNUserNotificationCenter delegate, handles incoming notifications |
| `TransactionParser` | Parses notification text into transaction models |
| `APIService` | URLSession-based networking layer |
| `AuthManager` | Handles login session and API key storage |
| `CoreData Stack` | Local caching for offline display |

### Security

- API keys stored in iOS Keychain
- HTTPS-only communication with backend
- Biometric (Face ID/Touch ID) for app access

---

## App Structure & Navigation

### Four Main Tabs

| Tab | Content |
|-----|---------|
| **Dashboard** | Summary cards (Balance, Income, Expenses), Recent transactions, Quick add button |
| **Analytics** | Period selector, Income vs Expense bar chart, Category pie chart, Source ranking |
| **Transactions** | Full list with filters (type, source, category, date), Pull-to-refresh, Detail view |
| **Settings** | User profile, Notification status, Connected sources, Theme/Biometric toggles |

### Navigation Flow

```
Onboarding (first launch) → Login Screen → Dashboard
Authenticated users → Direct to Dashboard on launch
Modals: Add Transaction, Transaction Details
```

---

## UI Components & Design System

### Color Palette (Modern Fintech)

```
Primary Gradient: #6366F1 → #8B5CF6 (indigo to violet)
Success (Income):  #10B981 (emerald)
Danger (Expense):  #EF4444 (red)
Background:        #F9FAFB (light gray)
Card Background:   #FFFFFF (white)
Text Primary:      #111827
Text Secondary:    #6B7280
```

### Custom Views

- `GradientButton` - Primary actions with gradient background
- `StatCard` - Metric display with icon, label, value
- `TransactionCell` - List item with amount, merchant, date, category badge
- `ChartCard` - Container for Swift Charts visualizations

### Charts

- **Bar Chart**: Income vs Expense trends (Swift Charts with gradient fills)
- **Donut Chart**: Category breakdown with legend
- **Interactions**: Touch for drill-down into data

### Typography

- SF Pro font family
- Headings: Bold, 20-28pt
- Body: Regular/Medium, 14-17pt
- Currency: Monospaced for alignment

---

## Notification Handling & Parsing

### Processing Pipeline

```
Incoming Notification
    ↓
Filter by Sender (known banks/wallets)
    ↓
Parse Text Content (regex patterns)
    ↓
Extract: Amount, Type, Merchant, Date
    ↓
Send to Backend API
    ↓
Refresh Dashboard
```

### Parser Strategy

- Bank-specific regex patterns in `NotificationPatterns.plist`
- Pattern matching: "Rp 50.000", "Spent $15 at", "Transfer from..."
- Fallback ML-based classification for transaction type
- Confidence scoring - low confidence flagged for review

### Background Execution

- `BGProcessingTask` for batch processing missed notifications
- `UNUserNotificationCenterDelegate` for real-time capture
- Queue failed API requests for retry

### Supported Types (MVP)

- SMS transaction alerts
- Push notifications from banking apps
- Future: Screenshot analysis

### Edge Cases

- Partial/ambiguous data → Save as "pending" for user completion
- Duplicate detection (same amount + merchant + time window)
- Unsupported formats → "Learned from new bank" prompt

---

## Networking & Backend Integration

### API Endpoints

```swift
enum APIEndpoint {
    case login(email: String, password: String)
    case register(email: String, password: String)
    case createTransaction(transaction: Transaction)
    case getAnalytics(period: TimePeriod)
    case getTransactions(page: Int, limit: Int)
    case getSummary
}
```

### API Key Management

1. After login, store `apiKey` in Keychain (`kSecAttrAccount` = "finance_tracker_api_key")
2. Add `X-API-Key` header to all authenticated requests
3. Handle token expiration if implemented

### Error Handling

| Error Type | Action |
|------------|--------|
| Network error | Retry with exponential backoff |
| 401 Unauthorized | Redirect to login |
| 4xx/5xx | User-friendly error message |
| Timeout | Queue for background retry |

### Offline Support

- Queue outgoing transactions in CoreData
- Background `URLSession` for sync on reconnect
- Show "Last synced: [timestamp]" indicator

---

## Testing Strategy

### Unit Tests

- `TransactionParserTests` - Regex pattern verification
- `APIServiceTests` - Mock URLSession for all endpoints
- `ViewModelTests` - State management and data transformation

### Integration Tests

- End-to-end notification flow
- Login flow (credentials → API key → authenticated requests)
- CoreData persistence

### UI Tests (minimal)

- Onboarding flow completion
- Login form submission
- Tab navigation

### Test Doubles

- `MockURLSession` for API tests
- `MockNotificationManager` for notification tests
- In-memory CoreData stack

---

## Project Structure

```
FinanceTracker/
├── FinanceTracker/
│   ├── App/
│   │   ├── FinanceTrackerApp.swift
│   │   ├── ContentView.swift
│   │   └── AppDelegate.swift
│   ├── Models/
│   │   ├── Transaction.swift
│   │   ├── Analytics.swift
│   │   └── User.swift
│   ├── ViewModels/
│   │   ├── DashboardViewModel.swift
│   │   ├── AnalyticsViewModel.swift
│   │   ├── TransactionsViewModel.swift
│   │   └── AuthViewModel.swift
│   ├── Views/
│   │   ├── Dashboard/
│   │   ├── Analytics/
│   │   ├── Transactions/
│   │   ├── Auth/
│   │   ├── Settings/
│   │   └── Components/
│   ├── Services/
│   │   ├── APIService.swift
│   │   ├── NotificationManager.swift
│   │   ├── TransactionParser.swift
│   │   ├── AuthManager.swift
│   │   └── KeychainManager.swift
│   ├── Persistence/
│   │   ├── CoreDataStack.swift
│   │   └── Models/
│   └── Resources/
│       ├── Assets.xcassets
│       ├── NotificationPatterns.plist
│       └── ColorPalette.swift
└── FinanceTrackerTests/
    ├── UnitTests/
    └── IntegrationTests/
```

### Dependencies

- No external packages for MVP
- Swift Charts (built-in iOS 16+)
- Future: Consider Alamofire for advanced networking

---

## Next Steps

1. Set up Xcode project with iOS 16 target
2. Implement core services (APIService, NotificationManager, TransactionParser)
3. Build authentication flow (Onboarding → Login → Dashboard)
4. Implement tab navigation and main views
5. Add notification parsing and background handling
6. Integrate analytics charts
7. Add unit and integration tests
8. Polish UI with animations and micro-interactions
