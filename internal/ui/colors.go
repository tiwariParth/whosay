package ui

import (
	"github.com/fatih/color"
)

// Color functions for consistent styling across the application
var (
	// Basic color scheme
	TitleColor    = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	SectionColor  = color.New(color.FgHiBlue, color.Bold).SprintFunc()
	LabelColor    = color.New(color.FgHiYellow).SprintFunc() 
	ValueColor    = color.New(color.FgHiWhite).SprintFunc()
	SuccessColor  = color.New(color.FgHiGreen).SprintFunc()
	WarningColor  = color.New(color.FgHiYellow).SprintFunc()
	DangerColor   = color.New(color.FgHiRed).SprintFunc()
	SubtitleColor = color.New(color.FgCyan).SprintFunc()
	AccentColor   = color.New(color.FgHiMagenta).SprintFunc()
	DimColor      = color.New(color.FgWhite).SprintFunc()
	
	// Special purpose colors
	HeaderBgColor = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()
	InfoColor     = color.New(color.FgHiCyan).SprintFunc()
)

// Constants for box drawing characters to create modern UI elements
const (
	BoxHorizontal      = "━"
	BoxVertical        = "┃"
	BoxTopLeft         = "┏"
	BoxTopRight        = "┓"
	BoxBottomLeft      = "┗"
	BoxBottomRight     = "┛"
	BoxT               = "┳"
	BoxInvertedT       = "┻"
	BoxCross           = "╋"
	BoxLeftT           = "┣"
	BoxRightT          = "┫"
	ThinBoxHorizontal  = "─"
	BulletPoint        = "•"
	RightArrow         = "→"
	
	// Visual indicators
	CheckMark          = "✓"
	XMark              = "✗"
	Warning            = "⚠"
	Info               = "ℹ"
	Cpu                = "⚙"
	Memory             = "□"
	Disk               = "○"
	Network            = "⤭"
)
