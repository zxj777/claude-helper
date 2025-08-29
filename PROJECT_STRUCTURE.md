# 🏗️ Claude Helper Project Structure

This document explains the project organization and build safety measures.

## 📁 Directory Structure

```
claude-helper/
├── 📋 Core Project Files
│   ├── go.mod, go.sum           # Go module files
│   ├── Makefile                 # Build automation
│   ├── CLAUDE.md               # Claude Code instructions
│   └── README.md               # Main documentation
│
├── 🔧 Critical Build Files (MUST stay in root)
│   ├── build.sh                # Cross-platform build script
│   ├── install.sh              # Unix install script (GitHub URL: /main/install.sh)
│   └── install.ps1             # Windows install script (GitHub URL: /main/install.ps1)
│
├── 📚 docs/                    # Documentation files
│   ├── AUDIO_NOTIFICATION_USAGE.md
│   ├── DESKTOP_NOTIFICATION_SUMMARY.md
│   ├── IMPLEMENTATION_SUMMARY.md
│   ├── audio-notification-design.md
│   ├── implement.md
│   ├── project-description.md
│   ├── test-remove-functionality.md
│   └── text-expander-implementation.md
│
├── 🔧 scripts/                 # Development and utility scripts
│   ├── auto-release.sh         # Automated release workflow
│   ├── cleanup.sh              # Project cleanup utilities
│   ├── cleanup_templates.sh    # Template cleanup
│   ├── copy_audio_manual.py    # Audio file management
│   ├── copy_sound.py           # Sound copying utilities
│   ├── fix-hook.sh             # Hook repair script
│   ├── generate-sounds.py      # Sound generation
│   ├── quick-release.sh        # Quick release workflow
│   ├── release.sh              # Release management
│   ├── setup_audio.py          # Audio setup
│   └── setup_embedded_sound.sh # Embedded sound setup
│
├── 🚀 cmd/                     # Application entry points
│   ├── claude-helper/
│   │   └── main.go            # Main CLI application
│   └── test-platform/
│       └── main.go            # Platform testing utility
│
├── 🔒 internal/                # Internal packages (CRITICAL PATHS)
│   ├── assets/                # 🚨 CANNOT MOVE - Go embed paths
│   │   ├── assets.go          # Asset loading logic
│   │   ├── templates/         # 🔒 Required for: //go:embed templates/*
│   │   │   ├── agents/        # Agent template files
│   │   │   └── hooks/         # Hook template files
│   │   └── sounds/            # 🔒 Required for: //go:embed sounds/*
│   │       ├── README.md
│   │       └── notification.aiff
│   ├── cli/                   # CLI command implementations
│   ├── config/                # Configuration management
│   └── notification/          # Notification system
│
├── 📦 pkg/                     # Public packages
│   └── types/                 # Public type definitions
│
├── 🏗️ build/                   # Build outputs
│   ├── bin/                   # Local binaries
│   └── dist/                  # Distribution files
│
└── 🧪 Development Files
    ├── organize-project.sh     # This reorganization script
    └── test-build-integrity.sh # Build validation script
```

## 🛡️ Build Safety Measures

### 🚨 Critical Files That Cannot Be Moved

1. **Go Embed Assets** (`internal/assets/templates/`, `internal/assets/sounds/`)
   - **Why:** Go embed directives use relative paths: `//go:embed templates/*`
   - **Impact:** Moving these would break binary builds

2. **GitHub-Referenced Install Scripts** (`install.sh`, `install.ps1`)
   - **Why:** Hardcoded in GitHub URLs: `https://raw.githubusercontent.com/repo/main/install.sh`
   - **Impact:** Moving would break public install commands

3. **Build Script** (`build.sh`)
   - **Why:** Referenced by multiple release scripts
   - **Impact:** Moving requires updating all references

### ✅ Safe Reorganization Features

- **Documentation Centralized:** All `.md` files moved to `docs/`
- **Scripts Organized:** Development scripts moved to `scripts/`
- **References Updated:** Release scripts updated to use `../build.sh`
- **Cleanup Performed:** Temporary and test files removed

## 🚀 Usage After Reorganization

### Building the Project
```bash
# From project root (unchanged)
make build
# or
./build.sh v1.0.0
```

### Running Release Scripts
```bash
# From project root
./scripts/quick-release.sh v1.2.0
./scripts/auto-release.sh --dry-run
```

### Accessing Documentation
```bash
# All docs now in docs/ directory
ls docs/
cat docs/IMPLEMENTATION_SUMMARY.md
```

## 🧪 Testing Build Integrity

After reorganization, run the build integrity test:

```bash
chmod +x test-build-integrity.sh
./test-build-integrity.sh
```

This validates:
- ✅ Go embed paths accessible
- ✅ Build script executable  
- ✅ Install scripts preserved
- ✅ Go compilation successful
- ✅ Release script references updated
- ✅ Critical project files intact

## 🔄 Migration Commands

### Safe Reorganization
```bash
chmod +x organize-project.sh
./organize-project.sh
```

### Validation
```bash
chmod +x test-build-integrity.sh
./test-build-integrity.sh
```

### Build Test
```bash
make build
./bin/cchp --help
```

## ⚠️ Important Notes

1. **IDE Configurations:** Update any IDE settings that reference moved scripts
2. **External References:** Check any external tools that reference the old script locations  
3. **Documentation Links:** Update any documentation that links to moved files
4. **CI/CD Pipelines:** Verify any automated workflows still reference correct paths

## 📞 Troubleshooting

If you encounter build issues after reorganization:

1. **Check Go Embed Paths:**
   ```bash
   ls -la internal/assets/templates/
   ls -la internal/assets/sounds/
   ```

2. **Verify Build Script:**
   ```bash
   ls -la build.sh
   ./build.sh --help
   ```

3. **Test Compilation:**
   ```bash
   go build ./cmd/claude-helper
   ```

4. **Run Full Test Suite:**
   ```bash
   make test
   ./test-build-integrity.sh
   ```