package discovery

import (
	"testing"

	"repoctr/pkg/models"
)

func TestHierarchyBuilder_Build(t *testing.T) {
	builder := NewHierarchyBuilder()

	projects := []*models.Project{
		{Name: "root", Path: "."},
		{Name: "lib-a", Path: "packages/lib-a"},
		{Name: "lib-b", Path: "packages/lib-b"},
		{Name: "app", Path: "apps/frontend"},
		{Name: "packages", Path: "packages"},
	}

	roots := builder.Build(projects)

	// Should have one root project
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}

	root := roots[0]
	if root.Name != "root" {
		t.Errorf("expected root name 'root', got %q", root.Name)
	}

	// Root should have 2 children: packages and apps/frontend
	if len(root.Children) != 2 {
		t.Errorf("expected 2 children of root, got %d", len(root.Children))
	}

	// Find packages node
	var packagesNode *models.Project
	for _, child := range root.Children {
		if child.Name == "packages" {
			packagesNode = child
			break
		}
	}

	if packagesNode == nil {
		t.Fatal("expected to find packages node")
	}

	// packages should have 2 children: lib-a and lib-b
	if len(packagesNode.Children) != 2 {
		t.Errorf("expected 2 children of packages, got %d", len(packagesNode.Children))
	}
}

func TestHierarchyBuilder_Flatten(t *testing.T) {
	builder := NewHierarchyBuilder()

	root := &models.Project{
		Name: "root",
		Path: ".",
		Children: []*models.Project{
			{Name: "child1", Path: "child1"},
			{
				Name: "child2",
				Path: "child2",
				Children: []*models.Project{
					{Name: "grandchild", Path: "child2/grandchild"},
				},
			},
		},
	}

	flat := builder.Flatten([]*models.Project{root})

	if len(flat) != 4 {
		t.Fatalf("expected 4 projects, got %d", len(flat))
	}

	// Check order (should be depth-first)
	names := make([]string, len(flat))
	for i, p := range flat {
		names[i] = p.Name
	}

	expected := []string{"root", "child1", "child2", "grandchild"}
	for i, exp := range expected {
		if names[i] != exp {
			t.Errorf("position %d: expected %q, got %q", i, exp, names[i])
		}
	}
}

func TestHierarchyBuilder_EmptyInput(t *testing.T) {
	builder := NewHierarchyBuilder()

	roots := builder.Build(nil)
	if roots != nil {
		t.Errorf("expected nil for empty input, got %v", roots)
	}

	roots = builder.Build([]*models.Project{})
	if roots != nil {
		t.Errorf("expected nil for empty slice, got %v", roots)
	}
}

func TestHierarchyBuilder_NoParents(t *testing.T) {
	builder := NewHierarchyBuilder()

	projects := []*models.Project{
		{Name: "project-a", Path: "a"},
		{Name: "project-b", Path: "b"},
		{Name: "project-c", Path: "c"},
	}

	roots := builder.Build(projects)

	// All should be roots since none is a parent of another
	if len(roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(roots))
	}
}
