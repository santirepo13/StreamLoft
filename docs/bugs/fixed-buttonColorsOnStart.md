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

### Attempt 3 — Implicit Button style in App.xaml to override Fluent theme

Added `<Style TargetType="Button" BasedOn="{StaticResource ButtonBaseStyle}"/>` to `App.xaml` resources. This creates an implicit style that all buttons inherit, intended to prevent the Fluent theme from applying its default button template.

**Result**: Did not fix the issue.

**Lesson learned**: The Fluent theme's button template is applied at a lower level (ControlTemplate) that an implicit style based on ButtonBaseStyle doesn't override. The theme's template takes precedence during initial render.

### Attempt 4 — Switch button style setters from `StaticResource` to `DynamicResource`

Changed all Background/Foreground setters in `SharedStyles.xaml` button styles (PrimaryButtonStyle, SecondaryButtonStyle, SmallDarkButtonStyle, SmallPrimaryButtonStyle, DangerButtonStyle) from `{StaticResource ...}` to `{DynamicResource ...}`. This ensures brushes resolve at render time rather than load time.

**Result**: Did not fix the issue.

**Lesson learned**: The issue is not resource resolution timing. The Fluent theme's ControlTemplate completely replaces the button's visual tree, so Background/Foreground setters on the style are ignored until a layout invalidation forces re-templating.

### Attempt 5 — Force `UpdateLayout()` in `Loaded` event via Dispatcher

Added `Loaded += (s, e) => Dispatcher.BeginInvoke(() => UpdateLayout(), DispatcherPriority.Loaded);` in DashboardView.xaml.cs. This forces a layout pass after the window is fully loaded, hoping to trigger the correct template application.

**Result**: Did not fix the issue.

**Lesson learned**: The initial render with Fluent theme happens before Loaded fires. Forcing UpdateLayout after Loaded doesn't re-apply templates because the visual tree is already realized with the wrong template.

## Resolution

### Fix — Custom ControlTemplate in ButtonBaseStyle (SharedStyles.xaml:30-47)

**Root cause**: The bug report mentions "Fluent theme" but the project uses a custom theme only. The actual issue is WPF's default Button ControlTemplate on Windows 11, which applies a white background that overrides style setters during initial render. The copy buttons (also `SmallDarkButtonStyle`) worked because they sit inside a `Border` with explicit background; the affected buttons sit directly on the Window background where the system template shows through.

**Solution**: Added a custom `ControlTemplate` to `ButtonBaseStyle` that replaces the system template entirely:

```xml
<Style x:Key="ButtonBaseStyle" TargetType="Button">
    <Setter Property="Cursor" Value="Hand"/>
    <Setter Property="BorderThickness" Value="0"/>
    <Setter Property="Template">
        <Setter.Value>
            <ControlTemplate TargetType="Button">
                <Border Background="{TemplateBinding Background}"
                        BorderBrush="{TemplateBinding BorderBrush}"
                        BorderThickness="{TemplateBinding BorderThickness}"
                        CornerRadius="4"
                        Padding="{TemplateBinding Padding}">
                    <ContentPresenter HorizontalAlignment="Center"
                                      VerticalAlignment="Center"/>
                </Border>
            </ControlTemplate>
        </Setter.Value>
    </Setter>
</Style>
```

**Why this works**: Instead of trying to override the system template's Background (which doesn't work on initial render), we replace the template entirely. All 5 button styles (`PrimaryButtonStyle`, `SecondaryButtonStyle`, `SmallDarkButtonStyle`, `SmallPrimaryButtonStyle`, `DangerButtonStyle`) inherit this template via `BasedOn`, so all buttons render with correct custom colors on first paint — no click, no layout pass, no code-behind needed.

**Result**: Fixed. Buttons render with correct colors immediately on DashboardView first open.

## Replication

1. Login with valid credentials
2. Click "Proceed to Dashboard" on WelcomeView
3. Observe Logout, Copy, Set, and View Events buttons — all white
4. Click anywhere on the DashboardView — buttons snap to correct colors
