package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"go.uber.org/zap"
)

func main() {
	// configure zap to log to a file, this is good enough for local debugging
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"chained-cni.log"}
	logger, err := cfg.Build()
	if err != nil {
		os.Exit(2)
	}
	zap.ReplaceGlobals(logger)
	skel.PluginMain(cmdAdd,
		nil,
		cmdDel,
		version.All,
		"Chained CNI example")
}

func parsePrevResult(n *types.NetConf) (*types.NetConf, error) {
	if n.RawPrevResult != nil {
		resultBytes, err := json.Marshal(n.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not serialize prevResult: %v", err)
		}
		res, err := version.NewResult(n.CNIVersion, resultBytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse prevResult: %v", err)
		}
		n.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
	}

	return n, nil
}

func cmdAdd(args *skel.CmdArgs) error {

	netConf := &types.NetConf{
		CNIVersion:    "",
		Name:          "",
		Type:          "",
		Capabilities:  map[string]bool{},
		IPAM:          types.IPAM{},
		DNS:           types.DNS{},
		RawPrevResult: map[string]interface{}{},
		PrevResult:    nil,
	}
	err := json.Unmarshal(args.StdinData, &netConf)
	if err != nil {
		zap.L().Error("Un-marshall error", zap.Error(err))
	}

	zap.L().Info("cmdADD with args", zap.Any("CNI args", args))

	n, err := parsePrevResult(netConf)
	if err != nil {
		zap.L().Error("parsePrevResult error", zap.Error(err))
	}
	if n.PrevResult != nil {
		zap.L().Info("CNI Previous result", zap.Any("prev-res", n.PrevResult))
	}
	res, err := current.NewResultFromResult(n.PrevResult)
	if err != nil {
		zap.L().Error("NewResultFromResult error", zap.Error(err))
	}

	fmt.Println("Result returning", res)
	_, err = os.Stdout.Write(args.StdinData)
	return err
}

func cmdDel(args *skel.CmdArgs) error {

	zap.L().Info("cmdDel with args", zap.Any("CNI args", args))

	return nil
}
