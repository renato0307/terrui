package ui

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-tfe"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v2"

	"github.com/renato0307/terrui/internal/client"
)

const WorkspacePageName string = "workspace"

type WorkspacePage struct {
	*tview.Flex

	app       *App
	workspace *tfe.Workspace
	variables *tfe.VariableList
	runs      *tfe.RunList
}

type workspaceBaseInfo struct {
	ID               string `yaml:"ID"`
	Name             string `yaml:"Name"`
	Description      string `yaml:"Description"`
	Resources        int    `yaml:"Resource Count"`
	TerraformVersion string `yaml:"Terraform Version"`
	Updated          string `yaml:"Updated"`
	Locked           bool   `yaml:"Locked"`
	WorkingDirectory string `yaml:"Working Directory"`
	ExecutionMode    string `yaml:"Execution Mode"`
	AutoApply        bool   `yaml:"Auto Apply"`
}

type workspaceLastRun struct {
	By               string `yaml:"By"`
	When             string `yaml:"When"`
	Status           string `yaml:"Status"`
	ResourcesAdded   int    `yaml:"Resources Added"`
	ResourcesUpdated int    `yaml:"Resources Updated"`
	ResourcesDeleted int    `yaml:"Resources Deleted"`
}

type workspaceMetrics struct {
	ApplyDurationAverage string `yaml:"Average Apply Duration"`
	PlanDurationAverage  string `yaml:"Average Plan Duration"`
	RunFailures          int    `yaml:"Total Failed Runs"`
}

func NewWorkspacePage(app *App) Page {
	ol := WorkspacePage{
		Flex: tview.NewFlex(),
		app:  app,
	}

	return &ol
}

func (w *WorkspacePage) Load() error {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	workspace, err := tfeClient.ReadWorkspace(w.app.config.Organization, w.app.config.Workspace)
	if err != nil {
		return fmt.Errorf("error reading the workspace: %w", err)
	}
	w.workspace = workspace

	vars, err := tfeClient.ListWorkspaceVariables(workspace.ID)
	if err != nil {
		return fmt.Errorf("error reading the workspace variables: %w", err)
	}
	w.variables = vars

	runs, err := tfeClient.ListWorkspaceRuns(workspace.ID)
	if err != nil {
		return fmt.Errorf("error reading the workspace variables: %w", err)
	}
	w.runs = runs

	return nil
}

func (w *WorkspacePage) View() string {
	workspace := w.workspace

	workspaceBase := workspaceBaseInfo{
		ID:               workspace.ID,
		Name:             workspace.Name,
		Description:      workspace.Description,
		TerraformVersion: workspace.TerraformVersion,
		Resources:        workspace.ResourceCount,
		Updated:          fmtTime(workspace.UpdatedAt),
		Locked:           workspace.Locked,
		WorkingDirectory: workspace.WorkingDirectory,
		ExecutionMode:    workspace.ExecutionMode,
		AutoApply:        workspace.AutoApply,
	}

	wLastRun := workspaceLastRun{}
	if workspace.CurrentRun != nil {
		if workspace.CurrentRun.CreatedBy != nil {
			wLastRun.By = workspace.CurrentRun.CreatedBy.Username
			wLastRun.When = fmtTime(workspace.CurrentRun.CreatedAt)
		}
		wLastRun.Status = string(workspace.CurrentRun.Status)
		wLastRun.ResourcesAdded = workspace.CurrentRun.Plan.ResourceAdditions
		wLastRun.ResourcesUpdated = workspace.CurrentRun.Plan.ResourceChanges
		wLastRun.ResourcesDeleted = workspace.CurrentRun.Plan.ResourceDestructions
	}

	workspaceMetrics := workspaceMetrics{
		ApplyDurationAverage: workspace.ApplyDurationAverage.String(),
		PlanDurationAverage:  workspace.PlanDurationAverage.String(),
		RunFailures:          workspace.RunFailures,
	}

	yamlBaseData, _ := yaml.Marshal(workspaceBase)
	yamlLastRunData, _ := yaml.Marshal(wLastRun)
	yamlMetrics, _ := yaml.Marshal(workspaceMetrics)

	details := tview.NewTextView()
	metrics := tview.NewTextView()
	tags := tview.NewList()
	lastRun := tview.NewTextView()
	variables := tview.NewList()
	runs := tview.NewList()

	details.SetBorder(true)
	details.SetBorderPadding(0, 1, 1, 1)
	details.SetTitle(" details ")
	details.SetText(colorizeYAML(string(yamlBaseData)))
	details.SetDynamicColors(true)

	metrics.SetBorder(true)
	metrics.SetBorderPadding(0, 1, 1, 1)
	metrics.SetTitle(fmt.Sprintf(" metrics (last %d runs) ", workspace.RunsCount))
	metrics.SetText(colorizeYAML(string(yamlMetrics)))
	metrics.SetDynamicColors(true)

	tags.SetBorder(true)
	tags.SetBorderPadding(0, 1, 1, 1)
	tags.SetTitle(" tags ")
	tags.SetMainTextStyle(tcell.StyleDefault.Bold(true))
	tags.SetSecondaryTextStyle(tcell.StyleDefault.Dim(true))
	tags.SetWrapAround(true)
	tags.SetSelectedFocusOnly(true)
	tags.ShowSecondaryText(false)
	for _, t := range workspace.TagNames {
		tags.AddItem(t, "", 0, nil)
	}

	lastRun.SetBorder(true)
	lastRun.SetBorderPadding(0, 1, 1, 1)
	lastRun.SetTitle(" last run ")
	lastRun.SetText(colorizeYAML(string(yamlLastRunData)))
	lastRun.SetDynamicColors(true)

	variables.SetBorder(true)
	variables.SetBorderPadding(0, 1, 1, 1)
	variables.SetSelectedFocusOnly(true)
	variables.SetMainTextStyle(tcell.StyleDefault.Bold(true))
	variables.SetSecondaryTextStyle(tcell.StyleDefault.Dim(true))
	variables.SetHighlightFullLine(true)
	variables.SetWrapAround(true)
	variables.SetTitle(" variables (most recent) ")

	runs.SetBorder(true)
	runs.SetBorderPadding(0, 1, 1, 1)
	runs.SetSelectedFocusOnly(true)
	runs.SetMainTextStyle(tcell.StyleDefault.Bold(true))
	runs.SetSecondaryTextStyle(tcell.StyleDefault.Dim(true))
	runs.SetHighlightFullLine(true)
	runs.SetWrapAround(true)
	runs.SetTitle(" runs (most recent) ")

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(details, 0, 3, false).
			AddItem(lastRun, 0, 1, false).
			AddItem(metrics, 0, 1, false).
			AddItem(tags, 0, 1, false), 0, 1, false)

	if w.app.config.WorkspaceShowVariables {
		flex.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(variables, 0, 2, false).
			AddItem(runs, 0, 2, false), 0, 1, false)
	}

	flex.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		details.ScrollToBeginning()
		lastRun.ScrollToBeginning()
		return x, y, width, height
	})
	w.Flex = flex

	w.showVariables(variables, true)
	w.showRuns(runs, true)

	return "workspace loaded"
}

func (w *WorkspacePage) showVariables(list *tview.List, showShortcuts bool) {
	max := 10
	shortcut := int('0')
	for i, v := range w.variables.Items {
		if i == max {
			break
		}

		value := v.Value
		if v.Sensitive {
			value = "******"
		}
		mainText := fmt.Sprintf("%s = %s", v.Key, value)
		secondaryText := fmt.Sprintf("%s", v.Category)
		if v.Sensitive {
			secondaryText += ", sensitive"
		}

		if !showShortcuts {
			shortcut = 0
		}
		list.AddItem(mainText, secondaryText, rune(shortcut), nil)

		if shortcut == '9' {
			shortcut = 'a' - 1
		} else if shortcut == 'z' {
			shortcut = 'A' - 1
		} else if shortcut == 'Z' {
			shortcut = '.'
		}

		if shortcut != '.' {
			shortcut = shortcut + 1
		}
	}
}

func (w *WorkspacePage) showRuns(list *tview.List, showShortcuts bool) {
	max := 10
	shortcut := int('0')
	for i, v := range w.runs.Items {
		if i == max {
			break
		}

		iconStatus := ""
		switch v.Status {
		case tfe.RunErrored:
			iconStatus = "❌ "
		case tfe.RunApplied, tfe.RunPlannedAndFinished:
			iconStatus = "✅	 "
		case tfe.RunPlanned, tfe.RunPlanning:
			iconStatus = "⏳ "
		case tfe.RunDiscarded:
			iconStatus = "🌪 "
		}

		mainText := fmt.Sprintf("%s » %s%s", v.Message, iconStatus, v.Status)

		createdBy := ""
		if v.CreatedBy != nil {
			createdBy = fmt.Sprintf(" | %s", v.CreatedBy.Username)
		}

		destroyRun := ""
		if v.IsDestroy {
			destroyRun = " | destroy run 🔥"
		}

		secondaryText := fmt.Sprintf("%s%s%s | %s", v.ID, createdBy, destroyRun, fmtTime(v.CreatedAt))

		if !showShortcuts {
			shortcut = 0
		}
		list.AddItem(mainText, secondaryText, rune(shortcut), nil)

		if shortcut == '9' {
			shortcut = 'a' - 1
		} else if shortcut == 'z' {
			shortcut = 'A' - 1
		} else if shortcut == 'Z' {
			shortcut = '.'
		}

		if shortcut != '.' {
			shortcut = shortcut + 1
		}
	}
}

func (w *WorkspacePage) BindKeys() KeyActions {
	return KeyActions{
		tcell.KeyCtrlL: NewKeyAction("list workspaces", w.listWorkspaces, true),
	}
}

func (w *WorkspacePage) Crumb() []string {
	return []string{
		w.app.config.Organization,
		w.app.config.Workspace,
	}
}

func (w *WorkspacePage) Name() string {
	return WorkspacePageName
}

func (w *WorkspacePage) Footer() string {
	return ""
}

func (w *WorkspacePage) listWorkspaces(ek *tcell.EventKey) *tcell.EventKey {
	w.app.config.Workspace = ""
	w.app.config.Save()

	w.app.activatePage(WorkspacesPageName, nil, false)

	return nil
}

func fmtTime(t time.Time) string {
	return fmt.Sprintf("%s (%s)", humanize.Time(t.Local()), t.Local().Format(time.RFC3339))
}
