# Module 6 EDA - GUI Mermaid Diagrams

> Auto-generated: 2026-01-30
> Source: Codebase analysis - module6-eda/pkg/gui/

## Quick Reference

| Diagram | Purpose | Description |
|---------|---------|-------------|
| Application Architecture | High-level entry points | Standalone vs Unified app modes |
| Component Hierarchy | Widget tree structure | Full layout and nesting |
| Builder Tab Layout | Configuration and generation | Cell config, array config, actions |
| Preview Tab Structure | Output visualization | Verilog, DEF, Layout images |
| Validation Tab Structure | Result display | Status indicators and logs |
| Generate All Flow | Sequence of generation steps | LEF/LIB/V/DEF/PNG creation |
| Validate All Flow | Validation sequence | Yosys, DEF, Cross-check, OpenLane |
| Threading Model | Goroutine architecture | Main thread + background workers |
| State Management | Data flow and callbacks | Entry → Stats → Updates |

## 1. Application Architecture

```mermaid
graph TD
    Start[Application Start] --> Mode{Deployment Mode?}
    Mode -->|Unified App| Unified["container.NewAppTabs<br/>embedded.go"]
    Mode -->|Standalone| Standalone["CreateMainWindow<br/>app.go"]

    Unified --> CreateContent["CreateModuleContent"]
    Standalone --> HeaderSetup["Header + View Selector"]

    CreateContent --> Tabs["AppTabs Container"]
    HeaderSetup --> ViewSelector["Stack-based View Selector"]

    Tabs --> Tab1["Tab 1: Builder & Validation"]
    Tabs --> Tab2["Tab 2: Learn"]

    ViewSelector --> VS1["Builder & Validation"]
    ViewSelector --> VS2["Learn"]

    Tab1 --> BuilderContent["MakeBuilderValidationTab"]
    Tab2 --> LearnContent["MakeLearnTab"]

    VS1 --> BuilderContent
    VS2 --> LearnContent

    style Start fill:#e1f5ff
    style Mode fill:#fff3e0
    style Unified fill:#f3e5f5
    style Standalone fill:#f3e5f5
    style BuilderContent fill:#fce4ec
    style LearnContent fill:#fce4ec
```

## 2. Builder & Validation Tab - Full Component Tree

```mermaid
graph TD
    BuilderTab["MakeBuilderValidationTab<br/>Main Container"] --> MainContent["MainContent<br/>BorderContainer"]

    MainContent --> TopSection["TopSection<br/>VBox"]
    MainContent --> MainSplit["MainSplit<br/>VSplit 75/25"]

    TopSection --> ConfigSplit["ConfigSplit<br/>HSplit 45/55"]
    TopSection --> ActionRow["ActionRow<br/>HBox"]
    TopSection --> StatusRow["StatusRow<br/>HBox"]

    ConfigSplit --> CellPanel["Cell Config Panel<br/>VBox"]
    ConfigSplit --> ArrayPanel["Array Config Panel<br/>VBox"]

    CellPanel --> CellGrid1["GridWithColumns 6<br/>Name, W, H, Rise, Fall, Cap"]
    CellPanel --> CellGrid2["GridWithColumns 4<br/>Leakage, CellArea"]

    ArrayPanel --> ArrayGrid["GridWithColumns 6<br/>Rows, Cols, Mode"]
    ArrayPanel --> ArchToggle["GridWithColumns 3<br/>PASSIVE | 1T1R | 2T1R"]
    ArrayPanel --> ModeHelp["Mode Help Text"]
    ArrayPanel --> StatsRow["HBox Stats"]

    StatsRow --> Stats["Total, Area, WL, BL, Density, Util"]

    ActionRow --> GenBtn["Generate All"]
    ActionRow --> ValBtn["Validate All"]
    ActionRow --> ExpBtn["Export Package"]

    StatusRow --> Status["Status Label"]

    MainSplit --> PreviewTabs["Preview AppTabs"]
    MainSplit --> ValidationSection["Validation VBox"]

    PreviewTabs --> VTab["Verilog Tab"]
    PreviewTabs --> DTab["DEF Tab"]
    PreviewTabs --> LTab["Layout Tab"]

    VTab --> VerilogPreview["MultiLineEntry Preview"]
    VTab --> VerilogStats["Stats Label"]

    DTab --> DEFPreview["MultiLineEntry Preview"]
    DTab --> DEFStats["Stats Label"]

    LTab --> LayoutHelp["Help Text + Buttons"]
    LTab --> LayoutStack["GridWithColumns 3<br/>KLayout, OpenROAD, Yosys"]

    ValidationSection --> ValidationRow["HBox Results"]
    ValidationSection --> OpenLaneRow["HBox Status"]
    ValidationSection --> LogHeader["Log Header + Clear"]
    ValidationSection --> LogScroll["Log Output"]

    ValidationRow --> YosysResult["Label Result"]
    ValidationRow --> DEFResult["Label Result"]
    ValidationRow --> CrossResult["Label Result"]
    ValidationRow --> PlacementResult["Label Result"]

    OpenLaneRow --> DockerStatus["Docker Status"]
    OpenLaneRow --> PDKStatus["PDK Status"]
    OpenLaneRow --> PullImageBtn["Pull Button"]

    style BuilderTab fill:#b3e5fc
    style CellPanel fill:#fff9c4
    style ArrayPanel fill:#fff9c4
    style PreviewTabs fill:#e1bee7
    style ValidationSection fill:#ffccbc
```

## 3. Builder Tab - Cell Configuration Panel

```mermaid
graph LR
    CellPanel["Cell Config Panel"] --> Inputs["Input Fields"]
    Inputs --> Name["Name: fecim_bitcell"]
    Inputs --> Width["Width: 0.460 µm"]
    Inputs --> Height["Height: 2.720 µm"]
    Inputs --> Rise["Rise: 0.1 ns"]
    Inputs --> Fall["Fall: 0.1 ns"]
    Inputs --> Cap["Cap: 0.002 pF"]
    Inputs --> Leak["Leak: 0.001 nW"]

    CellPanel --> Display["Display"]
    Display --> CellArea["Cell Area Label"]

    all_inputs -->|getCellConfig| CellStruct["CellConfig Struct"]

    CellStruct --> CellCfg["Name, Width, Height<br/>RiseTime, FallTime<br/>InputCap, LeakagePower"]

    style CellPanel fill:#fff9c4
    style Inputs fill:#fffde7
    style CellStruct fill:#f0f4c3
```

## 4. Builder Tab - Array Configuration Panel

```mermaid
graph TD
    ArrayPanel["Array Config Panel"] --> Entries["Entry Fields"]
    Entries --> RowsEntry["Rows: (0-999)"]
    Entries --> ColsEntry["Cols: (0-999)"]
    Entries --> ModeSelect["Mode Selector"]

    ModeSelect --> StorageMode["storage: Non-volatile retention"]
    ModeSelect --> MemoryMode["memory: Fast DRAM-like access"]
    ModeSelect --> ComputeMode["compute: Matrix-vector multiply"]

    ArrayPanel --> ArchButtons["Architecture Toggle"]
    ArchButtons --> PassiveBtn["PASSIVE<br/>0.46 x 2.72 µm"]
    ArchButtons --> OneToneRBtn["1T1R<br/>0.92 x 2.72 µm"]
    ArchButtons --> TwoToneRBtn["2T1R<br/>1.38 x 3.40 µm"]

    ArrayPanel --> StatsDisplay["Statistics Row"]
    StatsDisplay --> Total["Total Cells"]
    StatsDisplay --> Area["Array Area µm²"]
    StatsDisplay --> WL["WL Length µm"]
    StatsDisplay --> BL["BL Length µm"]
    StatsDisplay --> Density["Density cells/µm²"]
    StatsDisplay --> Utilization["Utilization %"]

    style ArrayPanel fill:#fff9c4
    style ArchButtons fill:#ffe0b2
    style StatsDisplay fill:#e0f2f1
```

## 5. Architecture Button State Machine

```mermaid
graph TD
    Start["Initial State"] --> CheckArch{Current Arch?}

    CheckArch -->|passive| PassiveSelected["PASSIVE Button<br/>Importance=High"]
    CheckArch -->|1t1r| OneToneRSelected["1T1R Button<br/>Importance=High"]
    CheckArch -->|2t1r| TwoToneRSelected["2T1R Button<br/>Importance=High"]

    PassiveSelected --> OnTapPassive["Passive OnTapped"]
    OneToneRSelected --> OnTap1T1R["1T1R OnTapped"]
    TwoToneRSelected --> OnTap2T1R["2T1R OnTapped"]

    OnTapPassive --> CheckPassive{Already Passive?}
    OnTap1T1R --> Check1T1R{Already 1T1R?}
    OnTap2T1R --> Check2T1R{Already 2T1R?}

    CheckPassive -->|Yes| Return1["Return (no change)"]
    CheckPassive -->|No| SetPassive["Set Architecture=passive"]
    SetPassive --> SetPassiveDims["Width=0.460, Height=2.720"]

    Check1T1R -->|Yes| Return2["Return (no change)"]
    Check1T1R -->|No| Set1T1R["Set Architecture=1t1r"]
    Set1T1R --> Set1T1RDims["Width=0.920, Height=3.400"]

    Check2T1R -->|Yes| Return3["Return (no change)"]
    Check2T1R -->|No| Set2T1R["Set Architecture=2t1r"]
    Set2T1R --> Set2T1RDims["Width=1.380, Height=3.400"]

    SetPassiveDims --> UpdateButtons["updateArchButtons()"]
    Set1T1RDims --> UpdateButtons
    Set2T1RDims --> UpdateButtons

    UpdateButtons --> RefreshDisplay["Update UI + updateStats()"]
    RefreshDisplay --> ReloadImage["updateLayoutImage()"]

    style PassiveSelected fill:#fff9c4
    style OneToneRSelected fill:#fff9c4
    style TwoToneRSelected fill:#fff9c4
    style RefreshDisplay fill:#c8e6c9
```

## 6. Preview Tabs - Three-Column Layout

```mermaid
graph TD
    PreviewTabs["Preview AppTabs Container"] --> VerilogTab["Tab: Verilog"]
    PreviewTabs --> DEFTab["Tab: DEF"]
    PreviewTabs --> LayoutTab["Tab: Layout"]

    VerilogTab --> VHeader["Header: VerilogStatsLabel<br/>Instances, Lines, Size KB"]
    VerilogTab --> VScroll["Scroll Container"]
    VScroll --> VPreview["MultiLineEntry<br/>Full Verilog content<br/>Monospace, ReadOnly"]

    DEFTab --> DHeader["Header: DEFStatsLabel<br/>Components, Filename"]
    DEFTab --> DScroll["Scroll Container"]
    DScroll --> DPreview["MultiLineEntry<br/>Full DEF content<br/>Monospace, ReadOnly"]

    LayoutTab --> LHeader["Header: HBox"]
    LHeader --> GenSchematicBtn["Gen Schematic (Yosys)"]
    LHeader --> GenLayoutBtn["Gen Layout (OpenROAD)"]
    LHeader --> LayoutHelp["Help Text"]

    LayoutTab --> LayoutStack["GridWithColumns 3"]
    LayoutStack --> KLayoutCard["KLayout Card<br/>Image + Status"]
    LayoutStack --> OpenROADCard["OpenROAD Card<br/>Image + Status"]
    LayoutStack --> YosysCard["Yosys Card<br/>Image + Status"]

    KLayoutCard --> KImage["Image Display<br/>fecim_crossbar_NxN.png"]
    KLayoutCard --> KStatus["Status Label"]

    OpenROADCard --> OImage["Image Display<br/>fecim_crossbar_NxN_openroad.png"]
    OpenROADCard --> OStatus["Status Label"]

    YosysCard --> YImage["Image Display<br/>fecim_crossbar_NxN_schematic.png"]
    YosysCard --> YStatus["Status Label"]

    style PreviewTabs fill:#e1bee7
    style VerilogTab fill:#f3e5f5
    style DEFTab fill:#f3e5f5
    style LayoutTab fill:#f3e5f5
    style KLayoutCard fill:#fff9c4
    style OpenROADCard fill:#fff9c4
    style YosysCard fill:#fff9c4
```

## 7. Generate All - Data Flow Sequence

```mermaid
sequenceDiagram
    participant User as User
    participant UI as UI Layer
    participant Goroutine as Generate Goroutine
    participant Export as pkg/export
    participant FileSystem as File System
    participant Docker as OpenLane/Docker

    User->>UI: Click "Generate All"
    UI->>UI: Disable buttons (Generate, Validate, Export)
    UI->>UI: Set status "Generating..."
    UI->>UI: Clear log output

    UI->>Goroutine: Launch goroutine

    Goroutine->>Goroutine: updateStats() - sync rows/cols
    Goroutine->>Goroutine: getCellConfig() - parse inputs

    rect rgb(200,220,255)
    Note over Goroutine,Export: Step 1: Cell Library
    Goroutine->>Export: GenerateLEF(cellCfg)
    Export-->>Goroutine: LEF content string
    Goroutine->>FileSystem: Write cells/fecim_*bitcell/*.lef
    Goroutine->>Export: GenerateLiberty(cellCfg)
    Export-->>Goroutine: Liberty content string
    Goroutine->>FileSystem: Write cells/fecim_*bitcell/*.lib
    Goroutine->>Export: GenerateCellVerilog(cellCfg)
    Export-->>Goroutine: Cell V content string
    Goroutine->>FileSystem: Write cells/fecim_*bitcell/*.v
    end

    rect rgb(220,200,255)
    Note over Goroutine,Export: Step 2: Array Verilog
    Goroutine->>Export: GenerateArrayVerilog(cfg)
    Export-->>Goroutine: Array Verilog content
    Goroutine->>UI: fyne.Do() Update verilogPreview
    Goroutine->>FileSystem: Write data/fecim_crossbar_NxN.v
    end

    rect rgb(200,255,220)
    Note over Goroutine,FileSystem: Step 3: DEF Placement
    Goroutine->>Goroutine: generateBuilderDEF(cfg)
    Goroutine->>UI: fyne.Do() Update defPreview
    Goroutine->>FileSystem: Write data/fecim_crossbar_NxN.def
    end

    rect rgb(255,240,200)
    Note over Goroutine,Docker: Step 4: Layout Image (KLayout)
    Goroutine->>Docker: IsKLayoutAvailable()?
    alt KLayout available
        Goroutine->>Docker: GenerateLayoutImage(def, lef, png)
        Docker-->>Goroutine: Result with PNG path
        Goroutine->>UI: fyne.Do() updateLayoutImage()
    else KLayout not available
        Goroutine->>UI: fyne.Do() Set status "Need Docker"
    end
    end

    rect rgb(200,200,220)
    Note over Goroutine,Export: Step 5: OpenLane Config
    Goroutine->>Export: GenerateOpenLaneConfig(cfg)
    Export-->>Goroutine: Config JSON string
    Goroutine->>FileSystem: Write data/config.json
    end

    Goroutine->>UI: fyne.Do() Enable buttons
    Goroutine->>UI: fyne.Do() Set status "All files generated"
    UI->>User: Display success
```

## 8. Validate All - Validation Sequence

```mermaid
sequenceDiagram
    participant User as User
    participant UI as UI Layer
    participant Goroutine as Validate Goroutine
    participant Validation as pkg/validation
    participant Tools as EDA Tools

    User->>UI: Click "Validate All"
    UI->>UI: Disable Generate/Export buttons
    UI->>UI: Set status "Validating..."
    UI->>UI: Clear log, set results to "..."

    UI->>Goroutine: Launch goroutine

    rect rgb(100,150,200)
    Note over Goroutine,Tools: Yosys Verilog Validation
    Goroutine->>Validation: ValidateVerilogWithCell(arrayPath, cellPath)
    Validation->>Tools: yosys -p "read_verilog cell.v array.v"
    Tools-->>Validation: Exit code + output
    Validation-->>Goroutine: error or nil
    alt No error
        Goroutine->>UI: fyne.Do() yosysResult = "✓ PASS"
    else Error
        Goroutine->>UI: fyne.Do() yosysResult = "✗ FAIL"
        Goroutine->>UI: fyne.Do() Log error
    end
    end

    rect rgb(150,100,150)
    Note over Goroutine,Validation: DEF Syntax Validation
    Goroutine->>Validation: ValidateDEF(defPath)
    Validation-->>Goroutine: error or nil
    alt No error
        Goroutine->>UI: fyne.Do() defResult = "✓ PASS"
    else Error
        Goroutine->>UI: fyne.Do() defResult = "✗ FAIL"
    end
    end

    rect rgb(200,150,100)
    Note over Goroutine,Validation: Cross-Check Files
    Goroutine->>Validation: CrossCheckFiles(lef, lib, v)
    Validation-->>Goroutine: error or nil
    alt No error
        Goroutine->>UI: fyne.Do() crossResult = "✓ PASS"
    else Error
        Goroutine->>UI: fyne.Do() crossResult = "✗ FAIL"
    end
    end

    rect rgb(100,200,150)
    Note over Goroutine,Tools: OpenLane Placement Check
    Goroutine->>Validation: manager.DetectMode()
    Validation-->>Goroutine: ModeDocker or ModeNone
    alt ModeDocker or Native
        Goroutine->>Validation: RunPlacementCheckWithCell(def, lef, ...)
        Validation->>Tools: OpenROAD placement check
        Tools-->>Validation: Result
        alt Passed
            Goroutine->>UI: fyne.Do() placementResult = "✓ PASS"
        else Failed
            Goroutine->>UI: fyne.Do() placementResult = "✗ FAIL"
            Goroutine->>UI: fyne.Do() Log violations
        end
    else ModeNone
        Goroutine->>UI: fyne.Do() placementResult = "⊝ SKIP"
    end
    end

    alt All passed
        Goroutine->>UI: fyne.Do() validationSummary = "✓ All checks passed"
        Goroutine->>UI: fyne.Do() statusLabel = "All validations passed"
    else Some failed
        Goroutine->>UI: fyne.Do() validationSummary = "✗ Some checks failed"
        Goroutine->>UI: fyne.Do() statusLabel = "Some validations failed"
    end

    Goroutine->>UI: fyne.Do() Enable all buttons
```

## 9. Threading Model & fyne.Do Pattern

```mermaid
graph LR
    MainThread["Main UI Thread<br/>Fyne Event Loop"]

    MainThread -->|Button Click| G1["Goroutine 1<br/>Generate All"]
    MainThread -->|Button Click| G2["Goroutine 2<br/>Validate All"]
    MainThread -->|Button Click| G3["Goroutine 3<br/>Gen Schematic"]
    MainThread -->|Button Click| G4["Goroutine 4<br/>Gen Layout"]

    G1 -->|I/O + Computation| Work1["File I/O + Export"]
    G2 -->|External Tools| Work2["Yosys, OpenROAD"]
    G3 -->|External Tools| Work3["Graphviz, dot2png"]
    G4 -->|External Tools| Work4["KLayout, OpenROAD"]

    Work1 -->|fyne.Do| MainThread
    Work2 -->|fyne.Do| MainThread
    Work3 -->|fyne.Do| MainThread
    Work4 -->|fyne.Do| MainThread

    MainThread -->|UI Updates| Label["Update Labels"]
    MainThread -->|UI Updates| Image["Update Images"]
    MainThread -->|UI Updates| Log["Append Log"]

    Label --> Render["Render Frame"]
    Image --> Render
    Log --> Render

    style MainThread fill:#b3e5fc
    style G1 fill:#fff9c4
    style G2 fill:#fff9c4
    style G3 fill:#fff9c4
    style G4 fill:#fff9c4
    style Render fill:#c8e6c9
```

## 10. Entry Field Change Handlers

```mermaid
graph TD
    RowsEntry["rowsEntry.OnChanged"] -->|User types| UpdateStats1["updateStats()"]
    ColsEntry["colsEntry.OnChanged"] -->|User types| UpdateStats1
    WidthEntry["widthEntry.OnChanged"] -->|User types| UpdateStats1
    HeightEntry["heightEntry.OnChanged"] -->|User types| UpdateStats1

    UpdateStats1 --> ParseAll["Parse: rows, cols, width, height"]
    ParseAll --> ValidateNumbers["Validate numeric values"]
    ValidateNumbers --> CalcMetrics["Calculate metrics"]

    CalcMetrics --> CalcTotal["total = rows × cols"]
    CalcMetrics --> CalcArea["area = total × width × height"]
    CalcMetrics --> CalcWL["wlLength = cols × width"]
    CalcMetrics --> CalcBL["blLength = rows × height"]
    CalcMetrics --> CalcDensity["density = total / arrayArea"]
    CalcMetrics --> CalcUtil["utilization = (area / arrayArea) × 100"]

    CalcTotal --> UpdateUI["fyne.Do(func())"]
    CalcArea --> UpdateUI
    CalcWL --> UpdateUI
    CalcBL --> UpdateUI
    CalcDensity --> UpdateUI
    CalcUtil --> UpdateUI

    UpdateUI --> SetLabels["Set all labels"]
    SetLabels --> Refresh["Call Refresh()"]

    style UpdateStats1 fill:#c8e6c9
    style UpdateUI fill:#b3e5fc
    style Refresh fill:#ffe0b2
```

## 11. Learn Tab - Component Structure

```mermaid
graph TD
    LearnTab["MakeLearnTab"] --> Header["Header VBox"]
    LearnTab --> Content["Content BorderContainer"]

    Header --> Title["FeCIM Array Builder<br/>Learning Center"]
    Header --> Subtitle["Understanding OpenLane..."]
    Header --> Sep["Separator"]

    Content --> Split["HSplit 25/75"]

    Split --> Sidebar["Sidebar (25%)"]
    Split --> ContentScroll["Content Scroll (75%)"]

    Sidebar --> SidebarTitle["Topics (title)"]
    Sidebar --> TopicList["List Widget"]

    TopicList --> T0["1. What is FeCIM EDA?"]
    TopicList --> T1["2. The Crossbar Architecture"]
    TopicList --> T2["3. EDA Files We Generate"]

    ContentScroll --> DynamicContent["Dynamic Content<br/>VBox (changes on selection)"]

    T0 -->|OnSelected| IntroContent["Intro Content"]
    T1 -->|OnSelected| CrossbarContent["Crossbar Content"]
    T2 -->|OnSelected| FilesContent["Files Content"]

    IntroContent --> IntroTitle["What is FeCIM EDA?"]
    IntroContent --> IntroText["Explanation text"]
    IntroContent --> OperationModes["OperationModesVisual()"]
    IntroContent --> OpenLaneFlow["OpenLaneFlowDiagram()"]
    IntroContent --> StagesExplained["The Stages Explained"]
    IntroContent --> DoColumns["Do/Don't Lists"]
    IntroContent --> DisclaimerCard["Disclaimer Banner"]

    CrossbarContent --> PassiveSection["Passive Crossbar"]
    CrossbarContent --> OneToneRSection["1T1R Section"]
    CrossbarContent --> ComparisonTable["Cell Comparison Table"]
    CrossbarContent --> SneakPathExplain["Sneak Path Explanation"]
    CrossbarContent --> RecommendCard["Recommendation Card"]

    FilesContent --> FileTitle["EDA Files We Generate"]
    FilesContent --> FileCards["4-card Grid<br/>LEF, DEF, Verilog, Liberty"]
    FilesContent --> GenSection["How We Generate Files"]
    FilesContent --> ValSection["How We Validate"]
    FilesContent --> ImgSection["Layout Visualization"]
    FilesContent --> PurposesSection["File Format Summary"]
    FilesContent --> ReferencesSection["References"]

    style LearnTab fill:#b3e5fc
    style Sidebar fill:#f0f4c3
    style ContentScroll fill:#e1bee7
    style T0 fill:#fff9c4
    style T1 fill:#fff9c4
    style T2 fill:#fff9c4
```

## 12. Learn Tab - Topic 1: Intro Content Structure

```mermaid
graph TD
    IntroContent["Intro Content VBox"] --> T0["Title: What is FeCIM EDA?"]
    IntroContent --> Sep1["Separator"]
    IntroContent --> Intro["Intro Paragraph<br/>Array builder explanation"]
    IntroContent --> Sep2["Separator"]

    IntroContent --> OperModes["OperationModesVisual()"]
    OperModes --> OMShowBox["Shows 3 modes side-by-side<br/>storage, memory, compute"]

    IntroContent --> Sep3["Separator"]

    IntroContent --> Flow["OpenLaneFlowDiagram()"]
    Flow --> FlowShow["RTL → Synthesis → Floorplan<br/>→ Placement → CTS → Routing → Signoff"]

    IntroContent --> Sep4["Separator"]

    IntroContent --> Stages["The Stages Explained"]
    Stages --> S1["1. SYNTHESIS: Yosys conversion"]
    Stages --> S2["2. FLOORPLAN: Die area definition"]
    Stages --> S3["3. PLACEMENT: X,Y assignment"]
    Stages --> S4["4. CTS: Clock tree"]
    Stages --> S5["5. ROUTING: Metal wires"]
    Stages --> S6["6. SIGNOFF: GDSII assembly"]

    IntroContent --> Sep5["Separator"]

    IntroContent --> DoColumns["2-column Grid"]
    DoColumns --> DoList["WHAT WE DO:<br/>LEF, LIB, V, DEF, Config"]
    DoColumns --> DontList["WHAT WE DON'T DO:<br/>No FeFET models, no GDSII,<br/>no timing characterization"]

    IntroContent --> Sep6["Separator"]

    IntroContent --> Disclaimer["Disclaimer Card<br/>No affiliation with Rice/Tour"]

    style IntroContent fill:#e1bee7
    style OperModes fill:#fff9c4
    style Flow fill:#fff9c4
    style Stages fill:#fff9c4
    style DoList fill:#fff9c4
    style DontList fill:#fff9c4
```

## 13. Learn Tab - Topic 2: Crossbar Content Structure

```mermaid
graph TD
    CrossbarContent["Crossbar Content VBox"] --> Title["The Crossbar Architecture"]
    CrossbarContent --> Sep1["Separator"]

    CrossbarContent --> PassiveSec["Passive Crossbar Section"]
    PassiveSec --> PassiveTitle["Passive Crossbar"]
    PassiveSec --> PassiveDesc["Description + Pros/Cons"]
    PassiveSec --> PassiveDiag["IsometricCrossbar(3,3,true)"]

    CrossbarContent --> Sep2["Separator"]

    CrossbarContent --> OneToneRSec["1T1R Section"]
    OneToneRSec --> OneToneRTitle["1T1R Crossbar"]
    OneToneRSec --> OneToneRDesc["Description + Pros/Cons"]
    OneToneRSec --> OneToneRDiag["Isometric1T1RCrossbar(3,3)"]

    CrossbarContent --> Sep3["Separator"]

    CrossbarContent --> CompTable["CellComparisonTable()"]
    CompTable --> TableShow["Rows: Passive, 1T1R, 2T1R<br/>Columns: Size, Cost, Scaling"]

    CrossbarContent --> Sep4["Separator"]

    CrossbarContent --> SneakPath["Sneak Path Problem"]
    SneakPath --> SneakDesc["Detailed explanation<br/>with current flow example"]

    CrossbarContent --> Sep5["Separator"]

    CrossbarContent --> Recommend["Recommendation Card"]
    Recommend --> RecText["<= 16x16: Passive<br/>32x32: Either<br/>>= 64x64: 1T1R required"]

    style CrossbarContent fill:#e1bee7
    style PassiveSec fill:#fff9c4
    style OneToneRSec fill:#fff9c4
    style CompTable fill:#fff9c4
    style SneakPath fill:#fff9c4
```

## 14. Learn Tab - Topic 3: Files Content Structure

```mermaid
graph TD
    FilesContent["Files Content VBox"] --> Title["EDA Files We Generate"]
    FilesContent --> Sep1["Separator"]

    FilesContent --> Cards["AdaptiveGrid 2"]
    Cards --> LEFCard["LEFPreviewCard()"]
    Cards --> DEFCard["DEFPreviewCard()"]
    Cards --> VCard["VerilogPreviewCard()"]
    Cards --> LibCard["LibertyPreviewCard()"]

    FilesContent --> Sep2["Separator"]

    FilesContent --> GenTitle["1. How We Generate Files"]
    GenTitle --> GenText["Verilog: Loop rows×cols, instantiate cells<br/>DEF: Calculate area, place cells<br/>LEF: Define geometry"]

    FilesContent --> Sep3["Separator"]

    FilesContent --> ValTitle["2. How We Validate"]
    ValTitle --> ValText["Yosys: read_verilog check<br/>DEF: syntax + structure<br/>Cross: pin name matching<br/>OpenLane: placement check"]

    FilesContent --> Sep4["Separator"]

    FilesContent --> ImgTitle["3. Layout Visualization"]
    ImgTitle --> ImgText["KLayout via Docker<br/>Magic via OpenROAD<br/>Manual tools supported"]

    FilesContent --> Sep5["Separator"]

    FilesContent --> PurpTitle["4. File Format Summary"]
    PurpTitle --> PurpText["LEF: geometry abstract<br/>DEF: placement<br/>V: netlist<br/>LIB: timing (placeholders!)"]

    FilesContent --> Sep6["Separator"]

    FilesContent --> RefTitle["References"]
    FilesContent --> RefCard["ReferencesCard()"]

    style FilesContent fill:#e1bee7
    style Cards fill:#fff9c4
    style GenText fill:#fffde7
    style ValText fill:#fffde7
    style ImgText fill:#fffde7
```

## 15. Widget Inventory - Complete Count

| Component Type | Widget Type | Count | Purpose | Notes |
|---|---|---|---|---|
| **Text Input** | Entry | 9 | Cell config, array config | Parse and validate |
| **Selection** | Select | 1 | Mode (storage/memory/compute) | UpdateModeHelp callback |
| **Toggle Buttons** | Button | 3 | Architecture (PASSIVE/1T1R/2T1R) | State tracking via Importance |
| **Action Buttons** | Button | 6 | Generate, Validate, Export, Gen Schematic, Gen Layout, Clear Log | Disable/Enable on state |
| **Status/Display** | Label | 20+ | Status, stats, results, help text | Dynamic updates via fyne.Do |
| **Containers** | HSplit, VSplit | 2 | Config layout, preview/validation split | Resizable 45/55, 75/25 |
| **Containers** | HBox, VBox | 15+ | Rows and columns throughout | Layout management |
| **Containers** | GridWithColumns | 4+ | Cell config (6), Array config (6), Architecture (3) | Compact grid layout |
| **Text Display** | MultiLineEntry | 3 | Verilog preview, DEF preview, Log output | Monospace, read-only |
| **Images** | canvas.Image | 3 | KLayout, OpenROAD, Yosys schematics | 400x350 each |
| **Cards** | widget.Card | 3 | Image containers, disclaimer, recommendation | Header + content |
| **Tabs** | AppTabs | 2 | Builder, Learn | TabLocation=Top |
| **Scroll** | Scroll | 5+ | Preview content, log, Learn content | Dynamic sizing |

## 16. Data Flow - Config Object Lifecycle

```mermaid
graph TD
    Init["Initialization"] --> ArrayCfg["ArrayConfig{<br/>Rows:4, Cols:4,<br/>Architecture:passive}"]

    ArrayCfg --> EntryFields["Entry Field OnChanged"]
    EntryFields --> UpdateStats["updateStats()"]
    UpdateStats --> ParseConfig["Parse from entries"]
    ParseConfig --> ModifyConfig["Modify cfg struct"]
    ModifyConfig --> CalcMetrics["Calculate new metrics"]
    CalcMetrics --> RefreshUI["fyne.Do() Update labels"]

    ArrayCfg --> ArchButton["Architecture Button OnTapped"]
    ArchButton --> SetArch["Set cfg.Architecture"]
    SetArch --> SetDims["Set Width/Height"]
    SetDims --> TriggerOnChanged["Trigger widthEntry.OnChanged"]
    TriggerOnChanged --> UpdateStats

    ArrayCfg --> GenerateBtn["Generate All Button"]
    GenerateBtn --> Goroutine["Launch goroutine"]
    Goroutine --> UseCfg["Use cfg for generation"]
    UseCfg --> Export["Export module"]
    Export --> GenerateAll["LEF, LIB, V, DEF, PNG, Config"]

    ArrayCfg --> ValidateBtn["Validate All Button"]
    ValidateBtn --> ValidateGo["Launch goroutine"]
    ValidateGo --> ValidateCfg["Use cfg for paths"]
    ValidateCfg --> ValidationModule["Validation module"]

    style ArrayCfg fill:#f3e5f5
    style UpdateStats fill:#c8e6c9
    style SetArch fill:#ffe0b2
    style GenerateAll fill:#b3e5fc
```

## 17. Error Handling & User Feedback

```mermaid
graph TD
    Action["User Action"] --> Goroutine["Background Goroutine"]

    Goroutine --> Try["Attempt Operation"]
    Try --> Success{Operation<br/>Successful?}

    Success -->|Yes| UpdateSuccess["fyne.Do()"]
    UpdateSuccess --> SetStatus["Set status to OK"]
    SetStatus --> UpdateLabel["Update relevant label"]
    UpdateLabel --> AddLogSuccess["Add success to log"]

    Success -->|No| CaptureError["Capture error"]
    CaptureError --> UpdateError["fyne.Do()"]
    UpdateError --> SetErrorStatus["Set status to ERROR"]
    SetErrorStatus --> UpdateErrorLabel["Update result label"]
    UpdateErrorLabel --> AddLogError["Add error to log"]
    AddLogError --> Optional["Show dialog if critical"]

    UpdateSuccess --> ReEnable["Re-enable buttons"]
    UpdateError --> ReEnable

    ReEnable --> Ready["Ready for next action"]

    style UpdateSuccess fill:#c8e6c9
    style UpdateError fill:#ffccbc
    style Ready fill:#b3e5fc
```

## 18. File Structure on Disk (After Generate All)

```mermaid
graph TD
    ProjectRoot["Project Root"] --> CellsDir["cells/"]
    ProjectRoot --> DataDir["data/"]

    CellsDir --> PassiveCell["fecim_bitcell/ (or 1t1r/2t1r)"]
    PassiveCell --> CellLEF["fecim_bitcell.lef"]
    PassiveCell --> CellLIB["fecim_bitcell.lib"]
    PassiveCell --> CellV["fecim_bitcell.v"]

    DataDir --> ArrayV["fecim_crossbar_NxN.v"]
    DataDir --> ArrayDEF["fecim_crossbar_NxN.def"]
    DataDir --> ArrayPNG["fecim_crossbar_NxN.png"]
    DataDir --> ArrayDOT["fecim_crossbar_NxN_schematic.dot"]
    DataDir --> ArraySchemPNG["fecim_crossbar_NxN_schematic.png"]
    DataDir --> ArrayOpenROADPNG["fecim_crossbar_NxN_openroad.png"]
    DataDir --> ConfigJSON["config.json"]
    DataDir --> ExportDir["fecim_crossbar_NxN/"]

    ExportDir --> ExportCells["cells/ (copy)"]
    ExportDir --> ExportV["fecim_crossbar_NxN.v"]
    ExportDir --> ExportDEF["fecim_crossbar_NxN.def"]
    ExportDir --> ExportJSON["fecim_crossbar_NxN.json"]
    ExportDir --> ExportConfig["config.json"]
    ExportDir --> ExportREADME["README.md"]

    style CellLEF fill:#fff9c4
    style ArrayV fill:#fff9c4
    style ArrayDEF fill:#fff9c4
    style ConfigJSON fill:#ffe0b2
```

## Related Documentation

- **Main Architecture**: `<local-path>`
- **Source Code**:
  - `<local-path>` - Standalone app entry
  - `<local-path>` - Unified app integration
  - `<local-path>` - Builder tab (1286 lines)
  - `<local-path>` - Learn tab
  - `<local-path>` - Visual components
- **Export Module**: `<local-path>`
- **Validation Module**: `<local-path>`

## Key Implementation Details

### Threading Safety
All UI updates use `fyne.Do(func() { ... })` to marshal operations back to main thread:
- Entry label updates
- Image display updates
- Button enable/disable
- Status and log output

### State Management Pattern
1. Entry fields trigger `OnChanged` callbacks
2. `updateStats()` parses and validates all entries
3. Metrics calculated in background (no blocking)
4. UI updated via `fyne.Do()` within goroutine
5. Buttons disabled during long operations

### Architecture Selection Flow
- Three toggle buttons (PASSIVE, 1T1R, 2T1R)
- Button `Importance` property shows selection (High=selected)
- Selecting architecture auto-updates cell dimensions
- Dimensions update triggers statistics recalculation
- All changes propagate through entry OnChanged callbacks

### Error Resilience
- File I/O errors caught and logged
- Invalid numeric entries fall back to current/default values
- Missing files handled gracefully (status shows reason)
- Docker/OpenLane detection non-blocking
- Validation continues even if individual checks fail
