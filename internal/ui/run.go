package ui

import (
	"fmt"
	"io"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-tfe"
	"github.com/renato0307/terrui/internal/client"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v2"
)

const RunPageName string = "Run"

type RunPage struct {
	*tview.Flex

	app *App

	run         *tfe.Run
	plan        *tfe.Plan
	apply       *tfe.Apply
	jsonPlan    []byte
	readerApply io.Reader

	loadingChan chan string
}

type runBaseInfo struct {
	ID        string `yaml:"ID"`
	Message   string `yaml:"Message"`
	CreatedAt string `yaml:"Created At"`
	Source    string `yaml:"Source"`
	Status    string `yaml:"Status"`
	AutoApply bool   `yaml:"AutoApply"`
	IsDestroy bool   `yaml:"Is Destroy"`
}

func NewRunPage(app *App) Page {
	r := RunPage{
		Flex: tview.NewFlex(),
		app:  app,
	}

	return &r
}

func (r *RunPage) Load() error {
	r.loadingChan = make(chan string)

	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	run, err := tfeClient.ReadWorkspaceRun(r.app.config.RunID)
	if err != nil {
		return fmt.Errorf("error reading the run: %w", err)
	}
	r.run = run

	plan, err := tfeClient.ReadWorkspacePlan(run.Plan.ID)
	if err != nil {
		return fmt.Errorf("error reading the plan: %w", err)
	}
	r.plan = plan

	var errPlan error
	go func() {
		planJSON, err := tfeClient.ReadWorkspacePlanJSON(run.Plan.ID)
		if err != nil {
			errPlan = fmt.Errorf("error reading the plan details: %w", err)
		} else {
			r.jsonPlan = planJSON
			r.loadingChan <- "plan"
		}
	}()

	var errLoadApply error
	go func() {
		readerApplyLogs, err := tfeClient.ReadWorkspaceApplyLogs(run.Plan.ID)
		if err != nil {
			errLoadApply = fmt.Errorf("error reading the apply details: %w", err)
		} else {
			r.readerApply = readerApplyLogs
			r.loadingChan <- "apply"
		}
	}()

	if errPlan != nil || errLoadApply != nil {
		return fmt.Errorf("could not load run details")
	}

	return nil
}

func (r *RunPage) View() string {

	runBase := runBaseInfo{
		ID:        r.run.ID,
		Message:   r.run.Message,
		Source:    string(r.run.Source),
		Status:    string(r.run.Status),
		AutoApply: r.run.AutoApply,
		IsDestroy: r.run.IsDestroy,
		CreatedAt: fmtTime(r.run.CreatedAt),
	}
	yamlBaseData, _ := yaml.Marshal(runBase)

	details := tview.NewTextView()
	plan := tview.NewTextView()
	apply := tview.NewTextView()

	details.SetBorder(true)
	details.SetBorderPadding(0, 1, 1, 1)
	details.SetTitle(" details ")
	details.SetText(colorizeYAML(string(yamlBaseData)))
	details.SetDynamicColors(true)

	plan.SetBorder(true)
	plan.SetBorderPadding(0, 1, 1, 1)
	plan.SetTitle(" plan ")
	plan.SetText("⏳ loading...")
	plan.SetDynamicColors(true)

	apply.SetBorder(true)
	apply.SetBorderPadding(0, 1, 1, 1)
	apply.SetTitle(" apply ")
	apply.SetText("⏳ loading...")
	apply.SetDynamicColors(true)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(details, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(plan, 0, 1, false).
			AddItem(apply, 0, 1, false), 0, 5, false)

	r.Flex = flex

	go func() {
		loadingAsyncsFinished := 0
		for v := range r.loadingChan {
			loadingAsyncsFinished++
			fmt.Println(v, loadingAsyncsFinished)
			if v == "plan" {
				if r.jsonPlan == nil {
					plan.SetText("plan details could not be loaded!")
					return
				}
				plan.SetText("loading done!")
			} else if v == "apply" {
				if r.jsonPlan == nil {
					apply.SetText("apply details could not be loaded!")
					return
				}
				apply.SetText("loading done!")
			}

			if loadingAsyncsFinished == 2 {
				break
			}
		}
	}()

	return "run loaded"
}

func (r *RunPage) BindKeys() KeyActions {
	return KeyActions{
		tcell.KeyCtrlW: NewKeyAction("go back to the workspace", r.actionReturnToWorkspace, true),
	}
}

func (r *RunPage) actionReturnToWorkspace(ek *tcell.EventKey) *tcell.EventKey {
	r.app.activatePage(WorkspacePageName, nil, false)

	return nil
}

func (r *RunPage) Crumb() []string {
	return []string{
		r.app.config.Organization,
		r.app.config.Workspace,
		"runs",
		r.app.config.RunID,
	}
}

func (r *RunPage) Name() string {
	return RunPageName
}

func (r *RunPage) Footer() string {
	return ""
}
