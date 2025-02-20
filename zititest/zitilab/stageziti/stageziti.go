package stageziti

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"ztna-core/ztna/logtrace"

	"ztna-core/ztna/common/getziti"
	"ztna-core/ztna/ztna/util"

	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func StageZitiOnce(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	op := "install.ziti-"
	if version == "" {
		op += "local"
	} else {
		op += version
	}

	return run.DoOnce(op, func() error {
		return StageZiti(run, component, version, source)
	})
}

func StageZrokOnce(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	op := "install.zrok-"
	if version == "" {
		op += "local"
	} else {
		op += version
	}

	return run.DoOnce(op, func() error {
		return StageZrok(run, component, version, source)
	})
}

func StageCaddyOnce(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	op := "install.caddy-"
	if version == "" {
		op += "local"
	} else {
		op += version
	}

	return run.DoOnce(op, func() error {
		return StageCaddy(run, component, version, source)
	})
}

func StageZitiEdgeTunnelOnce(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	op := "install.ziti-edge-tunnel-"
	if version == "" {
		op += "local"
	} else {
		op += version
	}

	return run.DoOnce(op, func() error {
		return StageZitiEdgeTunnel(run, component, version, source)
	})
}

func StageZiti(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	return StageExecutable(run, "ziti", component, version, source, func() error {
		return getziti.InstallZiti(version, "linux", "amd64", run.GetBinDir(), false)
	})
}

func StageZrok(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	return StageExecutable(run, "zrok", component, version, source, func() error {
		return getziti.InstallZrok(version, "linux", "amd64", run.GetBinDir(), false)
	})
}

func StageCaddy(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	return StageExecutable(run, "caddy", component, version, source, func() error {
		return getziti.InstallCaddy(version, "linux", "amd64", run.GetBinDir(), false)
	})
}

func StageLocalOnce(run model.Run, executable string, component *model.Component, source string) error {
	logtrace.LogWithFunctionName()
	op := fmt.Sprintf("install.%s-local", executable)
	return run.DoOnce(op, func() error {
		return StageExecutable(run, executable, component, "", source, func() error {
			return fmt.Errorf("unable to fetch %s, as it a local-only application", executable)
		})
	})
}

func StageExecutable(run model.Run, executable string, component *model.Component, version string, source string, fallbackF func() error) error {
	logtrace.LogWithFunctionName()
	fileName := executable
	if version != "" {
		fileName += "-" + version
	}

	target := filepath.Join(run.GetBinDir(), fileName)
	if version == "" || version == "latest" {
		_ = os.Remove(target)
	}

	envVar := strings.ToUpper(executable) + "_PATH"

	if version == "" {
		if source != "" {
			logrus.Infof("[%s] => [%s]", source, target)
			return util.CopyFile(source, target)
		}
		if envSource, found := component.GetStringVariable(envVar); found {
			logrus.Infof("[%s] => [%s]", envSource, target)
			return util.CopyFile(envSource, target)
		}
		if zitiPath, err := exec.LookPath(executable); err == nil {
			logrus.Infof("[%s] => [%s]", zitiPath, target)
			return util.CopyFile(zitiPath, target)
		}
		return fmt.Errorf("%s binary not found in path, no path provided and no %s env variable set", executable, envVar)
	}

	found, err := run.FileExists(filepath.Join(model.BuildKitDir, model.BuildBinDir, fileName))
	if err != nil {
		return err
	}

	if found {
		logrus.Infof("%s already present, not downloading again", target)
		return nil
	}

	logrus.Infof("%s not present, attempting to fetch", target)

	return fallbackF()
}

func StageZitiEdgeTunnel(run model.Run, component *model.Component, version string, source string) error {
	logtrace.LogWithFunctionName()
	fileName := "ziti-edge-tunnel"
	if version != "" {
		fileName += "-" + version
	}

	target := filepath.Join(run.GetBinDir(), fileName)
	if version == "" || version == "latest" {
		_ = os.Remove(target)
	}

	if version == "" {
		if source != "" {
			logrus.Infof("[%s] => [%s]", source, target)
			return util.CopyFile(source, target)
		}
		if envSource, found := component.GetStringVariable("ziti-edge-tunnel.path"); found {
			logrus.Infof("[%s] => [%s]", envSource, target)
			return util.CopyFile(envSource, target)
		}
		if zitiPath, err := exec.LookPath("ziti-edge-tunnel"); err == nil {
			logrus.Infof("[%s] => [%s]", zitiPath, target)
			return util.CopyFile(zitiPath, target)
		}
		return errors.New("ziti-edge-tunnel binary not found in path, no path provided and no ziti-edge-tunnel.path env variable set")
	}

	found, err := run.FileExists(filepath.Join(model.BuildKitDir, model.BuildBinDir, fileName))
	if err != nil {
		return err
	}

	if found {
		logrus.Infof("%s already present, not downloading again", target)
		return nil
	}
	logrus.Infof("%s not present, attempting to fetch", target)

	return getziti.InstallZitiEdgeTunnel(version, "linux", "amd64", run.GetBinDir(), false)
}
