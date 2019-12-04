/*
Main application package
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli"

	"github.com/Labbs/gitlab-ci-pipelines-exporter/settings"
)

var version = "v1.2"

var (
	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		[]string{"projectWithNamespace", "ref", "id"},
	)

	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		[]string{"projectWithNamespace", "ref", "id"},
	)

	status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_status",
			Help: "GitLab CI pipeline current status",
		},
		[]string{"projectWithNamespace", "ref", "status", "id"},
	)

	registryRepositoryTagCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_registry_tags_total",
			Help: "GitLab Registry Tag count",
		},
		[]string{"projectWithNamespace", "registryRepository"},
	)
)

// main init configuration
func main() {
	app := cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"
	app.Flags = settings.NewContext()
	app.Action = runWeb
	app.Version = version

	app.Run(os.Args)
}

// runWeb start http server
func runWeb(ctx *cli.Context) {
	go getGitlabInfo()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "/metrics")
	})
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("starting exporter with port %v", settings.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(settings.Port), nil))
}

// init prometheus metrics
func init() {
	prometheus.MustRegister(timeSinceLastRun)
	prometheus.MustRegister(lastRunDuration)
	prometheus.MustRegister(status)
	prometheus.MustRegister(registryRepositoryTagCount)
}

// getGitlabInfo get all needed informations from gitlab instance and update some metrics
func getGitlabInfo() {
	// init gitlab configuration
	client := gitlab.NewClient(nil, settings.Gitlab.Token)
	client.SetBaseURL(settings.Gitlab.Url)

	trueVal := false
	falseVal := false

	// get all projects
	opt := &gitlab.ListProjectsOptions{
		Archived: &falseVal,
		Simple:   &trueVal,
		Owned:    &settings.Gitlab.Owned,
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	for {
		for {

			projects, resp, err := client.Projects.ListProjects(opt)
			if err != nil {
				log.Fatalln(err)
			}

			// list all projects we've found
			for _, project := range projects {

				// list all pipelines within $project
				pipelines, _, _ := client.Pipelines.ListProjectPipelines(project.ID, &gitlab.ListProjectPipelinesOptions{})
				// TODO: ??
				var lastPipeline *gitlab.Pipeline

				// pipeline metrics
				if len(pipelines) != 0 {

					// get duration of last pipeline
					lastPipeline, _, _ = client.Pipelines.GetPipeline(project.ID, pipelines[0].ID)
					lastRunDuration.WithLabelValues(strings.Replace(project.PathWithNamespace, "/", "-", -1), pipelines[0].Ref, strconv.Itoa(pipelines[0].ID)).Set(float64(lastPipeline.Duration))

					// get status of last pipeline
					for _, s := range []string{"success", "failed", "running"} {
						if s == lastPipeline.Status {
							status.WithLabelValues(strings.Replace(project.PathWithNamespace, "/", "-", -1), pipelines[0].Ref, s, strconv.Itoa(pipelines[0].ID)).Set(1)
						} else {
							status.WithLabelValues(strings.Replace(project.PathWithNamespace, "/", "-", -1), pipelines[0].Ref, s, strconv.Itoa(pipelines[0].ID)).Set(0)
						}
					}

					// get last run of pipeline
					timeSinceLastRun.WithLabelValues(
						strings.Replace(project.PathWithNamespace, "/", "-", -1),
						pipelines[0].Ref,
						strconv.Itoa(pipelines[0].ID)).Set(
						float64(time.Since(*lastPipeline.CreatedAt).Round(time.Second).Seconds()))
				}

				// container-registry metrics
				if project.ContainerRegistryEnabled {

					// get all registryRepositories
					registryRepositories, _, err := client.ContainerRegistry.ListRegistryRepositories(project.ID, &gitlab.ListRegistryRepositoriesOptions{})
					if err != nil {
						log.Fatalln(err)
					}

					// iterate all registryRepositories
					for _, registryRepository := range registryRepositories {

						// get all tags per repository
						registryRepositoryTag, _, err := client.ContainerRegistry.ListRegistryRepositoryTags(project.ID, registryRepository.ID, &gitlab.ListRegistryRepositoryTagsOptions{})
						if err != nil {
							log.Fatalln(err)
						}

						// debug
						// fmt.Println("id: ", registryRepository.ID, "path:", registryRepository.Path, "count: ", float64(len(registryRepositoryTag)))

						// expose metric for tag-count per registryRepository
						registryRepositoryTagCount.WithLabelValues(strings.Replace(project.PathWithNamespace, "/", "-", -1), registryRepository.Path).Set(float64(len(registryRepositoryTag)))
					}
				}

			}

			// Exit the loop when we've seen all pages.
			if opt.Page >= resp.TotalPages {
				break
			}

			// Update the page number to get the next page.
			opt.Page = resp.NextPage
		}
		time.Sleep(time.Duration(settings.Gitlab.Refresh) * time.Second)
	}
}
