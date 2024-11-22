from pyspark.sql import SparkSession
from pyspark.sql.functions import col, window, count, avg
import argparse
import json

# Parse arguments for dynamic configuration
parser = argparse.ArgumentParser()
parser.add_argument("--input_topic", required=True, help="Kafka input topic")
parser.add_argument("--output_topic", required=True, help="Kafka output topic")
parser.add_argument("--aggregation_fields", required=True, help="Fields to aggregate on (JSON string)")
parser.add_argument("--time_window", default="10 minutes", help="Window size for aggregation")
parser.add_argument("--slide_interval", default="5 minutes", help="Slide interval for aggregation")
parser.add_argument("--kafka_brokers", default="localhost:9092", help="Kafka broker list")
args = parser.parse_args()

# Parse aggregation fields from JSON string
aggregation_fields = json.loads(args.aggregation_fields)

# Initialize SparkSession
spark = SparkSession.builder \
    .appName("DynamicWindowedAggregation") \
    .getOrCreate()

# Read data from Kafka
kafka_df = spark.readStream \
    .format("kafka") \
    .option("kafka.bootstrap.servers", args.kafka_brokers) \
    .option("subscribe", args.input_topic) \
    .load()

# Deserialize Kafka messages
json_df = kafka_df.selectExpr("CAST(value AS STRING) as json") \
    .selectExpr("from_json(json, 'task_id STRING, payload MAP<STRING, STRING>, timestamp TIMESTAMP') as data") \
    .select("data.*")

# Apply dynamic transformations
aggregations = []
for field, agg_func in aggregation_fields.items():
    if agg_func == "count":
        aggregations.append(count("*").alias(f"{field}_count"))
    elif agg_func == "avg":
        aggregations.append(avg(field).alias(f"{field}_avg"))

# Perform windowed aggregation
aggregated_df = json_df \
    .withWatermark("timestamp", "5 minutes") \
    .groupBy(window(col("timestamp"), args.time_window, args.slide_interval)) \
    .agg(*aggregations)

# Write aggregated results to Kafka
query = aggregated_df \
    .selectExpr("CAST(window AS STRING) as key", "to_json(struct(*)) as value") \
    .writeStream \
    .format("kafka") \
    .option("kafka.bootstrap.servers", args.kafka_brokers) \
    .option("topic", args.output_topic) \
    .outputMode("update") \
    .start()

query.awaitTermination()
