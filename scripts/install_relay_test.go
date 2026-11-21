//go:build linux

package scripts

import (
	"os/exec"
	"testing"
)

func TestInstallRelay(t *testing.T) {
	for _, scenario := range []string{"valid", "wrong-checksum", "missing-archive", "missing-checksum"} {
		t.Run(scenario, func(t *testing.T) {
			cmd := exec.Command("bash", "testdata/install-relay.sh", t.TempDir(), scenario)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%v\n%s", err, output)
			}
		})
	}
}
