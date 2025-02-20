package test_resources

import (
	"embed"
	"io/fs"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/fablab/resources"
)

//go:embed terraform
var terraformResource embed.FS

func TerraformResources() fs.FS {
	logtrace.LogWithFunctionName()
	return resources.SubFolder(terraformResource, resources.Terraform)
}
