FROM apache/airflow:slim-2.10.3-python3.8

# Switch to root user to install dependencies
USER airflow

# Install required Python dependencies
RUN pip install apache-airflow-providers-apache-spark psycopg2-binary

# Initialize Airflow database
RUN airflow db init

# Switch back to airflow user
USER airflow