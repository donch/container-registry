package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func updateK8sVersion(path string, tmpFileName string) ([]byte, error) {
	lines, err := readFromTempFile(tmpFileName)
	if err != nil {
		return nil, err
	}

	switch path {
	case "bases/pre.yaml", "bases/gstg.yaml", "bases/gprd.yaml":
		for i, line := range lines {
			if strings.Contains(line, "registry_version") {
				lines[i] = fmt.Sprintf("        registry_version: %s", release.version)
				break
			}
		}
	default:
		return nil, fmt.Errorf("unexpected file path %q", path)
	}

	out := strings.Join(lines, "\n")
	updatedFile, err := loadFileChange(tmpFileName, out)
	if err != nil {
		return nil, err
	}

	return updatedFile, nil
}

func updateGDKVersion(path string, tmpFileName string) ([]byte, error) {
	lines, err := readFromTempFile(tmpFileName)
	if err != nil {
		return nil, err
	}

	if path == "support/docker-registry" {
		for i, line := range lines {
			if strings.Contains(line, "registry_image:-registry.gitlab.com/gitlab-org/build/cng") {
				lines[i] = fmt.Sprintf("      \"${registry_image:-registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:%s}\"", release.version)
				break
			}
		}
	}

	if path == "lib/gdk/config.rb" {
		for i, line := range lines {
			if strings.Contains(line, "registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:") {
				lines[i+1] = fmt.Sprintf("        '%s'", release.version)
				break
			}
		}
	}

	if path == "spec/lib/gdk/config_spec.rb" {
		for i, line := range lines {
			if strings.Contains(line, "registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:") {
				lines[i] = fmt.Sprintf("         expect(config.registry.image).to eq('registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:%s')", release.version)
				break
			}
		}
	}

	if path == "gdk.example.yml" {
		for i, line := range lines {
			if strings.Contains(line, "registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:") {
				lines[i] = fmt.Sprintf("  image: registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:%s", release.version)
				break
			}
		}
	}

	out := strings.Join(lines, "\n")
	updatedFile, err := loadFileChange(tmpFileName, out)
	if err != nil {
		return nil, err
	}

	return updatedFile, nil
}

func loadFileChange(tempFilename string, output string) ([]byte, error) {
	err := os.WriteFile(tempFilename, []byte(output), 0644)
	if err != nil {
		return nil, err
	}

	cng, err := os.ReadFile(tempFilename)
	if err != nil {
		return nil, err
	}

	err = os.Remove(tempFilename)
	if err != nil {
		return nil, err
	}

	return cng, err
}

func readFromTempFile(tempFilename string) ([]string, error) {
	f, err := os.ReadFile(tempFilename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(f), "\n")
	return lines, err
}

func createAndCopyRepositoryFile(dec []byte) (string, error) {
	f, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write(dec); err != nil {
		return "", err
	}

	f.Seek(0, 0)

	return f.Name(), err
}
