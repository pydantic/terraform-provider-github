package github

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// GitHub returns lowercase logins on read, so state after a refresh holds
// "douwem" while config holds "DouweM". The plan must not touch members.
func TestGithubTeamMembersCaseOnlyDiff(t *testing.T) {
	r := resourceGithubTeamMembers()
	// The legacy team diffs call the API. This test covers the schema only.
	r.CustomizeDiff = nil

	state := r.Data(nil)
	state.SetId("123")
	if err := state.Set("team_slug", "core"); err != nil {
		t.Fatal(err)
	}
	if err := state.Set("team_id", "123"); err != nil {
		t.Fatal(err)
	}
	if err := state.Set("members", []any{
		map[string]any{"username": "douwem", "role": "member"},
		map[string]any{"username": "viicos", "role": "maintainer"},
	}); err != nil {
		t.Fatal(err)
	}

	config := func(members ...map[string]any) *terraform.ResourceConfig {
		ms := make([]any, len(members))
		for i, m := range members {
			ms[i] = m
		}
		return terraform.NewResourceConfigRaw(map[string]any{"team_slug": "core", "members": ms})
	}

	t.Run("case only", func(t *testing.T) {
		diff, err := r.Diff(context.Background(), state.State(), config(
			map[string]any{"username": "DouweM", "role": "member"},
			map[string]any{"username": "Viicos", "role": "maintainer"},
		), nil)
		if err != nil {
			t.Fatal(err)
		}
		if diff != nil && !diff.Empty() {
			t.Fatalf("expected no diff, got %#v", diff.Attributes)
		}
	})

	t.Run("role change", func(t *testing.T) {
		diff, err := r.Diff(context.Background(), state.State(), config(
			map[string]any{"username": "DouweM", "role": "maintainer"},
			map[string]any{"username": "Viicos", "role": "maintainer"},
		), nil)
		if err != nil {
			t.Fatal(err)
		}
		if diff == nil || diff.Empty() {
			t.Fatal("expected a members diff for a role change")
		}
	})

	t.Run("member removed", func(t *testing.T) {
		diff, err := r.Diff(context.Background(), state.State(), config(
			map[string]any{"username": "Viicos", "role": "maintainer"},
		), nil)
		if err != nil {
			t.Fatal(err)
		}
		if diff == nil || diff.Attributes["members.#"] == nil || diff.Attributes["members.#"].New != "1" {
			t.Fatalf("expected members.# to become 1, got %#v", diff)
		}
	})
}
