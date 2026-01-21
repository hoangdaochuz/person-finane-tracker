# How to Set Up the iOS Finance Tracker App in Xcode

## Step 1: Create a New Xcode Project

1. Open **Xcode** (Install from Mac App Store if needed)
2. Click **"Create a new Xcode project"**
3. Choose **iOS → App** template
4. Product Name: `FinanceTracker`
5. Team: Select your development team
6. Organization Identifier: `com.dev` (or your own)
7. Bundle Identifier will auto-generate: `com.dev.FinanceTracker`
8. Interface: **SwiftUI**
9. Language: **Swift**
10. Storage: **None** (we use Core Data programmatically)
11. Uncheck "Include Tests" for now (we can add them later)
12. Save to: Navigate to your project root and create inside a new `ios/FinanceTracker` folder

## Step 2: Add the Swift Files to the Project

### Option A: Using Finder (Easy Way)

1. In Xcode, right-click on the **"FinanceTracker"** folder in the Project Navigator
2. Select **"Add Files to 'FinanceTracker'"**
3. Navigate to `ios/FinanceTracker/FinanceTracker/`
4. **Select all folders** (Cmd+A), but **exclude**:
   - Any `.xcdatamodeld` folder (we'll handle separately)
   - `README.md` files
5. Make sure **"Copy items if needed"** is **checked**
6. Make sure **"Create groups"** is selected (not "Create folder references")
7. Click **Add**

### Option B: Manual Folder Structure

After adding, your Project Navigator should look like:
```
FinanceTracker/
├── App/
├── Core/
├── Data/
├── Domain/
├── Presentation/
├── Resources/
└── Widgets/
```

## Step 3: Add the Core Data Model

1. In Project Navigator, locate `Core/Persistence/`
2. Right-click on the `Persistence` folder
3. Select **"Add Files to 'Persistence'"**
4. Navigate to `FinanceTracker.xcdatamodeld`
5. **Add the entire folder** (this is important for Core Data models)
6. Select **"Create folder references"**

## Step 4: Configure Project Settings

### Deployment Target
1. Click the **FinanceTracker** project (blue icon)
2. Select **General** tab
3. Under **Deployment Info**:
   - **iOS Deployment Target**: Set to **16.0**
   - **iPhone/iPad**: Uncheck any you don't support

### Add Frameworks
1. Go to **Build Phases** → **Link Binary With Libraries**
2. Click **+** and add:
   - None needed! (we only use system frameworks)
3. But make sure under **Frameworks, Libraries, and Embedded Content**:
   - `SwiftUI.framework` should be there (auto-added)

### Info.plist
1. Open `Info.plist` (if it exists, or create in Resources)
2. Add these keys if not present:

```xml
<key>NSFaceIDUsageDescription</key>
<string>Use Face ID to securely access your financial data</string>

<key>NSBiometricUsageDescription</key>
<string>Authenticate to access Finance Tracker</string>
```

## Step 5: Fix Any Build Errors

### Expected Errors & Fixes

**Error 1: Missing Core Data references**
- Add `import CoreData` to files that need it if not already there

**Error 2: "Cannot find 'DesignSystem' in scope"**
- Make sure `DesignSystem.swift` is added to the target
- Clean build: **Product → Clean Build Folder** (Shift+Cmd+K)

**Error 3: Widget extension errors**
- For now, you can exclude the Widgets folder from the target
- We'll add widgets as a separate extension later

## Step 6: Build and Run

1. Select a simulator (e.g., **iPhone 15**)
2. Click the **Run** button (▶️) or press **Cmd+R**
3. The app should launch showing the onboarding screen

## Step 7: Configure API (for Backend Connection)

1. Create a `Config.plist` file:
   - File → New → File → Resource → Property List
   - Name it `Config.plist`
   - Add:
   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
   <dict>
       <key>baseURL</key>
       <string>http://localhost:8080/api/v1</string>
       <key>apiKey</key>
       <string></string>
       <key>timeout</key>
       <real>30</real>
   </dict>
   </plist>
   ```

2. Add `Config.plist` to the target (File Inspector → Target Membership)

## Project Structure After Setup

Your Xcode Project Navigator should have this structure:

```
FinanceTracker
├── FinanceTracker
│   ├── FinanceTrackerApp.swift ✓
│   ├── RootView.swift ✓
│   │
│   ├── Core/
│   │   ├── Network/
│   │   │   ├── APIClient.swift ✓
│   │   │   ├── APIEndpoint.swift ✓
│   │   │   └── NetworkError.swift ✓
│   │   ├── Keychain/
│   │   │   └── KeychainManager.swift ✓
│   │   ├── Security/
│   │   │   └── BiometricAuthManager.swift ✓
│   │   ├── Persistence/
│   │   │   ├── CoreDataManager.swift ✓
│   │   │   ├── FinanceTracker.xcdatamodeld ✓
│   │   │   └── Models/
│   │   ├── Notification/
│   │   │   └── NotificationManager.swift ✓
│   │   └── Export/
│   │       └── ExportManager.swift ✓
│   │
│   ├── Data/
│   │   ├── Models/ (Transaction.swift, etc.)
│   │   ├── Repositories/
│   │   └── DataSources/
│   │
│   ├── Domain/
│   │   ├── UseCases/
│   │   └── Services/
│   │
│   ├── Presentation/
│   │   ├── Common/
│   │   │   ├── DesignSystem/
│   │   │   ├── Components/
│   │   │   └── Views/
│   │   ├── Onboarding/
│   │   ├── Dashboard/
│   │   ├── Transactions/
│   │   ├── Analytics/
│   │   ├── Goals/
│   │   └── Settings/
│   │
│   ├── Resources/
│   │   └── Assets.xcassets/ (you can add images later)
│   │
│   └── Widgets/
│       └── FinanceTrackerBundle/ (can be added later)
```

## Quick Troubleshooting

| Error | Solution |
|-------|----------|
| "Command CompileSwift failed" | Make sure all Swift files are added to target |
| "Cannot find 'DesignSystem'" | Clean build, check DesignSystem.swift target membership |
| "Use of unresolved identifier" | Check imports, clean build folder |
| "Core Data model not found" | Ensure .xcdatamodeld is added as folder reference |
| Widget errors | Exclude Widgets folder for now, add later as extension |

## First Run Experience

When you run the app, you should see:
1. **Welcome Screen** - App logo and "Get Started" button
2. **API Key Setup** - Enter your backend API key (or skip)
3. **Biometric Setup** - Enable Face ID/Touch ID
4. **Dashboard** - Empty state with "Add Transaction" option
