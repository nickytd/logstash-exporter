# Logstash exporter
Prometheus exporter for the metrics available in Logstash since version 5.0. This repo is a fork
from https://github.com/BonnierNews/logstash_exporter
It has an updated prometheus client and a build based on go modules

### Flags

| Flag                   | Description                          | Default               |
|------------------------|--------------------------------------|-----------------------|
| -exporter.bind_address | Exporter bind address                | :9198                 |
| -logstash.endpoint     | Metrics endpoint address of logstash | http://localhost:9600 |

## Implemented metrics

* Node Info metrics ...
* Node Stats metrics ...
