# FortyOne Mobile App

[![Expo](https://img.shields.io/badge/Expo-~52.0.0-black.svg)](https://expo.dev/)
[![React Native](https://img.shields.io/badge/React%20Native-0.76-blue.svg)](https://reactnative.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.4-blue.svg)](https://www.typescriptlang.org/)

The React Native mobile application for FortyOne, built with Expo. Provides mobile access to project management features, real-time collaboration, and team productivity tools on iOS and Android devices.

## ✨ Features

- **📱 Native Mobile Experience**: iOS and Android apps with native performance
- **🎯 Project Management**: Full access to project management features
- **👥 Team Collaboration**: Real-time collaboration with team members
- **📋 Task Management**: Create, edit, and track tasks on mobile
- **📷 Camera Integration**: Photo and document capture for tasks
- **🔔 Push Notifications**: Real-time notifications and updates
- **🔐 Secure Authentication**: Biometric authentication support
- **📶 Offline Support**: Core functionality works offline
- **🎨 Native UI**: Platform-specific design and interactions
- **⚡ Fast Performance**: Optimized for mobile performance
- **🔄 Auto Updates**: Over-the-air updates via Expo

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- pnpm 9.3.0+
- Expo CLI: `npm install -g @expo/cli`
- iOS Simulator (macOS) or Android Studio (all platforms)

### Development Setup

1. **Install dependencies**:

   ```bash
   cd apps/mobile
   pnpm install
   ```

2. **Set up environment variables**:

   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

3. **Start the Expo development server**:

   ```bash
   pnpm start
   # or
   npx expo start
   ```

4. **Run on device/emulator**:
   - **iOS**: Press `i` in the terminal or scan QR code with Camera app
   - **Android**: Press `a` in the terminal or scan QR code with Expo Go
   - **Web**: Press `w` for web development

### Environment Variables

Create a `.env` file with:

```bash
# API Configuration
EXPO_PUBLIC_API_URL=api_url
```

## 🏗️ Architecture

### Tech Stack

- **Framework**: Expo SDK ~52.0
- **Runtime**: React Native 0.76
- **Navigation**: Expo Router (file-based routing)
- **State Management**: React Context + hooks
- **Styling**: NativeWind (Tailwind CSS for React Native)
- **HTTP Client**: Custom fetch utilities
- **Storage**: AsyncStorage for local data
- **Notifications**: Expo Notifications
- **Camera**: Expo Camera
- **Build**: EAS Build for production builds

### Key Features

- **Authentication**: Secure login with biometric support
- **Project Views**: Mobile-optimized project dashboards
- **Task Management**: Touch-friendly task creation and editing
- **File Attachments**: Camera integration for task documentation
- **Real-time Sync**: Background sync when online
- **Offline Mode**: Core functionality without internet
- **Push Notifications**: Important updates and reminders

### Directory Structure

```
apps/mobile/
├── app/                    # Expo Router app structure
│   ├── (auth)/            # Authentication screens
│   ├── (tabs)/            # Main tab navigation
│   ├── _layout.tsx        # Root layout
│   └── +not-found.tsx     # 404 screen
├── components/            # Reusable React Native components
│   ├── ui/               # Design system components
│   ├── shared/           # Shared mobile components
│   └── [feature]/        # Feature-specific components
├── lib/                  # Business logic and utilities
│   ├── actions/          # API actions
│   ├── queries/          # API queries
│   ├── hooks/            # Custom hooks
│   └── utils/            # Helper functions
├── constants/            # App constants
├── types/                # TypeScript definitions
├── assets/               # Images, fonts, and other assets
└── utils/                # Platform-specific utilities
```

## 📦 Available Scripts

```bash
# Development
pnpm start         # Start Expo development server
pnpm ios           # Run on iOS simulator
pnpm android       # Run on Android emulator
pnpm web           # Run in web browser

# Building
pnpm build:ios     # Build for iOS
pnpm build:android # Build for Android

# Utilities
pnpm reset-project # Reset to fresh Expo project
pnpm lint          # Run linting
```

## 🚀 Managed mobile delivery

Production builds, signing credentials, store submissions, and over-the-air
updates are owned by the internal mobile release process. Do not run an ad hoc
production EAS build or store submission from a personal Expo account.

Use the checked-in development profile only after receiving access through the
team's managed Expo organization. Changes to `eas.json`, native identifiers,
entitlements, signing, or release channels require review from the mobile
release owner.

## 📱 Device Testing

### Development Builds

For advanced testing features:

```bash
# Create development build
eas build --platform ios --profile development
eas build --platform android --profile development
```

### Testing Checklist

- [ ] Authentication flow works
- [ ] Offline functionality works
- [ ] Push notifications arrive
- [ ] Camera integration works
- [ ] Performance is smooth
- [ ] All screen sizes supported

## Internal development guidelines

- **Platform Testing**: Test on both iOS and Android
- **Performance**: Optimize for mobile performance
- **Accessibility**: Follow mobile accessibility guidelines
- **Offline First**: Consider offline functionality
- **Touch Interactions**: Design for touch interfaces

## 🔗 Related Projects

- **Web App**: [apps/projects/](../../apps/projects/)
- **Landing Page**: [apps/landing/](../../apps/landing/)
- **Documentation**: [apps/docs/](../../apps/docs/)
