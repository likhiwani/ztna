/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package agentcli

import (
	logtrace "ztna-core/ztna/logtrace"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	goversion "rsc.io/goversion/version"
)

type AgentPsAction struct {
	AgentOptions
}

// NewPsCmd Pss a command object for the "Ps" command
func NewPsCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &AgentPsAction{
		AgentOptions: AgentOptions{
			CommonOptions: p(),
		},
	}

	cmd := &cobra.Command{
		Use:   "ps",
		Short: "Show Ziti process info",
		Long:  "Show information about currently running Ziti processes",
		Run: func(cmd *cobra.Command, args []string) {
			action.Cmd = cmd
			action.Args = args
			err := action.Run()
			cmdhelper.CheckErr(err)
		},
	}

	return cmd
}

// Run implements this command
func (self *AgentPsAction) Run() error {
	logtrace.LogWithFunctionName()
	fmt.Println("ps is running")
	ps := FindAll()

	var maxPID, maxPPID, maxExec, maxVersion int
	for i, p := range ps {

		fmt.Println("p.Path is: " + p.Path)

		ps[i].BuildVersion = shortenVersion(p.BuildVersion)
		maxPID = max(maxPID, len(strconv.Itoa(p.PID)))
		maxPPID = max(maxPPID, len(strconv.Itoa(p.PPID)))
		maxExec = max(maxExec, len(p.Exec))
		maxVersion = max(maxVersion, len(ps[i].BuildVersion))

	}

	for _, p := range ps {
		buf := bytes.NewBuffer(nil)
		pid := strconv.Itoa(p.PID)
		fmt.Fprint(buf, pad(pid, maxPID))
		fmt.Fprint(buf, " ")

		ppid := strconv.Itoa(p.PPID)
		fmt.Fprint(buf, pad(ppid, maxPPID))
		fmt.Fprint(buf, " ")

		fmt.Fprint(buf, pad(p.Exec, maxExec))

		if p.Agent {
			fmt.Fprint(buf, "*")
		} else {
			fmt.Fprint(buf, " ")
		}
		fmt.Fprint(buf, " ")

		fmt.Fprint(buf, pad(p.BuildVersion, maxVersion))
		fmt.Fprint(buf, " ")

		fmt.Fprint(buf, p.Path)
		fmt.Fprintln(buf)
		if _, err := buf.WriteTo(os.Stdout); err != nil {
			return err
		}
	}

	return nil
}

var develRe = regexp.MustCompile(`devel\s+\+\w+`)

func shortenVersion(v string) string {
	logtrace.LogWithFunctionName()
	if !strings.HasPrefix(v, "devel") {
		return v
	}
	results := develRe.FindAllString(v, 1)
	if len(results) == 0 {
		return v
	}
	return results[0]
}

func pad(s string, total int) string {
	logtrace.LogWithFunctionName()
	if len(s) >= total {
		return s
	}
	return s + strings.Repeat(" ", total-len(s))
}

func max(i, j int) int {
	logtrace.LogWithFunctionName()
	if i > j {
		return i
	}
	return j
}

// P represents a Go process.
type P struct {
	PID          int
	PPID         int
	Exec         string
	Path         string
	BuildVersion string
	Agent        bool
}

// FindAll returns all the Ziti processes currently running on this host.
func FindAll() []P {
	logtrace.LogWithFunctionName()
	pss, err := process.Processes()
	// fmt.Println("FindAll, err is: %s", err)

	if err != nil {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(pss))
	found := make(chan P)

	for _, pr := range pss {
		pr := pr
		go func() {
			defer wg.Done()

			path, version, agent, ok, err := isGo(pr)
			if err != nil {
				// TODO(jbd): Return a list of errors.
				return
			}
			if !ok {
				return
			}
			ppid, err := pr.Ppid()
			if err != nil {
				return
			}
			name, err := pr.Name()
			if err != nil {
				return
			}
			if isZiti(pr) {
				found <- P{
					PID:          int(pr.Pid),
					PPID:         int(ppid),
					Exec:         name,
					Path:         path,
					BuildVersion: version,
					Agent:        agent,
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(found)
	}()
	var results []P
	for p := range found {
		results = append(results, p)
	}
	return results
}

// isGo looks up the runtime.buildVersion symbol
// in the process' binary and determines if the process
// if a Go process or not. If the process is a Go process,
// it reports PID, binary name and full path of the binary.
func isGo(pr *process.Process) (path, version string, agent, ok bool, err error) {
	logtrace.LogWithFunctionName()
	if pr.Pid == 0 {
		// ignore system process
		return
	}
	path, err = pr.Exe()
	if err != nil {
		return
	}
	var versionInfo goversion.Version
	versionInfo, err = goversion.ReadExe(path)
	if err != nil {
		return
	}
	ok = true
	version = versionInfo.Release
	pidfile, err := PIDFile(int(pr.Pid))
	if err == nil {
		_, err := os.Stat(pidfile)
		agent = err == nil
	}
	return path, version, agent, ok, nil
}

const gopsConfigDirEnvKey = "GOPS_CONFIG_DIR"

func ConfigDir() (string, error) {
	logtrace.LogWithFunctionName()
	if configDir := os.Getenv(gopsConfigDirEnvKey); configDir != "" {
		return configDir, nil
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gops"), nil
	}
	homeDir := guessUnixHomeDir()
	if homeDir == "" {
		return "", errors.New("unable to get current user home directory: os/user lookup failed; $HOME is empty")
	}
	return filepath.Join(homeDir, ".config", "gops"), nil
}

func guessUnixHomeDir() string {
	logtrace.LogWithFunctionName()
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}

func PIDFile(pid int) (string, error) {
	logtrace.LogWithFunctionName()
	gopsdir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%d", gopsdir, pid), nil
}

func isZiti(pr *process.Process) (ok bool) {
	logtrace.LogWithFunctionName()
	name, err := pr.Name()
	if err != nil {
		return false
	}
	return strings.HasPrefix(name, "ziti")
}
