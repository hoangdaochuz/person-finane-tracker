# Quick Start: Finance Tracker iOS App in Xcode

## 🚀 5-Minute Setup

### 1. Create Project (1 min)
```
Xcode → File → New → Project
Template: iOS → App
Product Name: FinanceTracker
Interface: SwiftUI ✓
Language: Swift ✓
```

### 2. Add Swift Files (2 min)
```
Right-click "FinanceTracker" folder → Add Files...
Navigate to: ios/FinanceTracker/FinanceTracker/
Select ALL folders (Cmd+A)
☑ Copy items if needed
☑ Create groups (NOT folder refs)
Add
```

### 3. Add Core Data Model (30 sec)
```
Right-click Persistence folder → Add Files...
Navigate to: FinanceTracker.xcdatamodeld
☑ Create folder references (IMPORTANT!)
Add
```

### 4. Configure Settings (1 min)
```
Project → General → Deployment Target
iOS: 16.0 ✓

Project → Build Settings → Search Paths
Add: $(SRCROOT)/FinanceTracker/**  (recursive)
```

### 5. Run! (30 sec)
```
Simulator: iPhone 15
Press ▶️ or Cmd+R
```

---

## 📁 Folders to Add

```
FinanceTracker/
├── App/                          ✓ Required
├── Core/                         ✓ Required
│   ├── Network/                  ✓
│   ├── Keychain/                 ✓
│   ├── Security/                 ✓
│   ├── Persistence/              ✓
│   │   └── FinanceTracker.xcdatamodeld/  ✓ (as folder reference)
│   ├── Notification/             ✓
│   └── Export/                   ✓
├── Data/                         ✓ Required
│   ├── Models/                   ✓
│   ├── Repositories/             ✓
│   └── DataSources/              ✓
├── Domain/                       ✓ Required
│   ├── UseCases/                 ✓
│   └── Services/                 ✓
├── Presentation/                 ✓ Required
│   ├── Common/                   ✓
│   ├── Onboarding/               ✓
│   ├── Dashboard/                ✓
│   ├── Transactions/             ✓
│   ├── Analytics/                ✓
│   ├── Goals/                    ✓
│   ├── Settings/                 ✓
│   └── Export/                   ✓
└── Resources/                    ✓ Optional
    └── Assets.xcassets/          ✓
```

---

## ❌ Skip These for Now

- `Widgets/` (needs separate extension target)
- `Tests/` (we'll add later)

---

## 🔧 Common Build Errors

| Error | Fix |
|-------|-----|
| **Cannot find 'DesignSystem'** | Clean Build (Shift+Cmd+K), check target membership |
| **Use of unresolved 'Color'** | Add `import SwiftUI` |
| **Missing CoreData** | Add `import CoreData` |
| **Module not found** | Add to Target Membership (File Inspector → Target Membership) |

---

## ✅ Success Check

When running, you should see:

1. **Welcome Screen**:
   ```
   Finance Tracker logo
   "Welcome to Finance Tracker"
   "Get Started" button
   ```

2. **API Key Screen** (after clicking Get Started):
   ```
   "Connect Your App"
   [API Key input field]
   "Continue" button
   ```

3. **Dashboard** (after setup):
   ```
   Balance card with gradient
   Quick action buttons
   "No Transactions" empty state
   ```

---

## 💡 Pro Tips

1. **Clean Build** after adding files: `Product → Clean Build Folder`
2. **Check Target Membership** for each file added
3. **Select iPhone 15 simulator** (iOS 17.0+)
4. **Press Cmd+.** to open code in focus when clicking an error

---

## 🎯 Next Steps After First Run

1. Test **Onboarding flow**
2. Add a **test transaction**
3. Configure **backend API**
4. Test **biometric authentication**
