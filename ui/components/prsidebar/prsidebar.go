package prsidebar

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/components/pr"
	"github.com/dlvhdr/gh-prs/ui/constants"
	"github.com/dlvhdr/gh-prs/ui/context"
	"github.com/dlvhdr/gh-prs/ui/markdown"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

type Model struct {
	pr              *pr.PullRequest
	IsOpen          bool
	sidebarViewport viewport.Model
	ctx             *context.ProgramContext
}

func NewModel() Model {
	return Model{
		pr:     nil,
		IsOpen: false,
		sidebarViewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
		ctx: nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keys.PageDown):
			m.sidebarViewport.HalfViewDown()

		case key.Matches(msg, constants.Keys.PageUp):
			m.sidebarViewport.HalfViewUp()
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.IsOpen {
		return ""
	}

	height := m.ctx.MainContentHeight
	style := sideBarStyle.Copy().
		Height(height).
		MaxHeight(height).
		Width(m.ctx.Config.Defaults.Preview.Width).
		MaxWidth(m.ctx.Config.Defaults.Preview.Width)

	if m.pr == nil {
		return style.Copy().Align(lipgloss.Center).Render(
			lipgloss.PlaceVertical(height, lipgloss.Center, "Select a Pull Request..."),
		)
	}

	return style.Copy().Render(lipgloss.JoinVertical(
		lipgloss.Top,
		m.sidebarViewport.View(),
		pagerStyle.Copy().Render(fmt.Sprintf("%d%%", int(m.sidebarViewport.ScrollPercent()*100))),
	))
}

func (m *Model) SetPrData(prData *data.PullRequestData) {
	if prData == nil {
		m.pr = nil
	} else {
		m.pr = &pr.PullRequest{Data: *prData}
	}
	m.setSidebarViewportContent()
}

func (m *Model) setSidebarViewportContent() {
	if m.pr == nil {
		return
	}

	s := strings.Builder{}
	s.WriteString(m.renderTitle())
	s.WriteString("\n")
	s.WriteString(m.renderBranches())
	s.WriteString("\n\n")
	s.WriteString(m.renderPills())
	s.WriteString("\n\n")
	s.WriteString(m.renderDescription())
	s.WriteString("\n\n")
	s.WriteString(m.renderChecks())
	s.WriteString("\n\n")
	s.WriteString(m.renderActivity())

	m.sidebarViewport.SetContent(s.String())
}

func (m *Model) renderTitle() string {
	return styles.MainTextStyle.Copy().Width(m.GetSidebarContentWidth() - 6).
		Render(m.pr.Data.Title)
}

func (m *Model) renderBranches() string {
	return lipgloss.NewStyle().
		Foreground(styles.DefaultTheme.SecondaryText).
		Render(m.pr.Data.BaseRefName + "  " + m.pr.Data.HeadRefName)
}

func (m *Model) renderStatusPill() string {
	bgColor := ""
	switch m.pr.Data.State {
	case "OPEN":
		bgColor = openPR.Dark
	case "CLOSED":
		bgColor = closedPR.Dark
	case "MERGED":
		bgColor = mergedPR.Dark
	}

	return pillStyle.
		Background(lipgloss.Color(bgColor)).
		Render(m.pr.RenderState())
}

func (m *Model) renderMergeablePill() string {
	status := m.pr.Data.Mergeable
	if status == "CONFLICTING" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.WarningText).
			Render(" Merge Conflicts")
	} else if status == "MERGEABLE" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.SuccessText).
			Render(" Mergeable")
	}

	return ""
}

func (m *Model) renderChecksPill() string {
	status := m.pr.GetStatusChecksRollup()
	if status == "FAILURE" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.WarningText).
			Render(" Checks")
	} else if status == "PENDING" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.FaintText).
			Render(constants.WaitingGlyph + " Checks")
	}

	return pillStyle.Copy().
		Background(styles.DefaultTheme.SuccessText).
		Foreground(styles.DefaultTheme.SubleMainText).
		Render(" Checks")
}

func (m *Model) renderPills() string {
	statusPill := m.renderStatusPill()
	mergeablePill := m.renderMergeablePill()
	checksPill := m.renderChecksPill()
	return lipgloss.JoinHorizontal(lipgloss.Top, statusPill, " ", mergeablePill, " ", checksPill)
}

func (m *Model) renderDescription() string {
	width := m.GetSidebarContentWidth() - 6
	regex := regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	body := regex.ReplaceAllString(m.pr.Data.Body, "")

	body = strings.TrimSpace(body)
	if body == "" {
		return lipgloss.NewStyle().Italic(true).Render("No description provided.")
	}

	markdownRenderer := markdown.GetMarkdownRenderer(width)
	rendered, err := markdownRenderer.Render(body)
	if err != nil {
		return ""
	}

	return lipgloss.NewStyle().
		MaxHeight(10).
		Width(width).
		MaxWidth(width).
		Align(lipgloss.Left).
		Render(rendered)
}

func (m *Model) GetSidebarContentWidth() int {
	if m.ctx.Config == nil {
		return 0
	}
	return m.ctx.Config.Defaults.Preview.Width - 2*contentPadding - borderWidth
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
	m.sidebarViewport.Height = m.ctx.MainContentHeight - pagerHeight
	m.sidebarViewport.Width = m.GetSidebarContentWidth()
}
