[![Go](https://github.com/Nonymus/reboot_watcher/actions/workflows/go.yml/badge.svg)](https://github.com/Nonymus/reboot_watcher/actions/workflows/go.yml)

Node Exporter Reboot Watcher
============================
Little daemon watching for presence or absence of a sentinel file,
updating a Node Exporter Textfile Collector compatible file accordingly.

Useful to export a `reboot_required` metric on debian-ish systems via
[Node Exporter's Textfile Collector](https://github.com/prometheus/node_exporter#textfile-collector),
exposing the presence (or absence) of the `/var/run/reboot-required` file.

# Usage
Configure Node Exporter to monitor a folder for files with additional
metrics to expose (e.g `node_exporter --collector.textfile.directory /var/lib/node_exporter`).
Let `reboot_watcher` run alongside to watch for the sentinel file.

Make sure the user you run the daemon with has write permissions for the `promfile`,
and read permission for the `sentinel` file location.

# Example
The generated file will look something like this

    node_reboot_required 1

# CLI options

| Option      | Description               | Default                              |
|-------------|---------------------------|--------------------------------------|
| `-sentinel` | path of file to watch     | `/var/run/reboot-required`           |
| `-promfile` | path of file to update    | `/var/lib/node_exporter/reboot.prom` |
| `-metric`   | name of the metric to use | `node_reboot_required`               |
| `-help`     | show help                 |                                      |
