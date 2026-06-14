# White buttons on Dashboard first open

## Bug Description

When DashboardView opens (navigated from WelcomeView), three specific buttons appear with their default WPF theme background (white on Windows 11 Fluent theme) instead of the custom colors from `SharedStyles.xaml`. Clicking any part of the DashboardView screen triggers a re-render that applies the correct styled colors.

## Affected

- **Window**: DashboardView only
- **Elements**: Logout button, Set (bitrate) button, View Events button
- **Not affected**: Window background, TextBlocks, TextBoxes, Borders, Copy buttons (all render with correct shared styles immediately)
- **Platform**: Windows 11, .NET 9, WPF Fluent theme

## What renders correctly vs incorrectly

| Element | Style | Renders correctly? |
|---------|-------|--------------------|
| Window background | `AppWindowStyle` (sets `Background` via Setter) | Yes |
| Page title | `PageTitleStyle` (sets `Foreground` via Setter) | Yes |
| TextBox inputs | `InputTextBoxStyle`, `ReadOnlyTextBoxStyle` (both set `Background` via Setter) | Yes |
| Section borders | Direct `Background="{StaticResource ColorSectionBg}"` | Yes |
| Logout button | `SecondaryButtonStyle` | **No — white until clicked** |
| Set button | `SmallDarkButtonStyle` | **No — white until clicked** |
| View Events button | `SmallPrimaryButtonStyle` | **No — white until clicked** |
| Copy buttons | `SmallDarkButtonStyle` | **Yes** (not affected) |

## Attempts

### Attempt 1 — Inline `Background`/`Foreground` with `{StaticResource}`

Added `Background="{StaticResource ColorSectionBg}"` and `Foreground="{StaticResource ColorWhiteText}"` directly on each Button element in DashboardView.xaml (and defensively in EventsView, DestinationView, WelcomeView). This creates a "local value" in WPF's property precedence, which should override both style setters and theme defaults.

**Result**: Did not fix the issue.

**Lesson learned**: The problem is not about WPF property precedence (local value vs style setter). The `StaticResource` is resolving to the correct brush (otherwise XAML would throw a `XamlParseException`), but the Button's rendering pipeline ignores it on the initial paint.

### Attempt 2 — Force style re-apply in `Loaded` event

Added a `Loaded` handler on DashboardView that iterates all Buttons via `VisualTreeHelper` and toggles `button.Style = null; button.Style = originalStyle;`. This forces a complete re-template and visual state reset after the window is fully loaded.

**Result**: Did not fix the issue.

**Lesson learned**: Even forcing a full style re-apply after the window is loaded does not trigger the correct rendering. This rules out a simple "styles not applied yet" timing issue. The incorrect rendered state persists through style toggling.

## Replication

1. Login with valid credentials
2. Click "Proceed to Dashboard" on WelcomeView
3. Observe Logout, Copy, Set, and View Events buttons — all white
4. Click anywhere on the DashboardView — buttons snap to correct colors
