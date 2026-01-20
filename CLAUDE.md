# Personal Finance Tracker Mobile App

## Project Overview
A mobile application for tracking financial transactions across multiple digital banks and e-wallets on a user's phone. The app automatically records transactions (money in/out) and provides dashboard and analytics features.

## Platform Priority
- **Primary**: iOS
- **Secondary**: Android (future)

## Core Features

### 1. Transaction Tracking
- Monitor incoming and outgoing transactions from:
  - Digital banking apps
  - E-wallet apps
- Record transaction details:
  - Amount
  - Date/time
  - Sender/recipient
  - Transaction type (transfer, payment, receipt, etc.)

### 2. Dashboard
- Visual overview of financial activity
- Balance summaries across all accounts
- Recent transactions list
- Income vs. expense comparison

### 3. Analytics
- Spending patterns analysis
- Income trends
- Category-based breakdowns
- Transaction history insights

## Technical Considerations

### iOS Integration
- Screen scraping / notification reading capabilities
- Bank API integration (if available)
- Secure data storage (Keychain, encrypted storage)
- Background transaction monitoring

### Data Privacy & Security
- All financial data stored locally
- No cloud synchronization without explicit user consent
- Secure authentication (Face ID / Touch ID)
- Encrypted data at rest

## Tech Stack

### Frontend (iOS)
- **Framework**: SwiftUI
- **Language**: Swift
- **Architecture**: MVVM (Model-View-ViewModel)
- **Networking**: URLSession
- **Data Storage**: Core Data + Keychain for sensitive data

### Backend
- **Language**: Golang
- **Framework**: Gin or Echo (web framework)
- **Architecture**: Clean Architecture / Hexagonal
- **API**: RESTful

### Database
- **Primary**: PostgreSQL
- **Migrations**: golang-migrate
- **ORM**: GORM

### Deployment
- **Orchestration**: Kubernetes
- **Container**: Docker
- **CI/CD**: GitHub Actions

## Current Status
- Project initialization phase
- iOS platform priority
