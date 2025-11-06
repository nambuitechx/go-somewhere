FROM prom/prometheus:latest

COPY ./prometheus.yaml /etc/prometheus/prometheus.yaml

CMD ["--config.file=/etc/prometheus/prometheus.yaml", "--storage.tsdb.path=/prometheus"]
