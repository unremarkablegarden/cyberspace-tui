# Cyberspace CLI - Bubbletea Implementation Plan

A minimal TUI client for Cyberspace built with Go and Bubbletea. Connects to the existing Firebase backend via REST API.

## Project Setup

### 1. Initialize Go Project

```bash
mkdir cyberspace-cli
cd cyberspace-cli
go mod init github.com/yourusername/cyberspace-cli
```

### 2. Install Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
go get github.com/joho/godotenv
```

- `bubbletea` - TUI framework (Elm architecture)
- `bubbles` - Pre-built components (text input, list, spinner, etc.)
- `lipgloss` - Terminal styling
- `godotenv` - Load environment from `.env` file

## Architecture

```
cyberspace-cli/
├── main.go              # Entry point, app initialization
├── config.go            # Firebase config, stored credentials
├── api/
│   ├── client.go        # HTTP client for Firebase REST API
│   ├── auth.go          # Firebase Auth REST endpoints
│   └── posts.go         # Firestore REST endpoints for posts
├── models/
│   ├── user.go          # User struct
│   └── post.go          # Post struct
├── views/
│   ├── login.go         # Login screen (email/password)
│   ├── feed.go          # Post feed list
│   └── post_detail.go   # Single post view with replies
└── styles/
    └── theme.go         # Lipgloss styles (terminal colors)
```

## Phase 1: Authentication

### Step 1.1: Firebase REST Auth Client

Create `api/auth.go`:

```go
package api

const firebaseAuthURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"

type AuthRequest struct {
    Email             string `json:"email"`
    Password          string `json:"password"`
    ReturnSecureToken bool   `json:"returnSecureToken"`
}

type AuthResponse struct {
    IDToken      string `json:"idToken"`
    RefreshToken string `json:"refreshToken"`
    ExpiresIn    string `json:"expiresIn"`
    LocalID      string `json:"localId"`
    Email        string `json:"email"`
}

func SignIn(email, password, apiKey string) (*AuthResponse, error) {
    // POST to firebaseAuthURL?key=API_KEY
    // Return tokens on success
}
```

### Step 1.2: Token Storage

Create `config.go`:

```go
package main

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    IDToken      string `json:"id_token"`
    RefreshToken string `json:"refresh_token"`
    UserID       string `json:"user_id"`
    Email        string `json:"email"`
}

func configPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".cyberspace", "config.json")
}

func LoadConfig() (*Config, error) { /* ... */ }
func SaveConfig(c *Config) error { /* ... */ }
```

### Step 1.3: Login View

Create `views/login.go`:

```go
package views

import (
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
)

type LoginModel struct {
    emailInput    textinput.Model
    passwordInput textinput.Model
    focusIndex    int
    err           error
    loading       bool
}

func (m LoginModel) Init() tea.Cmd { return textinput.Blink }

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab", "down":
            m.focusIndex = (m.focusIndex + 1) % 2
        case "shift+tab", "up":
            m.focusIndex = (m.focusIndex - 1 + 2) % 2
        case "enter":
            if m.focusIndex == 1 || allFieldsFilled(m) {
                return m, attemptLogin(m.emailInput.Value(), m.passwordInput.Value())
            }
        case "ctrl+c", "q":
            return m, tea.Quit
        }
    case LoginSuccessMsg:
        // Transition to feed view
    case LoginErrorMsg:
        m.err = msg.err
    }
    // Update focused input
    return m, nil
}

func (m LoginModel) View() string {
    return `
  ╭─────────────────────────────╮
  │      CYBERSPACE LOGIN       │
  ├─────────────────────────────┤
  │  Email:    [____________]   │
  │  Password: [____________]   │
  │                             │
  │      [ LOGIN ]              │
  ╰─────────────────────────────╯
    `
}
```

## Phase 2: Post Feed

### Step 2.1: Firestore REST Client

Create `api/posts.go`:

```go
package api

const firestoreURL = "https://firestore.googleapis.com/v1/projects/YOUR_PROJECT/databases/(default)/documents"

type Post struct {
    ID             string    `json:"id"`
    AuthorID       string    `json:"authorId"`
    AuthorUsername string    `json:"authorUsername"`
    Content        string    `json:"content"`
    CreatedAt      time.Time `json:"createdAt"`
    RepliesCount   int       `json:"repliesCount"`
    BookmarksCount int       `json:"bookmarksCount"`
    Topics         []string  `json:"topics"`
}

func FetchPosts(idToken string, limit int, startAfter string) ([]Post, error) {
    // Use Firestore REST API with structured query
    // GET /documents:runQuery with Bearer token
    // Filter: deleted == false
    // Order: createdAt DESC
    // Limit: 20
}
```

Firestore REST query format:
```json
{
  "structuredQuery": {
    "from": [{"collectionId": "posts"}],
    "where": {
      "fieldFilter": {
        "field": {"fieldPath": "deleted"},
        "op": "EQUAL",
        "value": {"booleanValue": false}
      }
    },
    "orderBy": [{"field": {"fieldPath": "createdAt"}, "direction": "DESCENDING"}],
    "limit": 20
  }
}
```

### Step 2.2: Feed View

Create `views/feed.go`:

```go
package views

import (
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
)

type FeedModel struct {
    list     list.Model
    posts    []Post
    loading  bool
    err      error
    cursor   int
}

func (m FeedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "r":
            return m, refreshFeed()
        case "enter":
            // Open post detail
            return m, openPost(m.posts[m.cursor].ID)
        case "q":
            return m, tea.Quit
        case "j", "down":
            m.cursor++
        case "k", "up":
            m.cursor--
        }
    case PostsLoadedMsg:
        m.posts = msg.posts
        m.loading = false
    }
    return m, nil
}

func (m FeedModel) View() string {
    // Render post list
    // Show: username, timestamp, content preview, reply/bookmark counts
}
```

### Step 2.3: Post Rendering

```go
func renderPost(p Post, selected bool) string {
    style := lipgloss.NewStyle()
    if selected {
        style = style.Background(lipgloss.Color("236"))
    }

    header := fmt.Sprintf("@%s · %s", p.AuthorUsername, timeAgo(p.CreatedAt))
    content := truncate(stripMarkdown(p.Content), 80)
    footer := fmt.Sprintf("↩ %d  ★ %d", p.RepliesCount, p.BookmarksCount)

    return style.Render(header + "\n" + content + "\n" + footer)
}
```

## Phase 3: Post Detail View

### Step 3.1: Fetch Single Post + Replies

```go
func FetchPost(idToken, postID string) (*Post, error) {
    // GET /documents/posts/{postID}
}

func FetchReplies(idToken, postID string) ([]Reply, error) {
    // Structured query on replies collection
    // where postId == postID, orderBy createdAt ASC
}
```

### Step 3.2: Detail View

```go
type PostDetailModel struct {
    post     Post
    replies  []Reply
    scroll   int
    loading  bool
}

func (m PostDetailModel) View() string {
    // Full post content (markdown stripped to plain text)
    // Divider
    // List of replies
}
```

## Key Bindings

```
Global:
  q / ctrl+c  - Quit
  ?           - Help

Feed:
  j/k / ↑↓    - Navigate posts
  enter       - View post
  r           - Refresh feed
  n           - New post (future)

Post Detail:
  esc / b     - Back to feed
  j/k / ↑↓    - Scroll content
  r           - Reply (future)
```

## Environment Configuration

### Install godotenv for .env support

```bash
go get github.com/joho/godotenv
```

### Create `.env` file

```bash
# .env (add to .gitignore!)
FIREBASE_API_KEY=
FIREBASE_AUTH_DOMAIN=
FIREBASE_PROJECT_ID=
FIREBASE_STORAGE_BUCKET=
FIREBASE_MESSAGING_SENDER_ID=
FIREBASE_APP_ID=
FIREBASE_DATABASE_URL=
```

### Create `.env.example` for reference

```bash
# .env.example (commit this)
FIREBASE_API_KEY=your-api-key-here
FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_STORAGE_BUCKET=your-project.firebasestorage.app
FIREBASE_MESSAGING_SENDER_ID=your-sender-id
FIREBASE_APP_ID=your-app-id
FIREBASE_DATABASE_URL=https://your-project.firebasedatabase.app/
```

### Load env in `main.go`

```go
package main

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file (optional - won't fail if missing)
    godotenv.Load()

    apiKey := os.Getenv("FIREBASE_API_KEY")
    projectID := os.Getenv("FIREBASE_PROJECT_ID")
    authDomain := os.Getenv("FIREBASE_AUTH_DOMAIN")
    databaseURL := os.Getenv("FIREBASE_DATABASE_URL")

    if apiKey == "" || projectID == "" {
        log.Fatal("FIREBASE_API_KEY and FIREBASE_PROJECT_ID must be set")
    }

    // Start app...
}
```

### For compiled binaries (no .env needed)

Users can set env vars directly:

```bash
# Linux/macOS
export FIREBASE_API_KEY=AIzaSy...
export FIREBASE_PROJECT_ID=cyberspace
./cyberspace-cli

# Or inline
FIREBASE_API_KEY=AIzaSy... FIREBASE_PROJECT_ID=cyberspace ./cyberspace-cli

# Windows PowerShell
$env:FIREBASE_API_KEY="AIzaSy..."
$env:FIREBASE_PROJECT_ID="cyberspace"
.\cyberspace-cli.exe
```

### Alternative: Embed at build time (for distribution)

```bash
# Build with hardcoded values (for releases)
go build -ldflags "-X main.FirebaseAPIKey=AIzaSy... -X main.FirebaseProjectID=cyberspace" -o cyberspace-cli
```

Then in `main.go`:

```go
var (
    FirebaseAPIKey   = "" // Set via -ldflags or env
    FirebaseProjectID = ""
)

func getConfig() (string, string) {
    apiKey := FirebaseAPIKey
    if apiKey == "" {
        apiKey = os.Getenv("FIREBASE_API_KEY")
    }
    projectID := FirebaseProjectID
    if projectID == "" {
        projectID = os.Getenv("FIREBASE_PROJECT_ID")
    }
    return apiKey, projectID
}
```

This allows:
1. Dev: Use `.env` file
2. Users: Set env vars
3. Releases: Embed values at compile time

## Build & Distribution

```bash
# Development
go run .

# Build for current platform
go build -o cyberspace

# Cross-compile for distribution
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/cyberspace-linux
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/cyberspace.exe
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/cyberspace-mac
```

## Implementation Order

1. [ ] Project setup (`go mod init`, dependencies)
2. [ ] Config loading/saving (`~/.cyberspace/config.json`)
3. [ ] Firebase Auth REST client (sign in with email/password)
4. [ ] Login view (email + password inputs)
5. [ ] Token persistence (save after login, load on startup)
6. [ ] Firestore REST client (structured queries)
7. [ ] Post model + parsing Firestore response
8. [ ] Feed view (list of posts)
9. [ ] Post detail view (full content + replies)
10. [ ] Polish: error handling, loading states, help screen

## Future Phases (Not In Scope)

- Write new posts
- Reply to posts
- Bookmarking
- User profiles
- DMs (polling-based)
- Notifications

## Notes for Implementation

1. **Firestore REST API** returns deeply nested JSON - you'll need to parse the `fields` object carefully
2. **Token refresh** - Firebase ID tokens expire after 1 hour; implement refresh token flow
3. **Markdown stripping** - Posts are stored as markdown; strip for display or use a simple terminal markdown renderer
4. **Rate limiting** - The main app has rate limits via Cloud Functions; the REST API should respect these
5. **No real-time** - This is intentional; use `r` to refresh manually
