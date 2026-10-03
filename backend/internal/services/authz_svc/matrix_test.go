package authz_svc

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const migrationPath = "../../database/migrations/000017_add_rbac.up.sql"

// expectedMatrix is the role/permission matrix from .plan/rbac.md.
var expectedMatrix = map[string][]string{
	"system_admin":  AllPermissions,
	"project_admin": {PackCreate, UserLookup},
	"user":          {},
	"viewer":        {PackConsume, PackView, PackLink},
	"editor": {
		PackConsume, PackView, PackLink,
		PackModAdd, PackModRemove, PackModUpdate, PackModConfigure,
		PackUpdatesCheck, PackUsersView,
	},
	"owner": {
		PackConsume, PackView, PackLink,
		PackModAdd, PackModRemove, PackModUpdate, PackModConfigure,
		PackUpdatesCheck, PackUsersView,
		PackMigrate, PackRehash, PackInfoEdit, PackPublish, PackVisibility, PackArchive, PackUsersManage,
	},
}

func readMigration(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(b)
}

func sorted(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}

// seededPermissions parses name -> resource from the INSERT INTO permissions block.
func seededPermissions(t *testing.T, sql string) map[string]string {
	t.Helper()
	start := strings.Index(sql, "INSERT INTO permissions")
	end := strings.Index(sql[start:], ";")
	rows := regexp.MustCompile(`\('([a-z.]+)', '(pack|global)'`).FindAllStringSubmatch(sql[start:start+end], -1)

	out := map[string]string{}
	for _, r := range rows {
		out[r[1]] = r[2]
	}
	return out
}

// seededRolePermissions evaluates the role_permissions seed for each role.
func seededRolePermissions(t *testing.T, sql string, perms map[string]string) map[string][]string {
	t.Helper()
	start := strings.Index(sql, "INSERT INTO role_permissions")
	end := strings.Index(sql[start:], ";")
	block := sql[start : start+end]

	all := make([]string, 0, len(perms))
	pack := []string{}
	for name, resource := range perms {
		all = append(all, name)
		if resource == "pack" {
			pack = append(pack, name)
		}
	}

	out := map[string][]string{"user": {}}
	if strings.Contains(block, "r.name = 'system_admin'") {
		out["system_admin"] = all
	}
	if strings.Contains(block, "r.name = 'owner' AND p.resource = 'pack'") {
		out["owner"] = pack
	}
	for _, m := range regexp.MustCompile(`r\.name = '(\w+)' AND p\.name IN \(([^)]*)\)`).FindAllStringSubmatch(block, -1) {
		out[m[1]] = regexp.MustCompile(`'([a-z.]+)'`).FindAllString(m[2], -1)
		for i, v := range out[m[1]] {
			out[m[1]][i] = strings.Trim(v, "'")
		}
	}
	return out
}

func TestSeededPermissionsMatchCode(t *testing.T) {
	perms := seededPermissions(t, readMigration(t))

	names := make([]string, 0, len(perms))
	for n, resource := range perms {
		names = append(names, n)
		if IsGlobal(n) != (resource == "global") {
			t.Errorf("%s: resource %q disagrees with IsGlobal=%v", n, resource, IsGlobal(n))
		}
	}

	if !reflect.DeepEqual(sorted(names), sorted(AllPermissions)) {
		t.Fatalf("seeded permissions differ from AllPermissions\nseeded: %v\ncode:   %v", sorted(names), sorted(AllPermissions))
	}
}

func TestSeededRolesMatchPlanMatrix(t *testing.T) {
	sql := readMigration(t)
	got := seededRolePermissions(t, sql, seededPermissions(t, sql))

	for role, want := range expectedMatrix {
		if !reflect.DeepEqual(sorted(got[role]), sorted(want)) {
			t.Errorf("role %s\n got:  %v\n want: %v", role, sorted(got[role]), sorted(want))
		}
	}
	if len(got) != len(expectedMatrix) {
		t.Errorf("seeded %d roles, plan has %d", len(got), len(expectedMatrix))
	}
}

func TestPackRolesHoldOnlyPackPermissions(t *testing.T) {
	for _, role := range []string{"viewer", "editor", "owner"} {
		for _, p := range expectedMatrix[role] {
			if IsGlobal(p) {
				t.Errorf("pack role %s holds global permission %s", role, p)
			}
		}
	}
}

// The permission matrix as decisions: a pack role grants on its own pack only,
// a system role grants on every pack.
func TestMatrixDecisions(t *testing.T) {
	set := func(role string) Set { return NewSet(expectedMatrix[role]...) }

	// viewer cannot edit mods, editor can, owner can archive, editor cannot
	if DecidePack(false, false, nil, set("viewer"), PackModAdd) == nil {
		t.Error("viewer must not add mods")
	}
	if err := DecidePack(false, false, nil, set("editor"), PackModAdd); err != nil {
		t.Errorf("editor must add mods: %v", err)
	}
	if DecidePack(false, false, nil, set("editor"), PackArchive) == nil {
		t.Error("editor must not archive")
	}
	if err := DecidePack(false, false, nil, set("owner"), PackArchive); err != nil {
		t.Errorf("owner must archive: %v", err)
	}

	// project_admin creates packs but gets no pack access by itself
	if !DecideGlobal(false, set("project_admin"), PackCreate) {
		t.Error("project_admin must create packs")
	}
	if DecidePack(false, false, set("project_admin"), nil, PackView) == nil {
		t.Error("project_admin must not view arbitrary packs")
	}

	// system_admin reaches every pack and every global permission
	for _, p := range AllPermissions {
		if IsGlobal(p) {
			if !DecideGlobal(false, set("system_admin"), p) {
				t.Errorf("system_admin lacks %s", p)
			}
		} else if err := DecidePack(false, false, set("system_admin"), nil, p); err != nil {
			t.Errorf("system_admin lacks %s on a pack: %v", p, err)
		}
	}

	// plain user holds nothing
	for _, p := range AllPermissions {
		if DecideGlobal(false, set("user"), p) {
			t.Errorf("user must not hold %s", p)
		}
	}
}
