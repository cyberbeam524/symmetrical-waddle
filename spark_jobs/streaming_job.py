from pyspark.sql import SparkSession
from pyspark.sql.functions import from_json, col
from pyspark.sql.types import StringType, StructType

# Spark session
spark = SparkSession.builder \
    .appName("Kafka-Spark-Workflow") \
    .getOrCreate()

# Schema for task messages
schema = StructType() \
    .add("task_type", StringType()) \
    .add("payload", StringType())

# Read data from Kafka
df = spark.readStream \
    .format("kafka") \
    .option("kafka.bootstrap.servers", "localhost:9092") \
    .option("subscribe", "workflow_<user_id>") \
    .load()

# Parse Kafka messages
parsed_df = df.selectExpr("CAST(value AS STRING)") \
    .select(from_json(col("value"), schema).alias("data")) \
    .select("data.*")

# Perform transformations
transformed_df = parsed_df.filter(col("task_type") == "transform") \
    .withColumn("processed_payload", col("payload"))

# Write back to Kafka or MongoDB
transformed_df.writeStream \
    .format("console") \
    .outputMode("append") \
    .start() \
    .awaitTermination()
