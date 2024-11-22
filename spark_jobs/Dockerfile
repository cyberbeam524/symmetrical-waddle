FROM bitnami/spark:latest

WORKDIR /app
COPY spark_jobs /app/spark_jobs
CMD ["spark-submit", "--packages", "org.apache.spark:spark-sql-kafka-0-10_2.12:3.1.1", "spark_jobs/windowed_aggregation.py"]
