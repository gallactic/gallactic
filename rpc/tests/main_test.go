package tests

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/proposal"

	"github.com/gallactic/gallactic/common"
)

var tChainName string
var tWorkingDir string
var tGenesis *proposal.Genesis
var tConfig *config.Config

func startServer(wg *sync.WaitGroup) *exec.Cmd {
	tChainName = "test-chain" + common.RandomHex(4)
	tWorkingDir = "/tmp/" + tChainName

	cmd := exec.Command("gallactic", "init", "-w", tWorkingDir, "-n", tChainName)
	cmd.Run()

	tGenesis, _ = proposal.LoadFromFile(tWorkingDir + "/genesis.json")
	tConfig, _ = config.LoadFromFile(tWorkingDir + "/config.toml")

	cmd = exec.Command("gallactic", "start", "-w", tWorkingDir)
	wg.Add(1)
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	time.Sleep(time.Second * 1)
	return cmd
}

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	cmd := startServer(&wg)

	exitCode := m.Run()

	cmd.Process.Signal(syscall.SIGINT)

	// waiting for gallactic to exit
	wg.Wait()
	cmd.Wait()

	os.Exit(exitCode)
}
