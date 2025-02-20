package getziti

import (
	"fmt"

	"ztna-core/ztna/logtrace"
	c "ztna-core/ztna/ztna/constants"
)

func InstallZiti(targetVersion, targetOS, targetArch, binDir string, verbose bool) error {
	logtrace.LogWithFunctionName()
	fmt.Println("Attempting to install '" + c.ZITI + "' version: " + targetVersion)
	return FindVersionAndInstallGitHubRelease(
		c.OpenZitiOrg, c.ZITI, c.ZITI, targetOS, targetArch, binDir, targetVersion, verbose)
}

func InstallZrok(targetVersion, targetOS, targetArch, binDir string, verbose bool) error {
	logtrace.LogWithFunctionName()
	fmt.Println("Attempting to install '" + c.ZROK + "' version: " + targetVersion)
	return FindVersionAndInstallGitHubRelease(
		c.OpenZitiOrg, c.ZROK, c.ZROK, targetOS, targetArch, binDir, targetVersion, verbose)
}

func InstallCaddy(targetVersion, targetOS, targetArch, binDir string, verbose bool) error {
	logtrace.LogWithFunctionName()
	fmt.Println("Attempting to install '" + c.Caddy + "' version: " + targetVersion)
	return FindVersionAndInstallGitHubRelease(
		c.CaddyOrg, c.Caddy, c.Caddy, targetOS, targetArch, binDir, targetVersion, verbose)
}
