package render_test

import (
	"strings"
	"testing"

	"github.com/makima333/uipath-xaml-tools/internal/normalize"
	"github.com/makima333/uipath-xaml-tools/internal/render"
	"github.com/makima333/uipath-xaml-tools/internal/xaml"
)

const sampleXAML = `<?xml version="1.0" encoding="utf-8"?>
<Sequence xmlns="http://schemas.microsoft.com/netfx/2009/xaml/activities"
 xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml"
 xmlns:sap="http://schemas.microsoft.com/netfx/2009/xaml/activities/presentation"
 xmlns:sap2010="http://schemas.microsoft.com/netfx/2010/xaml/activities/presentation"
 DisplayName="Main" sap:VirtualizedContainerService.HintSize="500,400" sap2010:WorkflowViewState.IdRef="Sequence_1">
  <Assign DisplayName="Set name">
    <Assign.To><OutArgument x:TypeArguments="x:String">[name]</OutArgument></Assign.To>
    <Assign.Value><InArgument x:TypeArguments="x:String">[&quot;Makima&quot;]</InArgument></Assign.Value>
  </Assign>
	  <If Condition="[String.IsNullOrEmpty(name)]">
	    <If.Then><Sequence><WriteLine Text="[&quot;empty&quot;]" /></Sequence></If.Then>
	  </If>
	  <WriteLine DisplayName="{x:Null}" Text="[name]" />
	</Sequence>`

func TestCompactSpecializesCommonActivities(t *testing.T) {
	root, err := xaml.Parse(strings.NewReader(sampleXAML))
	if err != nil {
		t.Fatal(err)
	}
	got := render.Compact(normalize.Document(root), render.CompactOptions{})
	for _, want := range []string{
		`Sequence "Main"`,
		`Assign [name] = ["Makima"] "Set name"`,
		`If [String.IsNullOrEmpty(name)]`,
		`Then:`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("compact output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "HintSize") {
		t.Fatalf("compact output retained designer noise:\n%s", got)
	}
	if strings.Contains(got, "{x:Null}") {
		t.Fatalf("compact output retained a null display name:\n%s", got)
	}
}

func TestCompactCanIncludeStableReference(t *testing.T) {
	root, err := xaml.Parse(strings.NewReader(sampleXAML))
	if err != nil {
		t.Fatal(err)
	}
	got := render.Compact(normalize.Document(root), render.CompactOptions{Refs: true})
	if !strings.Contains(got, "#Sequence_1") || !strings.Contains(got, "#/Sequence[1]/Assign[1]") {
		t.Fatalf("references missing:\n%s", got)
	}
}
