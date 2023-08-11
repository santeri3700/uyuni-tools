package install

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
	"github.com/uyuni-project/uyuni-tools/uyuniadm/shared/podman"
)

func waitForSystemStart(globalFlags *types.GlobalFlags, flags *InstallFlags) {
	// Setup the systemd service configuration options
	image := fmt.Sprintf("%s:%s", flags.Image.Name, flags.Image.Tag)
	podman.GenerateSystemdService(flags.TZ, image, flags.Podman.Args, globalFlags.Verbose)

	log.Println("Waiting for the server to start...")
	// Start the service
	if err := exec.Command("systemctl", "enable", "--now", "uyuni-server").Run(); err != nil {
		log.Fatalf("Failed to enable uyuni-server systemd service: %s\n", err)
	}

	utils.WaitForServer()
}

func pullImage(flags *InstallFlags) {
	image := fmt.Sprintf("%s:%s", flags.Image.Name, flags.Image.Tag)
	log.Printf("Running podman pull %s\n", image)
	cmd := exec.Command("podman", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to pull image: %s\n", err)
	}
}

func installForPodman(globalFlags *types.GlobalFlags, flags *InstallFlags, cmd *cobra.Command, args []string) {
	pullImage(flags)

	waitForSystemStart(globalFlags, flags)

	env := map[string]string{}
	if flags.Cert.UseExisting {
		// TODO Get existing certificates path and mount them
		// Set CA_CERT, SERVER_CERT, SERVER_KEY or run the rhn-ssl-check tool in a container
		// The SERVER_CERT needs to get the intermediate keys
	} else {
		env["CERT_O"] = flags.Cert.Org
		env["CERT_OU"] = flags.Cert.OU
		env["CERT_CITY"] = flags.Cert.City
		env["CERT_STATE"] = flags.Cert.State
		env["CERT_COUNTRY"] = flags.Cert.Country
		env["CERT_EMAIL"] = flags.Cert.Email
		env["CERT_CNAMES"] = strings.Join(append([]string{args[0]}, flags.Cert.Cnames...), ",")
		env["CERT_PASS"] = flags.Cert.Password
	}

	runSetup(globalFlags, flags, args[0], env)
}
