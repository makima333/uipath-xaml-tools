package strip_test

import (
	"strings"
	"testing"

	stripper "github.com/makima333/uipath-xaml-tools/internal/strip"
	"github.com/makima333/uipath-xaml-tools/internal/xaml"
)

const noisy = `<Sequence xmlns="urn:activities" xmlns:sap="urn:sap" xmlns:sap2010="urn:sap2010" sap:VirtualizedContainerService.HintSize="400,300" sap2010:WorkflowViewState.IdRef="Sequence_1">
  <sap:WorkflowViewStateService.ViewState><State>noise</State></sap:WorkflowViewStateService.ViewState>
  <TextExpression.NamespacesForImplementation><String>System</String></TextExpression.NamespacesForImplementation>
  <WriteLine Text="[&quot;hello&quot;]" InformativeScreenshot="abc" />
</Sequence>`

func TestAggressiveStripRemovesKnownNoiseAndStaysXML(t *testing.T) {
	got, err := stripper.XAML(noisy, stripper.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{"HintSize", "IdRef", "ViewState", "NamespacesForImplementation", "InformativeScreenshot"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("strip retained %s:\n%s", unwanted, got)
		}
	}
	if _, err := xaml.Parse(strings.NewReader(got)); err != nil {
		t.Fatalf("strip result is not XML: %v", err)
	}
}

func TestConservativeStripKeepsCompileMetadata(t *testing.T) {
	got, err := stripper.XAML(noisy, stripper.Options{Conservative: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "NamespacesForImplementation") || !strings.Contains(got, "InformativeScreenshot") {
		t.Fatalf("conservative strip removed non-designer metadata:\n%s", got)
	}
}
