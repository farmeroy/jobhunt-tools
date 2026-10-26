# Job Hunt tools

This is a set of job hunt tools to be run as a cli.

The main tool is `job-monitor`. This stores/builds a sqlite database of employeers and downloads new job postings.

It then filters the job postings with some regex, followed by haiku for titles that need a second look. 

Once the job titles are matched to target roles, the job descriptions are sent to claude sonnet for scoring. 


## How to run

Clone the repo. Ensure Go is installed.
run `go build -o job-monitor.bin ./job-monitor/`

```sh
Usage:
  job-monitor [command]

Available Commands:
  add         Add a company and auto-detect its ATS
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  import      Import companies from a CSV file
  list        List monitored companies
  postings    List stored job postings
  remove      Deactivate a company (soft delete)
  run         Fetch job postings and classify titles for relevance

Flags:
      --db string   path to SQLite database (default "job-monitor.db")
  -h, --help        help for job-monitor
```


## Roadmap

- run job fetch on cron
- set up notifications when new matches are found
- a simple web frontend to filer through jobs
- evals
