// Package ui — deprecated Phosphor* aliases for the Remix Icon API.
//
// Prefer Icons, IconRegistry, Icon*, InitIcons, DrawIcon, and SetIcon.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Deprecated: use IconWeight.
type PhosphorWeight = IconWeight

// Deprecated weight aliases — use IconThin / IconRegular / …
const (
	PhosphorThin    = IconThin
	PhosphorLight   = IconLight
	PhosphorRegular = IconRegular
	PhosphorBold    = IconBold
	PhosphorFill    = IconFill
	PhosphorDuotone = IconDuotone
)

// Deprecated icon name aliases — use IconHouse / IconBell / …
const (
	PhosphorHouse             = IconHouse
	PhosphorMagnifyingGlass   = IconMagnifyingGlass
	PhosphorBell              = IconBell
	PhosphorGear              = IconGear
	PhosphorUser              = IconUser
	PhosphorUsers             = IconUsers
	PhosphorEnvelope          = IconEnvelope
	PhosphorTray              = IconTray
	PhosphorList              = IconList
	PhosphorMarkdownLine      = IconMarkdownLine
	PhosphorMarkdownFill      = IconMarkdownFill
	PhosphorCodeView          = IconCodeView
	PhosphorCodeBlock         = IconCodeBlock
	PhosphorHome2             = IconHome2
	PhosphorEditBox           = IconEditBox
	PhosphorBookRead          = IconBookRead
	PhosphorSearch            = IconSearch
	PhosphorCodeBox           = IconCodeBox
	PhosphorArrowDropDown     = IconArrowDropDown
	PhosphorArrowDropUp       = IconArrowDropUp
	PhosphorInfoI             = IconInfoI
	PhosphorTable             = IconTable
	PhosphorPlus              = IconPlus
	PhosphorMinus             = IconMinus
	PhosphorSquare            = IconSquare
	PhosphorResize            = IconResize
	PhosphorCopy              = IconCopy
	PhosphorX                 = IconX
	PhosphorXCircle           = IconXCircle
	PhosphorCheck             = IconCheck
	PhosphorCheckbox          = IconCheckbox
	PhosphorCheckboxBlank     = IconCheckboxBlank
	PhosphorArrowGoBack       = IconArrowGoBack
	PhosphorArrowGoForward    = IconArrowGoForward
	PhosphorScissorsCut       = IconScissorsCut
	PhosphorClipboard         = IconClipboard
	PhosphorFindReplace       = IconFindReplace
	PhosphorZoomIn            = IconZoomIn
	PhosphorZoomOut           = IconZoomOut
	PhosphorFileClose         = IconFileClose
	PhosphorCaretLeft         = IconCaretLeft
	PhosphorCaretRight        = IconCaretRight
	PhosphorCaretDown         = IconCaretDown
	PhosphorCaretUp           = IconCaretUp
	PhosphorCaretCircleDown   = IconCaretCircleDown
	PhosphorCaretCircleUp     = IconCaretCircleUp
	PhosphorCaretCircleLeft   = IconCaretCircleLeft
	PhosphorCaretCircleRight  = IconCaretCircleRight
	PhosphorDotsThree         = IconDotsThree
	PhosphorDotsThreeVertical = IconDotsThreeVertical
	PhosphorStar              = IconStar
	PhosphorHeart             = IconHeart
	PhosphorCalendar          = IconCalendar
	PhosphorCalendarBlank     = IconCalendarBlank
	PhosphorFunnel            = IconFunnel
	PhosphorPencilSimple      = IconPencilSimple
	PhosphorTrash             = IconTrash
	PhosphorUpload            = IconUpload
	PhosphorDownload          = IconDownload
	PhosphorFolder            = IconFolder
	PhosphorFolderOpen        = IconFolderOpen
	PhosphorRestart           = IconRestart
	PhosphorDatabase          = IconDatabase
	PhosphorSave2             = IconSave2
	PhosphorSettings4         = IconSettings4
	PhosphorSettings3         = IconSettings3
	PhosphorListSettings      = IconListSettings
	PhosphorFile              = IconFile
	PhosphorWifiHigh          = IconWifiHigh
	PhosphorMoon              = IconMoon
	PhosphorSun               = IconSun
	PhosphorTextBold          = IconTextBold
	PhosphorTextItalic        = IconTextItalic
	PhosphorTextUnderline     = IconTextUnderline
	PhosphorTextWrap          = IconTextWrap
	PhosphorAlignLeft         = IconAlignLeft
	PhosphorAlignCenter       = IconAlignCenter
	PhosphorAlignRight        = IconAlignRight
	PhosphorImage             = IconImage
	PhosphorLink              = IconLink
)

// Deprecated: use IconRegistry.
type PhosphorRegistry = IconRegistry

// Deprecated: use Icons. Kept in sync by SetDefaultIconRegistry / InitIcons.
var Phosphor = Icons

// Deprecated: use NewIconRegistry.
func NewPhosphorRegistry(root string) *IconRegistry { return NewIconRegistry(root) }

// Deprecated: use InitIcons.
func InitPhosphor(atlasSize int32) {
	InitIcons(atlasSize)
	Phosphor = Icons
}

// Deprecated: use DrawIcon.
func DrawPhosphorIcon(dst rl.Rectangle, name string, weight IconWeight, tint rl.Color) bool {
	return DrawIcon(dst, name, weight, tint)
}
