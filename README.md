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


cleaning up database:
docker-compose down -v --remove-orphans
docker system prune -a --volumes





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





troubleshooting topics:


docker exec -it mongo mongo
use workflow_resources
db.topics.find({ "assigned": false })

docker exec -it symmetrical-waddle-kafka-1 bash
kafka-topics.sh --list --bootstrap-server localhost:9092



# Superset userguide

### Start the containers
docker-compose up -d

### Exec into the Superset container
docker exec -it superset superset fab create-admin \
    --username admin \
    --firstname Superset \
    --lastname Admin \
    --email admin@example.com \
    --password admin

    
docker exec -it superset bash
superset fab create-admin --username admin --firstname Superset --lastname Admin --email admin@example.com --password admin


### Initialize the database
docker exec -it superset superset db upgrade

### Load example dashboards (optional)
docker exec -it superset superset load_examples

### Start the Superset service
docker exec -it superset superset init




# trino guide:
trino> SHOW CATALOGS;
    ->
 Catalog 
---------
 mongodb
 system
(2 rows)

Query 20241201_051017_00003_pen96, FINISHED, 1 node
Splits: 19 total, 19 done (100.00%)
0.63 [0 rows, 0B] [0 rows/s, 0B/s]

trino> SHOW SCHEMAS IN mongodb;
    ->
       Schema       
--------------------
 information_schema
(1 row)

 SHOW TABLES IN mongodb.workflow_data;
    ->
 Table 
-------
 tasks
(1 row)

Query 20241201_063645_00012_m8934, FINISHED, 1 node
Splits: 19 total, 19 done (100.00%)
0.18 [1 rows, 28B] [5 rows/s, 158B/s]

trino> SELECT * FROM mongodb.workflow_data.tasks LIMIT 10;
    ->
 workflow_id | task_id |                                                                            data                                                            >
-------------+---------+-------------------------------------------------------------------------------------------------------------------------------------------->
           5 |       1 | {field1=Clinique Take The Day Off Cleansing Balm 125ml, field2=Save 30% + Free Gift Worth £33*, field4=£23.80}                             >
           5 |       1 | {field1=Bobbi Brown Vitamin Enriched Face Base 100ml, field2=Save 36%, field4=£53.76}                                                      >
           5 |       1 | {field1=Sol de Janeiro Bom Dia Bright Cream 240ml, field2=Save 30%, field4=£33.60}                                                         >
           5 |       1 | {field1=Perricone MD High Potency Hyaluronic Intensive Serum 59ml, field2=Save 30%, field4=£58.10}                                         >
           5 |       1 | {field1=ESPA Optimal Skin Pro-Serum 30ml, field2=Save 30%, field4=£38.50}                                                                  >
           5 |       1 | {field1=Aveeno Face Calm and Restore Oat Gel Moisturiser 50ml, field2=Save 30%, field4=£8.39}                                              >
           5 |       1 | {field1=Elizabeth Arden Eight Hour Cream Skin Protectant 50ml, field2=Save 30%, field4=£19.60}                                             >
           5 |       1 | {field1=Vichy Mineral 89 Hyaluronic Acid Booster Serum 75ml, field2=Save 30%, field4=£24.50}                                               >
           5 |       1 | {field1=Ole Henriksen Strength Trainer Peptide Boost Moisturiser 50ml, field2=Save 30%, field4=£29.40}                                     >
           5 |       1 | {field1=Estée Lauder Advanced Night Repair Synchronized Multi-Recovery Complex Serum (Various Sizes), field2=Save 30% + Free Gift Worth £68>
(10 rows)



Atlas atlas-wemwxx-shard-0 [primary] test> show collections

Atlas atlas-wemwxx-shard-0 [primary] test> db.user_data.find().pretty()

Atlas atlas-wemwxx-shard-0 [primary] test> show dbs
my_database            40.00 KiB
video_search_project   72.00 KiB
workflow_data         100.00 KiB
admin                 348.00 KiB
local                 941.55 MiB
Atlas atlas-wemwxx-shard-0 [primary] test> use workflow_data
switched to db workflow_data
Atlas atlas-wemwxx-shard-0 [primary] workflow_data> show collections
tasks


mongosh "mongodb://user505:lpmD1CYZ7amPJqYM@cluster0-shard-00-00.1dja6.mongodb.net:27017,cluster0-shard-00-01.1dja6.mongodb.net:27017,cluster0-shard-00-02.1dja6.mongodb.net:27017/workflow_data?authSource=admin&replicaSet=atlas-wemwxx-shard-0"


curl ifconfig.me




superset container:
pip install sqlalchemy-trino



connectoin string:
trino://root@trino:8082/mongodb/workflow_data


future todos:
https://medium.com/@bnprashanth256/oauth2-with-google-account-gmail-in-go-golang-1372c237d25e
https://gal-demo.withgoogle.com/




stripe listen --forward-to http://localhost:8083/webhook/stripe
stripe listen --forward-to http://localhost:8083/webhook/stripe --api-key sk_test_yourkey

stripe trigger checkout.session.completed

switching between live and fake env
1. switch STRIPE_SECRET_KEY2 and STRIPE_SECRET_KEY values
2. stripewebhook switch to live version
3. frontend switch stripe link to live one
4. restart webhook event listener to: stripe listen --forward-to http://localhost:8083/webhook/stripe --api-key stripe_api_key


https://docs.stripe.com/testing --- 5555555555554444



checking if views created:
1. trino --server http://localhost:8082 --catalog mongodb --schema workflow_data
2. SHOW TABLES;
3. SELECT *
FROM user_123_view
LIMIT 10;

SELECT *
FROM mongodb.workflow_data.tasks
LIMIT 10;



CREATE OR REPLACE VIEW workflow_data.test_view AS
SELECT * FROM mongodb.workflow_data.tasks
LIMIT 10;



db.getCollection("_schema").insert({
    "table": "tasks",
    "fields": [
        { "name": "workflow_id", "type": "bigint", "hidden": false },
        { "name": "task_id", "type": "bigint", "hidden": false },
        { "name": "field1", "type": "varchar", "hidden": false },
        { "name": "field2", "type": "varchar", "hidden": false },
        { "name": "field4", "type": "varchar", "hidden": false }
    ]
});



show schemas in mongodb;
       Schema       
--------------------
 information_schema
(1 row)

Query 20241202_141724_00001_29n43, FINISHED, 1 node
Splits: 19 total, 19 done (100.00%)
0.46 [1 rows, 23B] [2 rows/s, 50B/s]



curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
        "username": "admin",
        "password": "admin",
        "provider": "db",
        "refresh": true
      }' \
  http://localhost:8088/api/v1/security/login


curl -X GET \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmcmVzaCI6dHJ1ZSwiaWF0IjoxNzMzMTU0MjQ5LCJqdGkiOiJiYTgwY2IyMi1iNzM1LTRhNjUtOWYzZS05OWIzMWU1NzJmMzYiLCJ0eXBlIjoiYWNjZXNzIiwic3ViIjoxLCJuYmYiOjE3MzMxNTQyNDksImNzcmYiOiJjZjY4ZDA0ZS04MmZlLTQ3OGQtOWM3ZS01ZGIyYTM0ZmZlZTkiLCJleHAiOjE3MzMxNTUxNDl9.qeddDlOitJ6EJIMJv6Gv2Zs5yBVYcmmIWpeVR09bVY0" \
  -H "Referer: http://localhost:8088" \
  http://localhost:8088/api/v1/security/csrf_token/




go run superset_configure.go 
2024/12/03 08:36:41 Access Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmcmVzaCI6dHJ1ZSwiaWF0IjoxNzMzMTg2MjAxLCJqdGkiOiJkOWYyZTBhOS1hODYyLTQ0YjYtYmRhYi00NTBhOTY2YmM3ZDciLCJ0eXBlIjoiYWNjZXNzIiwic3ViIjoxLCJuYmYiOjE3MzMxODYyMDEsImNzcmYiOiJlZTYwNDkxNC1jNTNlLTQxMWUtYmM0MC0zMDcyYjBlNTRmNjAiLCJleHAiOjE3MzMxODcxMDF9.haUvCiE756J9fcxu7dBqr5mD9TD2uIdlDL4apSKCJs8
2024/12/03 08:36:41 CSRF Token: ImU3YzkzNzQ2ZjM3MTFhZmE5OWQ0NmRjOTNmZGNkOWI5Yzg1NDQzODUi.Z05SmQ.WSQ1_t0qv7iXWjGRkJlOitToUFY
2024/12/03 08:36:44 Checkpoint 1: %!s(int=201)
2024/12/03 08:36:44 Checkpoint 2
2024/12/03 08:36:44 Dataset successfully configured.