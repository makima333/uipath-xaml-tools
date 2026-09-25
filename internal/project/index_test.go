package project_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/makima333/uipath-xaml-tools/internal/project"
)

func TestAnalyzeAndGraph(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "project.json"), `{"name":"Demo","main":"Main.xaml","dependencies":{"UiPath.System.Activities":"[24.10.0]"}}`)
	write(t, filepath.Join(dir, "Main.xaml"), `<Activity xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml"><x:Members><x:Property Name="customer" Type="InArgument(x:String)" /></x:Members><TextExpression.NamespacesForImplementation><x:String>System</x:String></TextExpression.NamespacesForImplementation><Sequence><InvokeWorkflowFile WorkflowFileName="Sub.xaml" /></Sequence></Activity>`)
	write(t, filepath.Join(dir, "Sub.xaml"), `<Sequence><WriteLine Text="hello" /></Sequence>`)

	index, err := project.Analyze(dir)
	if err != nil {
		t.Fatal(err)
	}
	if index.Name != "Demo" || index.Main != "Main.xaml" || len(index.Workflows) != 2 {
		t.Fatalf("unexpected index: %#v", index)
	}
	if len(index.Workflows[0].Invokes) != 1 || index.Workflows[0].Invokes[0] != "Sub.xaml" {
		t.Fatalf("invoke not indexed: %#v", index.Workflows[0])
	}
	if index.Workflows[0].ActivityCount != 2 {
		t.Fatalf("metadata counted as activities: %#v", index.Workflows[0])
	}
	graph := project.GraphText(index)
	if !strings.Contains(graph, "Main.xaml -> Sub.xaml") {
		t.Fatalf("unexpected graph:\n%s", graph)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
