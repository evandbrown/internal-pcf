package pcf_pipelines_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var placeholderRegexp = regexp.MustCompile("[({]{2}([a-zA-Z0-9-_]+)[)}]{2}")

var _ = Describe("pcf-pipelines", func() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working dir: %s", err)
	}

	root := filepath.Dir(cwd)
	baseDir := filepath.Base(cwd)

	var pipelinePaths []string
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Base(path) == "pipeline.yml" {
			relPipelinePath, err := filepath.Rel(cwd, path)
			if err != nil {
				return err
			}
			pipelinePaths = append(pipelinePaths, relPipelinePath)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk: %s", err)
	}

	for _, path := range pipelinePaths {
		pipelinePath := path

		Context(fmt.Sprintf("pipeline at %s", pipelinePath), func() {
			var configBytes []byte

			BeforeEach(func() {
				var err error
				configBytes, err = ioutil.ReadFile(pipelinePath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("specifies only valid job names in any `passed` definitions in the buildplan", func() {
				var config atc.Config
				cleanConfigBytes := placeholderRegexp.ReplaceAll(configBytes, []byte("true"))
				err := yaml.Unmarshal(cleanConfigBytes, &config)
				Expect(err).NotTo(HaveOccurred())

				for _, job := range config.Jobs {
					for _, plan := range job.Plans() {
						checkValidJobsList(config.Jobs, plan.Passed, job.Name)
					}
				}
			})

			It("specifies all and only the params that the pipeline's tasks expect", func() {
				var config atc.Config
				cleanConfigBytes := placeholderRegexp.ReplaceAll(configBytes, []byte("true"))
				err := yaml.Unmarshal(cleanConfigBytes, &config)
				Expect(err).NotTo(HaveOccurred())

				for _, job := range config.Jobs {
					for _, task := range allTasksInPlan(&job.Plan) {
						failMessage := fmt.Sprintf("Found error in the following pipeline:\n    %s\n\nin the following task's params:\n    %s/%s\n", pipelinePath, job.Name, task.Name())

						var configParams []string
						for k := range task.Params {
							configParams = append(configParams, k)
						}

						if strings.HasPrefix(task.TaskConfigPath, baseDir) {
							taskPath := strings.TrimPrefix(task.TaskConfigPath, baseDir+"/")
							relpath, err := filepath.Rel(cwd, filepath.Join(cwd, taskPath))
							Expect(err).NotTo(HaveOccurred())

							bs, err := ioutil.ReadFile(relpath)
							Expect(err).NotTo(HaveOccurred())

							taskConfig := atc.TaskConfig{}
							err = yaml.Unmarshal(bs, &taskConfig)
							Expect(err).NotTo(HaveOccurred())

							var taskParams []string
							for k := range taskConfig.Params {
								taskParams = append(taskParams, k)
							}

							assertUnorderedEqual(taskParams, configParams, failMessage)
						}
					}
				}
			})

			It("has a params file with all and only the params that the pipeline specifies", func() {
				paramsPath := filepath.Join(filepath.Dir(pipelinePath), "params.yml")
				_, err := os.Lstat(paramsPath)
				Expect(err).NotTo(HaveOccurred())

				bs, err := ioutil.ReadFile(paramsPath)
				Expect(err).NotTo(HaveOccurred())

				paramsMap := map[string]interface{}{}
				err = yaml.Unmarshal(bs, paramsMap)
				Expect(err).NotTo(HaveOccurred())

				var params []string
				for k := range paramsMap {
					params = append(params, k)
				}

				matches := placeholderRegexp.FindAllStringSubmatch(string(configBytes), -1)

				uniqueMatches := map[string]struct{}{}
				for _, match := range matches {
					uniqueMatches[match[1]] = struct{}{}
				}

				var placeholders []string
				for match := range uniqueMatches {
					placeholders = append(placeholders, match)
				}

				failMessage := fmt.Sprintf(`
Found error with the following pipeline:
%s

in the following params template:
%s
`, pipelinePath, paramsPath)

				assertUnorderedEqual(placeholders, params, failMessage)
			})

			It("provides all of the resources that the tasks it defines require", func() {
				var config atc.Config
				cleanConfigBytes := placeholderRegexp.ReplaceAll(configBytes, []byte("true"))
				err := yaml.Unmarshal(cleanConfigBytes, &config)
				Expect(err).NotTo(HaveOccurred())

				for _, job := range config.Jobs {
					tasks := allTasksInPlan(&job.Plan)
					resources := availableResources(&job.Plan)

					for i, task := range tasks {
						if !strings.HasPrefix(task.TaskConfigPath, "pcf-pipelines") {
							continue
						}

						var inputs []atc.TaskInputConfig
						if task.TaskConfig != nil {
							inputs = task.TaskConfig.Inputs
						} else {
							inputs = taskInputConfigs(filepath.Join(root, task.TaskConfigPath))
						}

						for k := range task.InputMapping {
							var validInputMapping bool
							for _, input := range inputs {
								if input.Name == k {
									validInputMapping = true
									break
								}
							}
							if !validInputMapping {
								Fail(fmt.Sprintf("could not find input mapping for '%s' in '%#v'\n", k, inputs))
							}
						}

						for j := 0; j < i; j++ {
							upstreamTask := tasks[j]

							if upstreamTask.TaskConfig != nil {
								for _, output := range upstreamTask.TaskConfig.Outputs {
									resources = append(resources, output.Name)
								}
							}

							if upstreamTask.TaskConfigPath != "" {
								var upstreamTaskConfig atc.TaskConfig
								bs, err := ioutil.ReadFile(filepath.Join(root, upstreamTask.TaskConfigPath))
								Expect(err).NotTo(HaveOccurred())

								err = yaml.Unmarshal(bs, &upstreamTaskConfig)
								Expect(err).NotTo(HaveOccurred())

								for _, output := range upstreamTaskConfig.Outputs {
									resources = append(resources, output.Name)
								}
							}

							for _, v := range upstreamTask.OutputMapping {
								resources = append(resources, v)
							}
						}

					OUTER:
						for _, input := range inputs {
							for _, actual := range resources {
								if input.Name == actual {
									continue OUTER
								}

								for k, v := range task.InputMapping {
									if k == input.Name && v == actual {
										continue OUTER
									}
								}
							}

							Fail(fmt.Sprintf("did not find matching get, put, output or input_mapping of '%s', which is required by task '%s'", input.Name, task.Name()))
						}
					}
				}
			})
		})
	}
})

func checkValidJobsList(jobs atc.JobConfigs, jobNames []string, location string) {
	for _, jobName := range jobNames {
		_, exists := jobs.Lookup(jobName)
		Expect(exists).Should(BeTrue(), fmt.Sprintf("%s is not a valid job defined in %s", jobName, location))
	}
}

func allTasksInPlan(seq *atc.PlanSequence) []atc.PlanConfig {
	var tasks []atc.PlanConfig

	for _, planConfig := range *seq {
		if planConfig.Aggregate != nil {
			tasks = append(tasks, allTasksInPlan(planConfig.Aggregate)...)
		}
		if planConfig.Do != nil {
			tasks = append(tasks, allTasksInPlan(planConfig.Do)...)
		}
		if planConfig.Task != "" {
			tasks = append(tasks, planConfig)
		}
	}

	return tasks
}

func availableResources(seq *atc.PlanSequence) []string {
	var resources []string

	for _, planConfig := range *seq {
		if planConfig.Aggregate != nil {
			resources = append(resources, availableResources(planConfig.Aggregate)...)
		}

		if planConfig.Do != nil {
			resources = append(resources, availableResources(planConfig.Do)...)
		}

		if planConfig.Get != "" {
			resources = append(resources, planConfig.Get)
		}

		if planConfig.Put != "" {
			resources = append(resources, planConfig.Put)
		}
	}

	return resources
}

func assertUnorderedEqual(left, right []string, failMessage string) {
	for _, l := range left {
		Expect(right).To(ContainElement(l), failMessage)
	}

	for _, r := range right {
		var found bool

		for _, l := range left {
			if r == l {
				found = true
				break
			}
		}

		if !found {
			Expect(right).NotTo(ContainElement(r), failMessage)
		}
	}
}

var taskConfigs = map[string]*atc.TaskConfig{}

func taskInputConfigs(path string) []atc.TaskInputConfig {
	taskConfig, ok := taskConfigs[path]

	if !ok {
		bs, err := ioutil.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())

		err = yaml.Unmarshal(bs, &taskConfig)
		Expect(err).NotTo(HaveOccurred())

		taskConfigs[path] = taskConfig
	}

	return taskConfig.Inputs
}
