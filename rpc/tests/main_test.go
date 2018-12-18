package tests

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/proposal"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"google.golang.org/grpc"
)

var tChainName string
var tWorkingDir string
var tGenesis *proposal.Genesis
var tConfig *config.Config

func startServer(done chan struct{}) *exec.Cmd {
	tChainName = "test-chain" + common.RandomHex(4)
	tWorkingDir = "/tmp/" + tChainName

	cmd := exec.Command("gallactic", "init", "-w", tWorkingDir, "-n", tChainName)
	cmd.Run()

	tGenesis, _ = proposal.LoadFromFile(tWorkingDir + "/genesis.json")
	tConfig, _ = config.LoadFromFile(tWorkingDir + "/config.toml")

	cmd = exec.Command("gallactic", "start", "-w", tWorkingDir)

	go func() {
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		done <- struct{}{}
	}()

	//Just for wait to ensure command executed and instance object is ready
	addr := tConfig.GRPC.ListenAddress
	conn, errIns := grpc.Dial(addr, grpc.WithInsecure())
	if errIns != nil {
		panic(errIns)
	}
	grpcBCClient := pb.NewBlockChainClient(conn)

	for {
		_, getchain_err := grpcBCClient.GetChainID(context.Background(), &pb.Empty{})
		if getchain_err == nil {
			break
		}
		time.Sleep(100)
	}
	return cmd
}

func TestMain(m *testing.M) {
	done := make(chan struct{})
	cmd := startServer(done)

	exitCode := m.Run()

	cmd.Process.Signal(syscall.SIGINT)

	// waiting for gallactic to exit
	cmd.Wait()
	<-done

	os.Exit(exitCode)
}
