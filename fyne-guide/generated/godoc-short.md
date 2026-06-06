# Fyne v2.7.2 key package godoc snapshot

## fyne.io/fyne/v2

```text
const AnimationRepeatForever = -1
const KeyModifierShortcutDefault = KeyModifierControl
var AnimationEaseInOut = animationEaseInOut ...
func Do(fn func())
func DoAndWait(fn func())
func IsHorizontal(orient DeviceOrientation) bool
func IsVertical(orient DeviceOrientation) bool
func LogError(reason string, err error)
func Max(x, y float32) float32
func Min(x, y float32) float32
func SetCurrentApp(current App)
type Animation struct{ ... }
    func NewAnimation(d time.Duration, fn func(float32)) *Animation
type AnimationCurve func(float32) float32
type App interface{ ... }
    func CurrentApp() App
type AppMetadata struct{ ... }
type BuildType int
    const BuildStandard BuildType = iota ...
type Canvas interface{ ... }
type CanvasObject interface{ ... }
type Clipboard interface{ ... }
type CloudProvider interface{ ... }
type CloudProviderPreferences interface{ ... }
type CloudProviderStorage interface{ ... }
type Container struct{ ... }
    func NewContainer(objects ...CanvasObject) *Container
    func NewContainerWithLayout(layout Layout, objects ...CanvasObject) *Container
    func NewContainerWithoutLayout(objects ...CanvasObject) *Container
type Delta struct{ ... }
    func NewDelta(dx float32, dy float32) Delta
type Device interface{ ... }
    func CurrentDevice() Device
type DeviceOrientation int
    const OrientationVertical DeviceOrientation = iota ...
type Disableable interface{ ... }
type DoubleTappable interface{ ... }
type DragEvent struct{ ... }
type Draggable interface{ ... }
type Driver interface{ ... }
type Focusable interface{ ... }
type HardwareKey struct{ ... }
type KeyEvent struct{ ... }
type KeyModifier int
    const KeyModifierShift KeyModifier = 1 << iota ...
type KeyName string
    const KeyEscape KeyName = "Escape" ...
type KeyboardShortcut interface{ ... }
type Layout interface{ ... }
type LegacyTheme interface{ ... }
type Lifecycle interface{ ... }
type ListableURI interface{ ... }
type Locale string
type MainMenu struct{ ... }
    func NewMainMenu(items ...*Menu) *MainMenu
type Menu struct{ ... }
    func NewMenu(label string, items ...*MenuItem) *Menu
type MenuItem struct{ ... }
    func NewMenuItem(label string, action func()) *MenuItem
    func NewMenuItemSeparator() *MenuItem
    func NewMenuItemWithIcon(label string, icon Resource, action func()) *MenuItem
type Notification struct{ ... }
    func NewNotification(title, content string) *Notification
type OverlayStack interface{ ... }
type PointEvent struct{ ... }
type Position struct{ ... }
    func NewPos(x float32, y float32) Position
    func NewSquareOffsetPos(length float32) Position
type Preferences interface{ ... }
type Resource interface{ ... }
    func LoadResourceFromPath(path string) (Resource, error)
    func LoadResourceFromURLString(urlStr string) (Resource, error)
type ScrollDirection int
    const ScrollBoth ScrollDirection = iota ...
type ScrollEvent struct{ ... }
type Scrollable interface{ ... }
type SecondaryTappable interface{ ... }
type Settings interface{ ... }
type Shortcut interface{ ... }
type ShortcutCopy struct{ ... }
type ShortcutCut struct{ ... }
type ShortcutHandler struct{ ... }
type ShortcutPaste struct{ ... }
type ShortcutRedo struct{}
type ShortcutSelectAll struct{}
type ShortcutUndo struct{}
type Shortcutable interface{ ... }
type Size struct{ ... }
    func MeasureText(text string, size float32, style TextStyle) Size
    func NewSize(w float32, h float32) Size
    func NewSquareSize(side float32) Size
type StaticResource struct{ ... }
    func NewStaticResource(name string, content []byte) *StaticResource
type Storage interface{ ... }
type StringValidator func(string) error
type Tabbable interface{ ... }
type Tappable interface{ ... }
type TextAlign int
    const TextAlignLeading TextAlign = iota ...
type TextStyle struct{ ... }
type TextTruncation int
    const TextTruncateOff TextTruncation = iota ...
type TextWrap int
    const TextWrapOff TextWrap = iota ...
type Theme interface{ ... }
type ThemeColorName string
type ThemeIconName string
type ThemeSizeName string
type ThemeVariant uint
type ThemedResource interface{ ... }
type URI interface{ ... }
type URIReadCloser interface{ ... }
type URIWithIcon interface{ ... }
type URIWriteCloser interface{ ... }
type Validatable interface{ ... }
type Vector2 interface{ ... }
type Widget interface{ ... }
type WidgetRenderer interface{ ... }
type Window interface{ ... }
```

## fyne.io/fyne/v2/app

```text
func New() fyne.App
func NewWithID(id string) fyne.App
func SetMetadata(m fyne.AppMetadata)
type SettingsSchema struct{ ... }
```

## fyne.io/fyne/v2/canvas

```text
const DurationStandard = time.Millisecond * 300 ...
const RadiusMaximum float32 = math.MaxFloat32
func NewColorRGBAAnimation(start, stop color.Color, d time.Duration, fn func(color.Color)) *fyne.Animation
func NewPositionAnimation(start, stop fyne.Position, d time.Duration, fn func(fyne.Position)) *fyne.Animation
func NewSizeAnimation(start, stop fyne.Size, d time.Duration, fn func(fyne.Size)) *fyne.Animation
func RecolorSVG(svgContent []byte, color color.Color) ([]byte, error)
func Refresh(obj fyne.CanvasObject)
type Arc struct{ ... }
    func NewArc(startAngle, endAngle, cutoutRatio float32, color color.Color) *Arc
    func NewDoughnutArc(startAngle, endAngle float32, color color.Color) *Arc
    func NewPieArc(startAngle, endAngle float32, color color.Color) *Arc
type Circle struct{ ... }
    func NewCircle(color color.Color) *Circle
type Image struct{ ... }
    func NewImageFromFile(file string) *Image
    func NewImageFromImage(img image.Image) *Image
    func NewImageFromReader(read io.Reader, name string) *Image
    func NewImageFromResource(res fyne.Resource) *Image
    func NewImageFromURI(uri fyne.URI) *Image
type ImageFill int
    const ImageFillStretch ImageFill = iota ...
type ImageScale int32
    const ImageScaleSmooth ImageScale = iota ...
type Line struct{ ... }
    func NewLine(color color.Color) *Line
type LinearGradient struct{ ... }
    func NewHorizontalGradient(start, end color.Color) *LinearGradient
    func NewLinearGradient(start, end color.Color, angle float64) *LinearGradient
    func NewVerticalGradient(start color.Color, end color.Color) *LinearGradient
type Polygon struct{ ... }
    func NewPolygon(sides uint, color color.Color) *Polygon
type RadialGradient struct{ ... }
    func NewRadialGradient(start, end color.Color) *RadialGradient
type Raster struct{ ... }
    func NewRaster(generate func(w, h int) image.Image) *Raster
    func NewRasterFromImage(img image.Image) *Raster
    func NewRasterWithPixels(pixelColor func(x, y, w, h int) color.Color) *Raster
type Rectangle struct{ ... }
    func NewRectangle(color color.Color) *Rectangle
    func NewSquare(color color.Color) *Rectangle
type Text struct{ ... }
    func NewText(text string, color color.Color) *Text
```

## fyne.io/fyne/v2/container

```text
const ScrollBoth ScrollDirection = fyne.ScrollBoth ...
func New(layout fyne.Layout, objects ...fyne.CanvasObject) *fyne.Container
func NewAdaptiveGrid(rowcols int, objects ...fyne.CanvasObject) *fyne.Container
func NewBorder(top, bottom, left, right fyne.CanvasObject, objects ...fyne.CanvasObject) *fyne.Container
func NewCenter(objects ...fyne.CanvasObject) *fyne.Container
func NewGridWithColumns(cols int, objects ...fyne.CanvasObject) *fyne.Container
func NewGridWithRows(rows int, objects ...fyne.CanvasObject) *fyne.Container
func NewGridWrap(size fyne.Size, objects ...fyne.CanvasObject) *fyne.Container
func NewHBox(objects ...fyne.CanvasObject) *fyne.Container
func NewMax(objects ...fyne.CanvasObject) *fyne.Container
func NewPadded(objects ...fyne.CanvasObject) *fyne.Container
func NewStack(objects ...fyne.CanvasObject) *fyne.Container
func NewVBox(objects ...fyne.CanvasObject) *fyne.Container
func NewWithoutLayout(objects ...fyne.CanvasObject) *fyne.Container
type AppTabs struct{ ... }
    func NewAppTabs(items ...*TabItem) *AppTabs
type Clip struct{ ... }
    func NewClip(content fyne.CanvasObject) *Clip
type DocTabs struct{ ... }
    func NewDocTabs(items ...*TabItem) *DocTabs
type InnerWindow struct{ ... }
    func NewInnerWindow(title string, content fyne.CanvasObject) *InnerWindow
type MultipleWindows struct{ ... }
    func NewMultipleWindows(wins ...*InnerWindow) *MultipleWindows
type Navigation struct{ ... }
    func NewNavigation(root fyne.CanvasObject) *Navigation
    func NewNavigationWithTitle(root fyne.CanvasObject, s string) *Navigation
type Scroll = widget.Scroll
    func NewHScroll(content fyne.CanvasObject) *Scroll
    func NewScroll(content fyne.CanvasObject) *Scroll
    func NewVScroll(content fyne.CanvasObject) *Scroll
type ScrollDirection = fyne.ScrollDirection
type Split struct{ ... }
    func NewHSplit(leading, trailing fyne.CanvasObject) *Split
    func NewVSplit(top, bottom fyne.CanvasObject) *Split
type TabItem struct{ ... }
    func NewTabItem(text string, content fyne.CanvasObject) *TabItem
    func NewTabItemWithIcon(text string, icon fyne.Resource, content fyne.CanvasObject) *TabItem
type TabLocation int
    const TabLocationTop TabLocation = iota ...
type ThemeOverride struct{ ... }
    func NewThemeOverride(obj fyne.CanvasObject, th fyne.Theme) *ThemeOverride
```

## fyne.io/fyne/v2/widget

```text
var RichTextStyleBlockquote = RichTextStyle{ ... } ...
func NewSimpleRenderer(object fyne.CanvasObject) fyne.WidgetRenderer
func ShowModalPopUp(content fyne.CanvasObject, canvas fyne.Canvas)
func ShowPopUp(content fyne.CanvasObject, canvas fyne.Canvas)
func ShowPopUpAtPosition(content fyne.CanvasObject, canvas fyne.Canvas, pos fyne.Position)
func ShowPopUpAtRelativePosition(content fyne.CanvasObject, canvas fyne.Canvas, rel fyne.Position, ...)
func ShowPopUpMenuAtPosition(menu *fyne.Menu, c fyne.Canvas, pos fyne.Position)
func ShowPopUpMenuAtRelativePosition(menu *fyne.Menu, c fyne.Canvas, rel fyne.Position, to fyne.CanvasObject)
type Accordion struct{ ... }
    func NewAccordion(items ...*AccordionItem) *Accordion
type AccordionItem struct{ ... }
    func NewAccordionItem(title string, detail fyne.CanvasObject) *AccordionItem
type Activity struct{ ... }
    func NewActivity() *Activity
type BaseWidget struct{ ... }
type Button struct{ ... }
    func NewButton(label string, tapped func()) *Button
    func NewButtonWithIcon(label string, icon fyne.Resource, tapped func()) *Button
type ButtonAlign int
    const ButtonAlignCenter ButtonAlign = iota ...
type ButtonIconPlacement int
    const ButtonIconLeadingText ButtonIconPlacement = iota ...
type ButtonImportance = Importance
type ButtonStyle int
type Calendar struct{ ... }
    func NewCalendar(cT time.Time, changed func(time.Time)) *Calendar
type Card struct{ ... }
    func NewCard(title, subtitle string, content fyne.CanvasObject) *Card
type Check struct{ ... }
    func NewCheck(label string, changed func(bool)) *Check
    func NewCheckWithData(label string, data binding.Bool) *Check
type CheckGroup struct{ ... }
    func NewCheckGroup(options []string, changed func([]string)) *CheckGroup
type CustomTextGridStyle struct{ ... }
type DateEntry struct{ ... }
    func NewDateEntry() *DateEntry
type DisableableWidget struct{ ... }
type Entry struct{ ... }
    func NewEntry() *Entry
    func NewEntryWithData(data binding.String) *Entry
    func NewMultiLineEntry() *Entry
    func NewPasswordEntry() *Entry
type FileIcon struct{ ... }
    func NewFileIcon(uri fyne.URI) *FileIcon
type Form struct{ ... }
    func NewForm(items ...*FormItem) *Form
type FormItem struct{ ... }
    func NewFormItem(text string, widget fyne.CanvasObject) *FormItem
type GridWrap struct{ ... }
    func NewGridWrap(length func() int, createItem func() fyne.CanvasObject, ...) *GridWrap
    func NewGridWrapWithData(data binding.DataList, createItem func() fyne.CanvasObject, ...) *GridWrap
type GridWrapItemID = int
type Hyperlink struct{ ... }
    func NewHyperlink(text string, url *url.URL) *Hyperlink
    func NewHyperlinkWithStyle(text string, url *url.URL, alignment fyne.TextAlign, style fyne.TextStyle) *Hyperlink
type HyperlinkSegment struct{ ... }
type Icon struct{ ... }
    func NewIcon(res fyne.Resource) *Icon
type ImageSegment struct{ ... }
type Importance int
    const MediumImportance Importance = iota ...
type Label struct{ ... }
    func NewLabel(text string) *Label
    func NewLabelWithData(data binding.String) *Label
    func NewLabelWithStyle(text string, alignment fyne.TextAlign, style fyne.TextStyle) *Label
type List struct{ ... }
    func NewList(length func() int, createItem func() fyne.CanvasObject, ...) *List
    func NewListWithData(data binding.DataList, createItem func() fyne.CanvasObject, ...) *List
type ListItemID = int
type ListSegment struct{ ... }
type Menu struct{ ... }
    func NewMenu(menu *fyne.Menu) *Menu
type Orientation int
    const Horizontal Orientation = 0 ...
type ParagraphSegment struct{ ... }
type PopUp struct{ ... }
    func NewModalPopUp(content fyne.CanvasObject, canvas fyne.Canvas) *PopUp
    func NewPopUp(content fyne.CanvasObject, canvas fyne.Canvas) *PopUp
type PopUpMenu struct{ ... }
    func NewPopUpMenu(menu *fyne.Menu, c fyne.Canvas) *PopUpMenu
type ProgressBar struct{ ... }
    func NewProgressBar() *ProgressBar
    func NewProgressBarWithData(data binding.Float) *ProgressBar
type ProgressBarInfinite struct{ ... }
    func NewProgressBarInfinite() *ProgressBarInfinite
type RadioGroup struct{ ... }
    func NewRadioGroup(options []string, changed func(string)) *RadioGroup
type RichText struct{ ... }
    func NewRichText(segments ...RichTextSegment) *RichText
    func NewRichTextFromMarkdown(content string) *RichText
    func NewRichTextWithText(text string) *RichText
type RichTextBlock interface{ ... }
type RichTextSegment interface{ ... }
type RichTextStyle struct{ ... }
type Select struct{ ... }
    func NewSelect(options []string, changed func(string)) *Select
    func NewSelectWithData(options []string, data binding.String) *Select
type SelectEntry struct{ ... }
    func NewSelectEntry(options []string) *SelectEntry
type Separator struct{ ... }
    func NewSeparator() *Separator
type SeparatorSegment struct{ ... }
type Slider struct{ ... }
    func NewSlider(min, max float64) *Slider
    func NewSliderWithData(min, max float64, data binding.Float) *Slider
type Table struct{ ... }
    func NewTable(length func() (rows int, cols int), create func() fyne.CanvasObject, ...) *Table
    func NewTableWithHeaders(length func() (rows int, cols int), create func() fyne.CanvasObject, ...) *Table
type TableCellID struct{ ... }
type TextGrid struct{ ... }
    func NewTextGrid() *TextGrid
    func NewTextGridFromString(content string) *TextGrid
type TextGridCell struct{ ... }
type TextGridRow struct{ ... }
type TextGridStyle interface{ ... }
    var TextGridStyleDefault TextGridStyle ...
type TextSegment struct{ ... }
type Toolbar struct{ ... }
    func NewToolbar(items ...ToolbarItem) *Toolbar
type ToolbarAction struct{ ... }
    func NewToolbarAction(icon fyne.Resource, onActivated func()) *ToolbarAction
type ToolbarItem interface{ ... }
type ToolbarSeparator struct{}
    func NewToolbarSeparator() *ToolbarSeparator
type ToolbarSpacer struct{}
    func NewToolbarSpacer() *ToolbarSpacer
type Tree struct{ ... }
    func NewTree(childUIDs func(TreeNodeID) []TreeNodeID, isBranch func(TreeNodeID) bool, ...) *Tree
    func NewTreeWithData(data binding.DataTree, createItem func(bool) fyne.CanvasObject, ...) *Tree
    func NewTreeWithStrings(data map[string][]string) (t *Tree)
type TreeNodeID = string
```

## fyne.io/fyne/v2/dialog

```text
func ShowColorPicker(title, message string, callback func(c color.Color), parent fyne.Window)
func ShowConfirm(title, message string, callback func(bool), parent fyne.Window)
func ShowCustom(title, dismiss string, content fyne.CanvasObject, parent fyne.Window)
func ShowCustomConfirm(title, confirm, dismiss string, content fyne.CanvasObject, callback func(bool), ...)
func ShowCustomWithoutButtons(title string, content fyne.CanvasObject, parent fyne.Window)
func ShowEntryDialog(title, message string, onConfirm func(string), parent fyne.Window)
func ShowError(err error, parent fyne.Window)
func ShowFileOpen(callback func(reader fyne.URIReadCloser, err error), parent fyne.Window)
func ShowFileSave(callback func(writer fyne.URIWriteCloser, err error), parent fyne.Window)
func ShowFolderOpen(callback func(fyne.ListableURI, error), parent fyne.Window)
func ShowForm(title, confirm, dismiss string, content []*widget.FormItem, ...)
func ShowInformation(title, message string, parent fyne.Window)
type ColorPickerDialog struct{ ... }
    func NewColorPicker(title, message string, callback func(c color.Color), parent fyne.Window) *ColorPickerDialog
type ConfirmDialog struct{ ... }
    func NewConfirm(title, message string, callback func(bool), parent fyne.Window) *ConfirmDialog
    func NewCustomConfirm(title, confirm, dismiss string, content fyne.CanvasObject, callback func(bool), ...) *ConfirmDialog
type CustomDialog struct{ ... }
    func NewCustom(title, dismiss string, content fyne.CanvasObject, parent fyne.Window) *CustomDialog
    func NewCustomWithoutButtons(title string, content fyne.CanvasObject, parent fyne.Window) *CustomDialog
type Dialog interface{ ... }
    func NewError(err error, parent fyne.Window) Dialog
    func NewInformation(title, message string, parent fyne.Window) Dialog
type EntryDialog struct{ ... }
    func NewEntryDialog(title, message string, onConfirm func(string), parent fyne.Window) *EntryDialog
type FileDialog struct{ ... }
    func NewFileOpen(callback func(reader fyne.URIReadCloser, err error), parent fyne.Window) *FileDialog
    func NewFileSave(callback func(writer fyne.URIWriteCloser, err error), parent fyne.Window) *FileDialog
    func NewFolderOpen(callback func(fyne.ListableURI, error), parent fyne.Window) *FileDialog
type FormDialog struct{ ... }
    func NewForm(title, confirm, dismiss string, items []*widget.FormItem, callback func(bool), ...) *FormDialog
type ProgressDialog struct{ ... }
    func NewProgress(title, message string, parent fyne.Window) *ProgressDialog
type ProgressInfiniteDialog struct{ ... }
    func NewProgressInfinite(title, message string, parent fyne.Window) *ProgressInfiniteDialog
type ViewLayout int
    const ListView ViewLayout ...
```

## fyne.io/fyne/v2/layout

```text
func NewAdaptiveGridLayout(rowcols int) fyne.Layout
func NewBorderLayout(top, bottom, left, right fyne.CanvasObject) fyne.Layout
func NewCenterLayout() fyne.Layout
func NewCustomPaddedHBoxLayout(padding float32) fyne.Layout
func NewCustomPaddedLayout(padTop, padBottom, padLeft, padRight float32) fyne.Layout
func NewCustomPaddedVBoxLayout(padding float32) fyne.Layout
func NewFormLayout() fyne.Layout
func NewGridLayout(cols int) fyne.Layout
func NewGridLayoutWithColumns(cols int) fyne.Layout
func NewGridLayoutWithRows(rows int) fyne.Layout
func NewGridWrapLayout(size fyne.Size) fyne.Layout
func NewHBoxLayout() fyne.Layout
func NewMaxLayout() fyne.Layout
func NewPaddedLayout() fyne.Layout
func NewRowWrapLayout() fyne.Layout
func NewRowWrapLayoutWithCustomPadding(horizontal, vertical float32) fyne.Layout
func NewSpacer() fyne.CanvasObject
func NewStackLayout() fyne.Layout
func NewVBoxLayout() fyne.Layout
type CustomPaddedLayout struct{ ... }
type Spacer struct{ ... }
type SpacerObject interface{ ... }
```

## fyne.io/fyne/v2/data/binding

```text
const DataTreeRootID = ""
type Bool = Item[bool]
    func And(data ...Bool) Bool
    func BindPreferenceBool(key string, p fyne.Preferences) Bool
    func NewBool() Bool
    func Not(data Bool) Bool
    func Or(data ...Bool) Bool
    func StringToBool(str String) Bool
    func StringToBoolWithFormat(str String, format string) Bool
type BoolList = List[bool]
    func BindPreferenceBoolList(key string, p fyne.Preferences) BoolList
type BoolTree = Tree[bool]
type Bytes = Item[[]byte]
    func NewBytes() Bytes
type BytesList = List[[]byte]
type BytesTree = Tree[[]byte]
type DataItem interface{ ... }
type DataList interface{ ... }
type DataListener interface{ ... }
    func NewDataListener(fn func()) DataListener
type DataMap interface{ ... }
type DataTree interface{ ... }
type ExternalBool = ExternalItem[bool]
    func BindBool(v *bool) ExternalBool
type ExternalBoolList = ExternalList[bool]
type ExternalBoolTree = ExternalTree[bool]
type ExternalBytes = ExternalItem[[]byte]
    func BindBytes(v *[]byte) ExternalBytes
type ExternalBytesList = ExternalList[[]byte]
type ExternalBytesTree = ExternalTree[[]byte]
type ExternalFloat = ExternalItem[float64]
    func BindFloat(v *float64) ExternalFloat
type ExternalFloatList = ExternalList[float64]
type ExternalFloatTree = ExternalTree[float64]
type ExternalInt = ExternalItem[int]
    func BindInt(v *int) ExternalInt
type ExternalIntList = ExternalList[int]
type ExternalIntTree = ExternalTree[int]
type ExternalItem[T any] interface{ ... }
    func BindItem[T any](val *T, comparator func(T, T) bool) ExternalItem[T]
type ExternalList[T any] interface{ ... }
    func BindBoolList(v *[]bool) ExternalList[bool]
    func BindBytesList(v *[][]byte) ExternalList[[]byte]
    func BindFloatList(v *[]float64) ExternalList[float64]
    func BindIntList(v *[]int) ExternalList[int]
    func BindList[T any](v *[]T, comparator func(T, T) bool) ExternalList[T]
    func BindRuneList(v *[]rune) ExternalList[rune]
    func BindStringList(v *[]string) ExternalList[string]
    func BindURIList(v *[]fyne.URI) ExternalList[fyne.URI]
    func BindUntypedList(v *[]any) ExternalList[any]
type ExternalRune = ExternalItem[rune]
    func BindRune(v *rune) ExternalRune
type ExternalRuneList = ExternalList[rune]
type ExternalRuneTree = ExternalTree[rune]
type ExternalString = ExternalItem[string]
    func BindString(v *string) ExternalString
type ExternalStringList = ExternalList[string]
type ExternalStringTree = ExternalTree[string]
type ExternalTree[T any] interface{ ... }
    func BindBoolTree(ids *map[string][]string, v *map[string]bool) ExternalTree[bool]
    func BindBytesTree(ids *map[string][]string, v *map[string][]byte) ExternalTree[[]byte]
    func BindFloatTree(ids *map[string][]string, v *map[string]float64) ExternalTree[float64]
    func BindIntTree(ids *map[string][]string, v *map[string]int) ExternalTree[int]
    func BindRuneTree(ids *map[string][]string, v *map[string]rune) ExternalTree[rune]
    func BindStringTree(ids *map[string][]string, v *map[string]string) ExternalTree[string]
    func BindTree[T any](ids *map[string][]string, v *map[string]T, comparator func(T, T) bool) ExternalTree[T]
    func BindURITree(ids *map[string][]string, v *map[string]fyne.URI) ExternalTree[fyne.URI]
    func BindUntypedTree(ids *map[string][]string, v *map[string]any) ExternalTree[any]
type ExternalURI = ExternalItem[fyne.URI]
    func BindURI(v *fyne.URI) ExternalURI
type ExternalURIList = ExternalList[fyne.URI]
type ExternalURITree = ExternalTree[fyne.URI]
type ExternalUntyped = ExternalItem[any]
    func BindUntyped(v any) ExternalUntyped
type ExternalUntypedList = ExternalList[any]
type ExternalUntypedMap interface{ ... }
    func BindUntypedMap(d *map[string]any) ExternalUntypedMap
type ExternalUntypedTree = ExternalTree[any]
type Float = Item[float64]
    func BindPreferenceFloat(key string, p fyne.Preferences) Float
    func IntToFloat(val Int) Float
    func NewFloat() Float
    func StringToFloat(str String) Float
    func StringToFloatWithFormat(str String, format string) Float
type FloatList = List[float64]
    func BindPreferenceFloatList(key string, p fyne.Preferences) FloatList
type FloatTree = Tree[float64]
type Int = Item[int]
    func BindPreferenceInt(key string, p fyne.Preferences) Int
    func FloatToInt(v Float) Int
    func NewInt() Int
    func StringToInt(str String) Int
    func StringToIntWithFormat(str String, format string) Int
type IntList = List[int]
    func BindPreferenceIntList(key string, p fyne.Preferences) IntList
type IntTree = Tree[int]
type Item[T any] interface{ ... }
    func NewItem[T any](comparator func(T, T) bool) Item[T]
type List[T any] interface{ ... }
    func NewBoolList() List[bool]
    func NewBytesList() List[[]byte]
    func NewFloatList() List[float64]
    func NewIntList() List[int]
    func NewList[T any](comparator func(T, T) bool) List[T]
    func NewRuneList() List[rune]
    func NewStringList() List[string]
    func NewURIList() List[fyne.URI]
    func NewUntypedList() List[any]
type Rune = Item[rune]
    func NewRune() Rune
type RuneList = List[rune]
type RuneTree = Tree[rune]
type String = Item[string]
    func BindPreferenceString(key string, p fyne.Preferences) String
    func BoolToString(v Bool) String
    func BoolToStringWithFormat(v Bool, format string) String
    func FloatToString(v Float) String
    func FloatToStringWithFormat(v Float, format string) String
    func IntToString(v Int) String
    func IntToStringWithFormat(v Int, format string) String
    func NewSprintf(format string, b ...DataItem) String
    func NewString() String
    func StringToStringWithFormat(str String, format string) String
    func URIToString(v URI) String
type StringList = List[string]
    func BindPreferenceStringList(key string, p fyne.Preferences) StringList
type StringTree = Tree[string]
type Struct interface{ ... }
    func BindStruct(i any) Struct
type Tree[T any] interface{ ... }
    func NewBoolTree() Tree[bool]
    func NewBytesTree() Tree[[]byte]
    func NewFloatTree() Tree[float64]
    func NewIntTree() Tree[int]
    func NewRuneTree() Tree[rune]
    func NewStringTree() Tree[string]
    func NewTree[T any](comparator func(T, T) bool) Tree[T]
    func NewURITree() Tree[fyne.URI]
    func NewUntypedTree() Tree[any]
type URI = Item[fyne.URI]
    func NewURI() URI
    func StringToURI(str String) URI
type URIList = List[fyne.URI]
type URITree = Tree[fyne.URI]
type Untyped = Item[any]
    func NewUntyped() Untyped
type UntypedList = List[any]
type UntypedMap interface{ ... }
    func NewUntypedMap() UntypedMap
type UntypedTree = Tree[any]
```

## fyne.io/fyne/v2/data/validation

```text
func NewAllStrings(validators ...fyne.StringValidator) fyne.StringValidator
func NewRegexp(regexpstr, reason string) fyne.StringValidator
func NewTime(format string) fyne.StringValidator
```

## fyne.io/fyne/v2/storage

```text
var ErrAlreadyExists = errors.New("document already exists") ...
var URIRootError = repository.ErrURIRoot
func Appender(u fyne.URI) (fyne.URIWriteCloser, error)
func CanList(u fyne.URI) (bool, error)
func CanRead(u fyne.URI) (bool, error)
func CanWrite(u fyne.URI) (bool, error)
func Child(u fyne.URI, component string) (fyne.URI, error)
func Copy(source fyne.URI, destination fyne.URI) error
func CreateListable(u fyne.URI) error
func Delete(u fyne.URI) error
func DeleteAll(u fyne.URI) error
func EqualURI(t1, t2 fyne.URI) bool
func Exists(u fyne.URI) (bool, error)
func List(u fyne.URI) ([]fyne.URI, error)
func ListerForURI(uri fyne.URI) (fyne.ListableURI, error)
func LoadResourceFromURI(u fyne.URI) (fyne.Resource, error)
func Move(source fyne.URI, destination fyne.URI) error
func NewFileURI(fpath string) fyne.URI
func NewURI(s string) fyne.URI
func OpenFileFromURI(uri fyne.URI) (fyne.URIReadCloser, error)
func Parent(u fyne.URI) (fyne.URI, error)
func ParseURI(s string) (fyne.URI, error)
func Reader(u fyne.URI) (fyne.URIReadCloser, error)
func SaveFileToURI(uri fyne.URI) (fyne.URIWriteCloser, error)
func Writer(u fyne.URI) (fyne.URIWriteCloser, error)
type ExtensionFileFilter struct{ ... }
type FileFilter interface{ ... }
    func NewExtensionFileFilter(extensions []string) FileFilter
    func NewMimeTypeFileFilter(mimeTypes []string) FileFilter
type MimeTypeFileFilter struct{ ... }
```

## fyne.io/fyne/v2/theme

```text
const ColorRed = internaltheme.ColorRed ...
const IconNameCancel fyne.ThemeIconName = "cancel" ...
const SizeNameCaptionText fyne.ThemeSizeName = "helperText" ...
const VariantDark = internaltheme.VariantDark ...
func AccountIcon() fyne.Resource
func BackgroundColor() color.Color
func BrokenImageIcon() fyne.Resource
func ButtonColor() color.Color
func CalendarIcon() fyne.Resource
func CancelIcon() fyne.Resource
func CaptionTextSize() float32
func CheckButtonCheckedIcon() fyne.Resource
func CheckButtonFillIcon() fyne.Resource
func CheckButtonIcon() fyne.Resource
func Color(name fyne.ThemeColorName) color.Color
func ColorAchromaticIcon() fyne.Resource
func ColorChromaticIcon() fyne.Resource
func ColorForWidget(name fyne.ThemeColorName, w fyne.Widget) color.Color
func ColorPaletteIcon() fyne.Resource
func ComputerIcon() fyne.Resource
func ConfirmIcon() fyne.Resource
func ContentAddIcon() fyne.Resource
func ContentClearIcon() fyne.Resource
func ContentCopyIcon() fyne.Resource
func ContentCutIcon() fyne.Resource
func ContentPasteIcon() fyne.Resource
func ContentRedoIcon() fyne.Resource
func ContentRemoveIcon() fyne.Resource
func ContentUndoIcon() fyne.Resource
func Current() fyne.Theme
func CurrentForWidget(w fyne.CanvasObject) fyne.Theme
func DarkTheme() fyne.Theme
func DefaultEmojiFont() fyne.Resource
func DefaultSymbolFont() fyne.Resource
func DefaultTextBoldFont() fyne.Resource
func DefaultTextBoldItalicFont() fyne.Resource
func DefaultTextFont() fyne.Resource
func DefaultTextItalicFont() fyne.Resource
func DefaultTextMonospaceFont() fyne.Resource
func DefaultTheme() fyne.Theme
func DeleteIcon() fyne.Resource
func DesktopIcon() fyne.Resource
func DisabledButtonColor() color.Color
func DisabledColor() color.Color
func DisabledTextColor() color.Color
func DocumentCreateIcon() fyne.Resource
func DocumentIcon() fyne.Resource
func DocumentPrintIcon() fyne.Resource
func DocumentSaveIcon() fyne.Resource
func DownloadIcon() fyne.Resource
func ErrorColor() color.Color
func ErrorIcon() fyne.Resource
func FileApplicationIcon() fyne.Resource
func FileAudioIcon() fyne.Resource
func FileIcon() fyne.Resource
func FileImageIcon() fyne.Resource
func FileTextIcon() fyne.Resource
func FileVideoIcon() fyne.Resource
func FocusColor() color.Color
func FolderIcon() fyne.Resource
func FolderNewIcon() fyne.Resource
func FolderOpenIcon() fyne.Resource
func Font(style fyne.TextStyle) fyne.Resource
func ForegroundColor() color.Color
func FromJSON(data string) (fyne.Theme, error)
func FromJSONReader(r io.Reader) (fyne.Theme, error)
func FromJSONReaderWithFallback(r io.Reader, fallback fyne.Theme) (fyne.Theme, error)
func FromJSONWithFallback(data string, fallback fyne.Theme) (fyne.Theme, error)
func FromLegacy(t fyne.LegacyTheme) fyne.Theme
func FyneLogo() fyne.Resource
func GridIcon() fyne.Resource
func HeaderBackgroundColor() color.Color
func HelpIcon() fyne.Resource
func HistoryIcon() fyne.Resource
func HomeIcon() fyne.Resource
func HoverColor() color.Color
func HyperlinkColor() color.Color
func Icon(name fyne.ThemeIconName) fyne.Resource
func IconForWidget(name fyne.ThemeIconName, w fyne.Widget) fyne.Resource
func IconInlineSize() float32
func InfoIcon() fyne.Resource
func InnerPadding() float32
func InputBackgroundColor() color.Color
func InputBorderColor() color.Color
func InputBorderSize() float32
func InputRadiusSize() float32
func LightTheme() fyne.Theme
func LineSpacing() float32
func ListIcon() fyne.Resource
func LoginIcon() fyne.Resource
func LogoutIcon() fyne.Resource
func MailAttachmentIcon() fyne.Resource
func MailComposeIcon() fyne.Resource
func MailForwardIcon() fyne.Resource
func MailReplyAllIcon() fyne.Resource
func MailReplyIcon() fyne.Resource
func MailSendIcon() fyne.Resource
func MediaFastForwardIcon() fyne.Resource
func MediaFastRewindIcon() fyne.Resource
func MediaMusicIcon() fyne.Resource
func MediaPauseIcon() fyne.Resource
func MediaPhotoIcon() fyne.Resource
func MediaPlayIcon() fyne.Resource
func MediaRecordIcon() fyne.Resource
func MediaReplayIcon() fyne.Resource
func MediaSkipNextIcon() fyne.Resource
func MediaSkipPreviousIcon() fyne.Resource
func MediaStopIcon() fyne.Resource
func MediaVideoIcon() fyne.Resource
func MenuBackgroundColor() color.Color
func MenuDropDownIcon() fyne.Resource
func MenuDropUpIcon() fyne.Resource
func MenuExpandIcon() fyne.Resource
func MenuIcon() fyne.Resource
func MoreHorizontalIcon() fyne.Resource
func MoreVerticalIcon() fyne.Resource
func MoveDownIcon() fyne.Resource
func MoveUpIcon() fyne.Resource
func NavigateBackIcon() fyne.Resource
func NavigateNextIcon() fyne.Resource
func OverlayBackgroundColor() color.Color
func Padding() float32
func PlaceHolderColor() color.Color
func PressedColor() color.Color
func PrimaryColor() color.Color
func PrimaryColorNamed(name string) color.Color
func PrimaryColorNames() []string
func QuestionIcon() fyne.Resource
func RadioButtonCheckedIcon() fyne.Resource
func RadioButtonFillIcon() fyne.Resource
func RadioButtonIcon() fyne.Resource
func ScrollBarColor() color.Color
func ScrollBarSize() float32
func ScrollBarSmallSize() float32
func SearchIcon() fyne.Resource
func SearchReplaceIcon() fyne.Resource
func SelectionColor() color.Color
func SelectionRadiusSize() float32
func SeparatorColor() color.Color
func SeparatorThicknessSize() float32
func SettingsIcon() fyne.Resource
func ShadowColor() color.Color
func Size(name fyne.ThemeSizeName) float32
func SizeForWidget(name fyne.ThemeSizeName, w fyne.Widget) float32
func StorageIcon() fyne.Resource
func SuccessColor() color.Color
func SymbolFont() fyne.Resource
func TextBoldFont() fyne.Resource
func TextBoldItalicFont() fyne.Resource
func TextColor() color.Color
func TextFont() fyne.Resource
func TextHeadingSize() float32
func TextItalicFont() fyne.Resource
func TextMonospaceFont() fyne.Resource
func TextSize() float32
func TextSubHeadingSize() float32
func UploadIcon() fyne.Resource
func ViewFullScreenIcon() fyne.Resource
func ViewRefreshIcon() fyne.Resource
func ViewRestoreIcon() fyne.Resource
func VisibilityIcon() fyne.Resource
func VisibilityOffIcon() fyne.Resource
func VolumeDownIcon() fyne.Resource
func VolumeMuteIcon() fyne.Resource
func VolumeUpIcon() fyne.Resource
func WarningColor() color.Color
func WarningIcon() fyne.Resource
func WindowCloseIcon() fyne.Resource
func WindowMaximizeIcon() fyne.Resource
func WindowMinimizeIcon() fyne.Resource
func ZoomFitIcon() fyne.Resource
func ZoomInIcon() fyne.Resource
func ZoomOutIcon() fyne.Resource
type DisabledResource struct{ ... }
    func NewDisabledResource(res fyne.Resource) *DisabledResource
type ErrorThemedResource struct{ ... }
    func NewErrorThemedResource(orig fyne.Resource) *ErrorThemedResource
type InvertedThemedResource struct{ ... }
    func NewInvertedThemedResource(orig fyne.Resource) *InvertedThemedResource
type PrimaryThemedResource struct{ ... }
    func NewPrimaryThemedResource(orig fyne.Resource) *PrimaryThemedResource
type ThemedResource struct{ ... }
    func NewColoredResource(src fyne.Resource, name fyne.ThemeColorName) *ThemedResource
    func NewSuccessThemedResource(src fyne.Resource) *ThemedResource
    func NewThemedResource(src fyne.Resource) *ThemedResource
    func NewWarningThemedResource(src fyne.Resource) *ThemedResource
```

## fyne.io/fyne/v2/test

```text
func ApplyTheme(t *testing.T, theme fyne.Theme)
func AssertAllColorNamesDefined(t *testing.T, th fyne.Theme, themeName string)
func AssertCanvasTappableAt(t *testing.T, c fyne.Canvas, pos fyne.Position) bool
func AssertImageMatches(t *testing.T, masterFilename string, img image.Image, msgAndArgs ...any) bool
func AssertNotificationSent(t *testing.T, n *fyne.Notification, f func())
func AssertObjectRendersToImage(t *testing.T, masterFilename string, o fyne.CanvasObject, msgAndArgs ...any) bool
func AssertObjectRendersToMarkup(t *testing.T, masterFilename string, o fyne.CanvasObject, msgAndArgs ...any) bool
func AssertRendersToImage(t *testing.T, masterFilename string, c fyne.Canvas, msgAndArgs ...any) bool
func AssertRendersToMarkup(t *testing.T, masterFilename string, c fyne.Canvas, msgAndArgs ...any) bool
func Canvas() fyne.Canvas
func DoubleTap(obj fyne.DoubleTappable)
func Drag(c fyne.Canvas, pos fyne.Position, deltaX, deltaY float32)
func FocusNext(c fyne.Canvas)
func FocusPrevious(c fyne.Canvas)
func KnownThemeVariants() map[string]fyne.ThemeVariant
func LaidOutObjects(o fyne.CanvasObject) (objects []fyne.CanvasObject)
func MoveMouse(c fyne.Canvas, pos fyne.Position)
func NewApp() fyne.App
func NewClipboard() fyne.Clipboard
func NewDriver() fyne.Driver
func NewDriverWithPainter(painter SoftwarePainter) fyne.Driver
func NewTempApp(t testing.TB) fyne.App
func NewTempWindow(t testing.TB, content fyne.CanvasObject) fyne.Window
func NewTheme() fyne.Theme
func NewWindow(content fyne.CanvasObject) fyne.Window
func RenderObjectToMarkup(o fyne.CanvasObject) string
func RenderToMarkup(c fyne.Canvas) string
func Scroll(c fyne.Canvas, pos fyne.Position, deltaX, deltaY float32)
func Tap(obj fyne.Tappable)
func TapAt(obj fyne.Tappable, pos fyne.Position)
func TapCanvas(c fyne.Canvas, pos fyne.Position)
func TapSecondary(obj fyne.SecondaryTappable)
func TapSecondaryAt(obj fyne.SecondaryTappable, pos fyne.Position)
func TempWidgetRenderer(t *testing.T, wid fyne.Widget) fyne.WidgetRenderer
func Theme() fyne.Theme
func Type(obj fyne.Focusable, chars string)
func TypeOnCanvas(c fyne.Canvas, chars string)
func WidgetRenderer(wid fyne.Widget) fyne.WidgetRenderer
func WithTestTheme(t *testing.T, f func())
type SoftwarePainter interface{ ... }
type WindowlessCanvas interface{ ... }
    func NewCanvas() WindowlessCanvas
    func NewCanvasWithPainter(painter SoftwarePainter) WindowlessCanvas
    func NewTransparentCanvasWithPainter(painter SoftwarePainter) WindowlessCanvas
```

## fyne.io/fyne/v2/driver/desktop

```text
const KeyNone fyne.KeyName = "" ...
const ShiftModifier = fyne.KeyModifierShift ...
type App interface{ ... }
type Canvas interface{ ... }
type Cursor interface{ ... }
type Cursorable interface{ ... }
type CustomShortcut struct{ ... }
type Driver interface{ ... }
type Hoverable interface{ ... }
type Keyable interface{ ... }
type Modifier = fyne.KeyModifier
type MouseButton int
    const MouseButtonPrimary MouseButton = 1 << iota ...
type MouseEvent struct{ ... }
type Mouseable interface{ ... }
type StandardCursor int
    const DefaultCursor StandardCursor = iota ...
```
