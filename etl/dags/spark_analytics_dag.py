from datetime import datetime, timedelta
from airflow import DAG
from airflow.providers.apache.spark.operators.spark_submit import SparkSubmitOperator
from airflow.operators.email import EmailOperator
from airflow.providers.slack.operators.slack_webhook import SlackWebhookOperator
from airflow.operators.python import PythonOperator
from airflow.exceptions import AirflowException


SPARK_SCRIPT_PATH = '/opt/airflow/dags/scripts/spark_analytics.py'
REPORTS_DIR = '/opt/airflow/reports'

default_args = {
    'owner': 'ecommerce_team',
    'depends_on_past': False,
    'start_date': datetime(2023, 6, 1),
    'retries': 2,
    'retry_delay': timedelta(minutes=5),
    'email_on_failure': False,
    'email_on_retry': False,
}

def slack_failure_alert(context):
    """Отправка уведомления в Slack при ошибке"""
    alert = SlackWebhookOperator(
        task_id='slack_failed',
        slack_webhook_conn_id='slack_webhook',
        message=f"""
        :red_circle: DAG Failed: {context['dag'].dag_id}
        *Task*: {context['task'].task_id}
        *Execution Time*: {context['ts']}
        *Log*: {context['task_instance'].log_url}
        """
    )
    return alert.execute(context)

def check_report_files():
    """Проверка генерации отчетов"""
    import os
    required_files = [
        f"{REPORTS_DIR}/sales_analysis_{datetime.today().strftime('%Y-%m-%d')}.png",
        f"{REPORTS_DIR}/customer_analysis_{datetime.today().strftime('%Y-%m-%d')}.png"
    ]
    missing_files = [f for f in required_files if not os.path.exists(f)]
    
    if missing_files:
        raise AirflowException(f"Отсутствуют файлы отчетов: {missing_files}")

with DAG(
    dag_id='ecommerce_daily_analytics',
    default_args=default_args,
    description='Ежедневный анализ данных e-commerce',
    schedule_interval='0 2 * * *',  
    catchup=False,
    tags=['analytics', 'spark'],
    on_failure_callback=slack_failure_alert,
) as dag:

    run_spark_analysis = SparkSubmitOperator(
        task_id='run_spark_analysis',
        application=SPARK_SCRIPT_PATH,
        conn_id='spark_cluster',
        application_args=[
            '--date', '{{ ds }}',  ё
            '--reports-dir', REPORTS_DIR
        ],
        executor_memory='4g',
        driver_memory='2g',
        conf={
            'spark.sql.shuffle.partitions': '50',
            'spark.dynamicAllocation.enabled': 'true'
        },
    )

    verify_reports = PythonOperator(
        task_id='verify_reports',
        python_callable=check_report_files,
    )

    send_email_report = EmailOperator(
        task_id='send_email_report',
        to='analytics-team@company.com',
        subject='E-commerce Daily Report {{ ds }}',
        html_content="""
        <h3>Ежедневный отчет по электронной коммерции</h3>
        <p>Анализ за {{ ds }} успешно завершен.</p>
        <p>Доступные отчеты:</p>
        <ul>
            <li>Анализ продаж</li>
            <li>RFM-анализ клиентов</li>
        </ul>
        """,
        files=[
            f"{REPORTS_DIR}/sales_analysis_{{{{ ds }}}}.png",
            f"{REPORTS_DIR}/customer_analysis_{{{{ ds }}}}.png"
        ],
        trigger_rule='all_success'
    )

    slack_success = SlackWebhookOperator(
        task_id='slack_success',
        slack_webhook_conn_id='slack_webhook',
        message=":white_check_mark: E-commerce анализ успешно завершен за {{ ds }}",
        trigger_rule='all_success'
    )

    run_spark_analysis >> verify_reports >> [send_email_report, slack_success]