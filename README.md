# symmetrical-waddle



docker-compose down --volumes --remove-orphans
docker system prune -f

docker-compose up --build
docker exec -it airflow airflow users  create --role Admin --username admin --email admin --firstname admin --lastname admin --password admin


docker exec -it airflow ls /usr/local/airflow/dags
__pycache__  example_dag.py  workflow_dag.py  workflow_user1_wf1.py  workflow_user2_wf1.py  workflow_user2_wf2.py  workflow_user3_wf3.py

docker inspect airflow
check "Mounts"

docker exec -it airflow airflow dags list

docker exec -it airflow airflow dags unpause <dag_id>


docker exec -it airflow airflow config get-value core dags_folder


chmod -R 755 ./dags


docker restart airflow-scheduler



docker exec -it airflow cat /usr/local/airflow/airflow.cfg
docker exec -it airflow airflow config list



docker exec -it airflow airflow dags trigger example_dag


docker-compose restart airflow-scheduler


docker exec -it airflow airflow db upgrade




Here are the key steps to troubleshoot this error:

Verify DAG file location and accessibility:

bashCopy# Check if DAG file exists in correct directory
ls -l /opt/airflow/dags/example_dag.py

# Check file permissions
chmod 644 /opt/airflow/dags/example_dag.py

Refresh DAG list:

bashCopyairflow dags reserialize
airflow dags list  # Verify your DAG appears

Check DAG configuration:

pythonCopy# In your DAG file, ensure proper configuration:
dag = DAG(
    'example_dag',
    default_args=default_args,
    schedule_interval=timedelta(days=1),
    catchup=False,
    is_paused_upon_creation=False  # Important
)