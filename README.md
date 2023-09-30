# Compute Scheduler

Like running fly.io on your own infra!

This is specifically designed for running short-lived compute jobs such as a serverless database query.

Future support could include stateful long-running jobs (e.g. supporting persistent disks like fly volumes) and exposing to the network (e.g. getting a HTTP subdomain and listening for requests to respond to).

The design has been somewhat generalized with the above in mind, but primarily for one-off short-lived jobs.