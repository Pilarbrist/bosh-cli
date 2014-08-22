package compile

import (
	"fmt"
	"os"
	"path"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type PackageCompiler interface {
	Compile(*bmrel.Package) error
}

type packageCompiler struct {
	runner              boshsys.CmdRunner
	packagesDir         string
	fileSystem          boshsys.FileSystem
	compressor          boshcmd.Compressor
	blobstore           boshblob.Blobstore
	compiledPackageRepo CompiledPackageRepo
	ui                  bmui.UI
}

func NewPackageCompiler(
	runner boshsys.CmdRunner,
	packagesDir string,
	fileSystem boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	compiledPackageRepo CompiledPackageRepo,
	ui bmui.UI,
) PackageCompiler {
	return &packageCompiler{
		runner:              runner,
		packagesDir:         packagesDir,
		fileSystem:          fileSystem,
		compressor:          compressor,
		blobstore:           blobstore,
		compiledPackageRepo: compiledPackageRepo,
		ui:                  ui,
	}
}

func (pc *packageCompiler) Compile(pkg *bmrel.Package) error {
	_, found, err := pc.compiledPackageRepo.Find(*pkg)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Attempting to find compiled package `%s'", pkg.Name))
	}
	if found {
		pc.ui.Sayln(fmt.Sprintf("Skipping compilation of package `%s'", pkg.Name))
		return nil
	}

	packageSrcDir := pkg.ExtractedPath

	installDir := path.Join(pc.packagesDir, pkg.Name)
	err = pc.fileSystem.MkdirAll(installDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating package install dir")
	}

	defer pc.fileSystem.RemoveAll(pc.packagesDir)

	if !pc.fileSystem.FileExists(path.Join(packageSrcDir, "packaging")) {
		return bosherr.New("Packaging script for package `%s' not found", pkg.Name)
	}

	cmd := boshsys.Command{
		Name: "bash",
		Args: []string{"-x", "packaging"},
		Env: map[string]string{
			"BOSH_COMPILE_TARGET":  packageSrcDir,
			"BOSH_INSTALL_TARGET":  installDir,
			"BOSH_PACKAGE_NAME":    pkg.Name,
			"BOSH_PACKAGE_VERSION": pkg.Version,
			"BOSH_PACKAGES_DIR":    pc.packagesDir,
		},
		WorkingDir: packageSrcDir,
	}

	pc.ui.Sayln(fmt.Sprintf("Compiling package `%s'", pkg.Name))

	_, _, _, err = pc.runner.RunComplexCommand(cmd)
	if err != nil {
		return bosherr.WrapError(err, "Compiling package")
	}

	tarball, err := pc.compressor.CompressFilesInDir(installDir)
	if err != nil {
		return bosherr.WrapError(err, "Compressing compiled package")
	}
	defer pc.compressor.CleanUp(tarball)

	blobID, fingerprint, err := pc.blobstore.Create(tarball)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	err = pc.compiledPackageRepo.Save(*pkg, CompiledPackageRecord{
		BlobID:      blobID,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return bosherr.WrapError(err, "Saving compiled package")
	}

	return nil
}