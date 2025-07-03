# UI Package Structure

This package contains the user interface components for the GitHub Profile Manager, organized in a modular, component-based architecture.

## Directory Structure

```
internal/ui/
├── README.md                 # This file - UI architecture documentation
├── ui.go                     # Main UI coordinator (112 lines)
├── profile_list.go           # Profile list component (77 lines)
├── status_display.go         # Status display component (48 lines)
├── toolbar.go                # Main toolbar coordinator (142 lines)
├── actions/
│   └── profile_actions.go    # Profile management actions (136 lines)
└── dialogs/
    ├── detect_dialog.go      # Current profile detection dialog (66 lines)
    └── profile_dialog.go     # Profile creation/editing dialog (140 lines)
```

## Component Responsibilities

### Core Components

- **ui.go**: Main UI coordinator that manages window setup, component lifecycle, and data flow
- **profile_list.go**: Displays and manages the list of profiles with visual indicators
- **status_display.go**: Shows current git configuration and active profile status
- **toolbar.go**: Coordinates all user actions through buttons and delegates to specialized components

### Actions Package

- **profile_actions.go**: Handles all profile operations (import, export, delete, switch, SSH testing)

### Dialogs Package

- **detect_dialog.go**: Dialog for detecting and creating profiles from current system configuration
- **profile_dialog.go**: Dialog for creating new profiles or editing existing ones with SSH key management

## Architecture Benefits

1. **Separation of Concerns**: Each component has a single, well-defined responsibility
2. **Modularity**: Components can be modified independently without affecting others
3. **Testability**: Each component can be unit tested in isolation
4. **Maintainability**: Smaller files are easier to understand and modify
5. **Reusability**: Dialog and action components can be reused across different parts of the UI

## Data Flow

```
UI (Coordinator)
├── Creates and manages all components
├── Provides data access methods (GetProfiles, GetConfig, etc.)
└── Handles refresh operations

Toolbar (Action Coordinator)
├── Creates action handlers and dialogs
├── Delegates specific operations to specialized components
└── Handles user interaction callbacks

Actions & Dialogs
├── Perform specific operations (profile management, system detection)
├── Handle their own error states and user feedback
└── Call back to UI for refresh operations
```

## File Size Comparison

**Before Refactoring:**
- `ui.go`: 622 lines (monolithic)

**After Refactoring:**
- Total lines: ~579 lines across 7 files
- Largest single file: 142 lines (toolbar.go)
- Average file size: ~83 lines
- Improved maintainability and readability

## Adding New Features

To add new UI features:

1. **Simple operations**: Add to existing action classes
2. **Complex dialogs**: Create new dialog in `/dialogs` package
3. **New data displays**: Create new component in main UI package
4. **Cross-cutting features**: Add to UI coordinator with proper delegation

This architecture supports easy extension while maintaining clean separation of concerns.