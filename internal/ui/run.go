package ui

import (
	"bufio"
	"bytes"
	"encoding/json"
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

	planPrimitive  *tview.TextView
	applyPrimitive *tview.TextView

	app *App

	run   *tfe.Run
	plan  *tfe.Plan
	apply *tfe.Apply

	planReader  io.Reader
	applyReader io.Reader

	loadingChan chan string

	sections []*tview.Box
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

type applyLogEntry struct {
	Level            string         `json:"@level"`
	Message          string         `json:"@message"`
	Module           string         `json:"@module"`
	Change           applyLogChange `json:"change"`
	Timestamp        string         `json:"@timestamp"`
	TerraformVersion string         `json:"@terraform"`
	Type             string         `json:"type"`
	UI               string         `json:"@ui"`
}

type applyLogChange struct {
	Resource applyLogChangeResource `json:"resource"`
	Action   string                 `json:"action"`
}

type applyLogChangeResource struct {
	Addr            string `json:"addr"`
	Module          string `json:"module"`
	Resource        string `json:"resource"`
	ImpliedProvider string `json:"implied_provider"`
	ResourceType    string `json:"resource_type"`
	ResourceName    string `json:"resource_name"`
	ResourceKey     string `json:"resource_key"`
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

	errLoadPlan, errLoadApply := r.loadPlanAndApplyDetails(tfeClient, run)

	if errLoadPlan != nil || errLoadApply != nil {
		return fmt.Errorf("could not load run details")
	}

	return nil
}

func (r *RunPage) loadPlanAndApplyDetails(tfeClient client.TFEClient, run *tfe.Run) (error, error) {
	var errLoadPlan error
	go func() {
		planLogsReader, err := tfeClient.ReadWorkspacePlanLogs(run.Plan.ID)
		if err != nil {
			errLoadPlan = fmt.Errorf("error reading the plan details: %w", err)
		} else {
			r.planReader = planLogsReader
		}
		r.loadingChan <- "plan"
	}()

	var errLoadApply error
	go func() {
		applyLogsReader, err := tfeClient.ReadWorkspaceApplyLogs(run.Apply.ID)
		if err != nil {
			errLoadApply = fmt.Errorf("error reading the apply details: %w", err)
		} else {
			r.applyReader = applyLogsReader
		}
		r.loadingChan <- "apply"
	}()
	return errLoadPlan, errLoadApply
}

func (r *RunPage) View() string {
	r.sections = []*tview.Box{}

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
	details.SetTitle(" run details ")
	details.SetText(colorizeYAML(string(yamlBaseData)))
	details.SetDynamicColors(true)

	plan.SetBorder(true)
	plan.SetBorderPadding(0, 1, 1, 1)
	plan.SetTitle(" plan ")
	plan.SetText("⏳ loading...")
	plan.SetDynamicColors(true)
	plan.SetWrap(false)
	r.sections = append(r.sections, plan.Box)

	apply.SetBorder(true)
	apply.SetBorderPadding(0, 1, 1, 1)
	apply.SetTitle(" apply ")
	apply.SetText("⏳ loading...")
	apply.SetDynamicColors(true)
	apply.SetWrap(false)
	r.sections = append(r.sections, apply.Box)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(details, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(plan, 0, 1, false).
			AddItem(apply, 0, 1, false), 0, 5, false)

	r.Flex = flex
	r.planPrimitive = plan
	r.applyPrimitive = apply

	go r.viewPlanAndApplyDetails()

	return "run loaded"
}

func (r *RunPage) viewPlanAndApplyDetails() {
	loadingAsyncsFinished := 0
	for v := range r.loadingChan {
		if v == "plan" {
			loadingAsyncsFinished++
			if r.planReader == nil {
				r.planPrimitive.SetText("plan details could not be loaded!")
				continue
			}
			go r.viewPlanDetails()
		} else if v == "apply" {
			loadingAsyncsFinished++
			if r.applyReader == nil {
				r.applyPrimitive.SetText("apply details could not be loaded!")
				continue
			}
			go r.viewApplyDetails()
		}

		if loadingAsyncsFinished == 2 {
			break
		}
	}
}

func (r *RunPage) viewPlanDetails() {
	scanner := bufio.NewScanner(r.planReader)
	scanner.Split(bufio.ScanLines)

	outputBuf := bytes.Buffer{}
	for scanner.Scan() {
		log := applyLogEntry{}
		err := json.Unmarshal(scanner.Bytes(), &log)

		if err != nil {
			outputBuf.WriteString(fmt.Sprintln(scanner.Text()))
			continue
		}

		outputBuf.WriteString(fmt.Sprintln(log.Message))
	}
	r.planPrimitive.SetText(outputBuf.String())
}

func (r *RunPage) viewApplyDetails() {
	scanner := bufio.NewScanner(r.applyReader)
	scanner.Split(bufio.ScanLines)

	outputBuf := bytes.Buffer{}
	for scanner.Scan() {
		log := applyLogEntry{}
		err := json.Unmarshal(scanner.Bytes(), &log)

		if err != nil {
			outputBuf.WriteString(fmt.Sprintln(scanner.Text()))
			continue
		}

		outputBuf.WriteString(fmt.Sprintln(log.Message))
	}
	r.applyPrimitive.SetText(outputBuf.String())
}

func (r *RunPage) BindKeys() KeyActions {
	return KeyActions{
		tcell.KeyCtrlW: NewKeyAction("go back to the workspace", r.actionReturnToWorkspace, true),
		tcell.KeyTab:   NewKeyAction("focus plan and apply cells", r.actionFocusNextList, true),
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

func (r *RunPage) actionFocusNextList(ek *tcell.EventKey) *tcell.EventKey {
	for i, b := range r.sections {
		if !b.HasFocus() {
			continue
		}

		nextToFocus := i + 1
		if nextToFocus == len(r.sections) {
			r.app.SetFocus(r)
		} else {
			r.app.SetFocus(r.sections[nextToFocus])
		}

		return nil
	}

	// No section was focused
	r.app.SetFocus(r.sections[0])
	return nil
}
