package pcf_pipelines_test

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var set_eu_regexp = regexp.MustCompile("^set -eu$")
var shebang_regexp = regexp.MustCompile("^#!/bin/bash$")

var _ = Describe("Tasks", func() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working dir: %s", err)
	}

	root := filepath.Dir(cwd)

	var taskPaths []string
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fi, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if !fi.IsDir() && fi.Mode().IsRegular() && strings.Contains(path, "tasks") && strings.HasSuffix(path, ".yml") {
			relTaskPath, err := filepath.Rel(cwd, path)
			if err != nil {
				return err
			}
			taskPaths = append(taskPaths, relTaskPath)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk: %s", err)
	}

	for _, path := range taskPaths {
		taskPath := path

		var task atc.TaskConfig

		configBytes, err := ioutil.ReadFile(taskPath)
		if err != nil {
			log.Fatalf("failed to load task file: %s", err)
		}

		err = yaml.Unmarshal(configBytes, &task)
		if err != nil {
			log.Fatalf("failed to unmarshal task at %s: %s", taskPath, err)
		}

		if strings.HasPrefix(task.Run.Path, "pcf-pipelines") {
			failMessage := fmt.Sprintf(`
Found error with the following task:
%s
`, taskPath)

			Context(fmt.Sprintf("task at %s", taskPath), func() {
				It("includes a pcf-pipelines input", func() {
					match, err := ContainElement(atc.TaskInputConfig{Name: "pcf-pipelines", Path: ""}).Match(task.Inputs)
					Expect(err).NotTo(HaveOccurred())

					if !match {
						Fail(fmt.Sprintf("%s\n%s", failMessage, "Fix it by adding the following input:\n- name: pcf-pipelines"))
					}
				})

				It("does not pass any args", func() {
					match, err := HaveLen(0).Match(task.Run.Args)
					Expect(err).NotTo(HaveOccurred())
					if !match {
						bs, err := yaml.Marshal(task.Run.Args)
						Expect(err).NotTo(HaveOccurred())

						Fail(fmt.Sprintf("%s\nFix it by removing this:\nargs:\n%s", failMessage, bs))
					}
				})

				It("points to a task.sh that is executable", func() {
					fi, err := os.Lstat(filepath.Join(root, task.Run.Path))
					Expect(err).NotTo(HaveOccurred())

					Expect(fi.Mode().IsRegular()).To(BeTrue())
					Expect(fi.Mode().IsDir()).To(BeFalse())
					if fi.Mode() < os.FileMode(0700) {
						Fail(fmt.Sprintf("Expected '%s' to be executable (>= 0700), but was 0%o", task.Run.Path, fi.Mode()))
					}
				})

				It("points to a task.sh that has the correct bash flags set", func() {
					f, err := os.Open(filepath.Join(root, task.Run.Path))
					Expect(err).NotTo(HaveOccurred())
					defer f.Close()

					var has_correct_set_eu bool
					var has_flags_in_shebang_line bool
					scanner := bufio.NewScanner(f)
					for scanner.Scan() && !(has_correct_set_eu || has_flags_in_shebang_line) {
						if set_eu_regexp.Match(scanner.Bytes()) {
							has_correct_set_eu = true
						}

						if shebang_regexp.Match(scanner.Bytes()) {
							has_flags_in_shebang_line = true
						}
					}

					if !(has_correct_set_eu || has_flags_in_shebang_line) {
						msg := "incorrect bash args specified;"

						if has_flags_in_shebang_line {
							msg = fmt.Sprintf("%s\n%s", msg, "remove flags from shebang line")
						}

						if !has_correct_set_eu {
							msg = fmt.Sprintf("%s\n%s", msg, "add 'set -eu' two lines after shebang line")
						}

						Fail(msg)
					}
				})
			})
		}
	}
})
