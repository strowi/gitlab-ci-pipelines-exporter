# gitlab-ci-exporter

`gitlab-ci-exporter` is an exporter for prometheus that shows some
useful metrics from gitlab:

* gitlab_ci_pipeline_last_run_duration_seconds
* gitlab_ci_pipeline_time_since_last_run_seconds
* gitlab_ci_pipeline_status

## TODOs

* registry metrics (count of repos/image-tags)
* .. suggestions..?

```
# HELP gitlab_ci_pipeline_last_run_duration_seconds Duration of last pipeline run
# TYPE gitlab_ci_pipeline_last_run_duration_seconds gauge
gitlab_ci_pipeline_last_run_duration_seconds{id="10350",projectWithNamespace="group-subgroup-project1",ref="master"} 0

# HELP gitlab_ci_pipeline_status GitLab CI pipeline current status
# TYPE gitlab_ci_pipeline_status gauge
gitlab_ci_pipeline_status{id="10350",projectWithNamespace="group-subgroup-project1",ref="master",status="failed"} 0
gitlab_ci_pipeline_status{id="10350",projectWithNamespace="group-subgroup-project1",ref="master",status="running"} 0
gitlab_ci_pipeline_status{id="10350",projectWithNamespace="group-subgroup-project1",ref="master",status="success"} 0

# HELP gitlab_ci_pipeline_time_since_last_run_seconds Elapsed time since most recent GitLab CI pipeline run.
# TYPE gitlab_ci_pipeline_time_since_last_run_seconds gauge
gitlab_ci_pipeline_time_since_last_run_seconds{id="10350",projectWithNamespace="group-subgroup-project1",ref="master"} 1.9792004e+07
```

## Options

| Name | Env | Parameter | Default | Description |
|--|--|--|--|--|
| Gitlab Url | GITLAB_URL | gitlab_url, gu | [https://gitlab.com](https://gitlab.com) | If you
want use your own Gitlab instance |
| Gitlab Token | GITLAB_TOKEN | gitlab_token, gt | - | Create token in
your profile with API and read options |
| Gitlab Refresh | GITLAB_REFRESH | gitlab_refresh, gr | 30 | In seconds,
refresh every x seconds projects list and pipelines |
| Gitlab Owned | GITLAB_OWNED | gitlab_owned, go | false | If you want just
yours projects |
| Port | PORT | port, p | 9999 | Exporter listening port |
