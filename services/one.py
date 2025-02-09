
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.dummy import DummyOperator
from datetime import datetime, timedelta
import requests

def assign_resources(task_id, workflow_id):
    url = "http://localhost:8082/assign-resources"
    payload = {
        "workflow_id": workflow_id,
        "tasks": [{"task_id": task_id}]
    }
    response = requests.post(url, json=payload)
    response.raise_for_status()
    return response.json()

def release_resources(workflow_id):
    url = f"http://localhost:8082/release-resources?workflow_id={workflow_id}"
    response = requests.post(url)
    response.raise_for_status()

def execute_task(task_id, workflow_id):
    # Request resource assignment
    topic_map = assign_resources(task_id, workflow_id)
    print(f"Resources assigned for task {task_id}: {topic_map}")

    try:
        # Simulate task execution
        print(f"Executing task {task_id} for workflow {workflow_id}")
    finally:
        # Release resources after execution
        release_resources(workflow_id)

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

with DAG(
    dag_id="{{ .DagID }}",
    default_args=default_args,
    description="{{ .Description }}",
    schedule_interval="{{ .Schedule }}",
    start_date=datetime.now() - timedelta(days=1),
    catchup=True,
    tags=["dynamic"],
) as dag:
    start = DummyOperator(task_id="start")
    end = DummyOperator(task_id="end")
    workflow_id = "{{ .WorkflowID }}"

    # Dictionary to store created tasks
    task_dict = {}

    {{ range $index, $task := .Tasks }}
    task_dict["{{ $task.TaskID }}"] = PythonOperator(
        task_id="{{ $task.TaskID }}",
        python_callable=execute_task,
        op_args=["{{ $task.TaskID }}", workflow_id],
    )
    {{ end }}

    {{ range $index, $task := .Tasks }}
    {{if !$task.DependsOn}}
    start >> task_dict["{{ $task.TaskID }}"]
    {{ else }}
    {{ range $index, $dependency := $task.DependsOn }}
    task_dict["{{ $dependency }}"] >> task_dict["{{ $task.TaskID }}"]
    {{ end }}
    {{ end }}
    {{ end }}

    terminal_tasks = []
    {{ range $index, $task := .Tasks }}
    {{ if !$task.DependsOn }}
    terminal_tasks.append("{{ $task.TaskID }}")
    {{ end }}
    {{ end }}

    for task_id, task in task_dict.items():
        if task_id not in terminal_tasks:
            task >> end