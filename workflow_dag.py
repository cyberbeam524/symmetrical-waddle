from airflow import DAG
from airflow.providers.apache.spark.operators.spark_submit import SparkSubmitOperator
from datetime import datetime
from airflow import DAG
from airflow.operators.dummy import DummyOperator
from airflow.operators.python import PythonOperator
from datetime import datetime, timedelta

# Generate DAG dynamically
def create_dag(dag_id, workflow_config):
    default_args = {"start_date": datetime(2023, 1, 1), "retries": 1}
    with DAG(dag_id, default_args=default_args, schedule_interval=None) as dag:
        tasks = {}

        for task_config in workflow_config["tasks"]:
            task_id = task_config["task_id"]
            task_type = task_config["task_type"]

            # Spark Task for windowed aggregation
            if task_type == "windowed_aggregation":
                tasks[task_id] = SparkSubmitOperator(
                    task_id=task_id,
                    application="/path/to/spark_jobs/windowed_aggregation.py",
                    application_args=[
                        f"--input_topic workflow_{workflow_config['user_id']}_{workflow_config['workflow_id']}",
                        f"--output_topic aggregated_results",
                    ],
                    spark_binary="/path/to/spark-submit",
                    conn_id="spark_default"
                )

            # Other task types (e.g., scrape, store, etc.)
            else:
                tasks[task_id] = PythonOperator(
                    task_id=task_id,
                    python_callable=TASK_HANDLERS.get(task_type),
                    provide_context=True,
                )

        # Set dependencies dynamically
        for task_config in workflow_config["tasks"]:
            task_id = task_config["task_id"]
            depends_on = task_config["depends_on"]
            for dependency in depends_on:
                tasks[dependency] >> tasks[task_id]

    return dag
