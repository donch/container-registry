package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func updateK8sVersion(tmp string) ([]byte, error) {
	f, err := os.ReadFile(tmp)
	if err != nil {
		return nil, err
	}

	m := regexp.MustCompile("registry_version: (.*)")
	output := m.ReplaceAllString(string(f), fmt.Sprintf("registry_version: %s", release.version))

	err = os.WriteFile(tmp, []byte(output), 0644)
	if err != nil {
		return nil, err
	}
	updatedFile, err := os.ReadFile(tmp)
	if err != nil {
		return nil, err
	}

	err = os.Remove(tmp)
	if err != nil {
		return nil, err
	}

	return updatedFile, err
}

func updateGDKVersion(tmp string) ([]byte, error) {
	f, err := os.ReadFile(tmp)
	if err != nil {
		return nil, err
	}

	m := regexp.MustCompile(`gitlab-container-registry:(.[^"'" "\n)}]+)`)
	output := m.ReplaceAllString(string(f), fmt.Sprintf("gitlab-container-registry:%s", release.version))

	err = os.WriteFile(tmp, []byte(output), 0644)
	if err != nil {
		return nil, err
	}
	updatedFile, err := os.ReadFile(tmp)
	if err != nil {
		return nil, err
	}

	err = os.Remove(tmp)
	if err != nil {
		return nil, err
	}

	return updatedFile, err
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
