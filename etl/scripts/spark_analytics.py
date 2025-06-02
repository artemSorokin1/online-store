import os
import datetime
import matplotlib.pyplot as plt
import seaborn as sns
from pyspark.sql import SparkSession
from pyspark.sql.functions import col, count, avg, min, max, sum, expr
from pyspark.sql.types import IntegerType

# Инициализация Spark с настройками для работы с JDBC
spark = SparkSession.builder \
    .appName("ECommerceAnalytics") \
    .config("spark.sql.shuffle.partitions", "4") \
    .config("spark.sql.jsonGenerator.ignoreNullFields", "false") \
    .getOrCreate()

# Параметры подключения к БД
DB_URL = "jdbc:postgresql://your-db-host:5432/ecommerce"
DB_PROPERTIES = {
    "user": "your_username",
    "password": "your_password",
    "driver": "org.postgresql.Driver"
}

def load_data_from_db():
    """Загрузка данных из снежинчной схемы БД"""
    try:
        # Таблица фактов - платежи
        payments = spark.read.jdbc(
            url=DB_URL,
            table="payments",
            properties=DB_PROPERTIES
        ).filter(col("status") == "completed")
        
        # Таблица измерений - товары
        products = spark.read.jdbc(
            url=DB_URL,
            table="products",
            properties=DB_PROPERTIES
        )
        
        # Таблица измерений - комментарии
        comments = spark.read.jdbc(
            url=DB_URL,
            table="comments",
            properties=DB_PROPERTIES
        )
        
        # Таблица измерений - пользователи
        users = spark.read.jdbc(
            url=DB_URL,
            table="users",
            properties=DB_PROPERTIES
        )
        
        return payments, products, comments, users
    except Exception as e:
        print(f"Ошибка загрузки данных из БД: {str(e)}")
        raise

def generate_sales_analysis(payments, products, today):
    """Анализ продаж с учетом снежинчной схемы"""
    try:
        # Соединяем платежи с товарами
        sales_with_products = payments.join(
            products,
            payments.product_id == products.id,
            "inner"
        )
        
        # Анализ по категориям
        category_stats = sales_with_products.groupBy("category") \
            .agg(
                count("*").alias("sales_count"),
                sum("amount").alias("total_revenue"),
                avg("amount").alias("avg_price"),
                countDistinct("user_id").alias("unique_customers")
            ) \
            .orderBy(col("total_revenue").desc())
        
        # Визуализация
        plt.figure(figsize=(14, 10))
        
        # Топ категорий по выручке
        plt.subplot(2, 2, 1)
        cat_pd = category_stats.limit(10).toPandas()
        if not cat_pd.empty:
            sns.barplot(x="total_revenue", y="category", data=cat_pd, palette="viridis")
            plt.title("Топ категорий по выручке")
            plt.xlabel("Выручка, руб")
            plt.ylabel("Категория")
        
        # Распределение среднего чека
        plt.subplot(2, 2, 2)
        avg_check = sales_with_products.groupBy("user_id") \
            .agg(sum("amount").alias("total_spent")) \
            .select("total_spent").toPandas()
        if not avg_check.empty:
            sns.histplot(avg_check["total_spent"], bins=30, kde=True)
            plt.title("Распределение суммы покупок на клиента")
            plt.xlabel("Сумма, руб")
            plt.ylabel("Количество клиентов")
        
        # Динамика продаж по времени
        plt.subplot(2, 2, 3)
        time_stats = payments.withColumn("date", expr("to_date(created_at)")) \
            .groupBy("date") \
            .agg(sum("amount").alias("daily_revenue")) \
            .orderBy("date")
        
        if not time_stats.isEmpty():
            time_pd = time_stats.toPandas()
            sns.lineplot(x="date", y="daily_revenue", data=time_pd)
            plt.title("Динамика ежедневной выручки")
            plt.xlabel("Дата")
            plt.ylabel("Выручка, руб")
        
        # Соотношение методов оплаты
        plt.subplot(2, 2, 4)
        payment_methods = payments.groupBy("payment_method") \
            .agg(
                count("*").alias("transactions"),
                sum("amount").alias("total_amount")
            ).toPandas()
        
        if not payment_methods.empty:
            payment_methods.plot.pie(
                y="total_amount",
                labels=payment_methods["payment_method"],
                autopct='%1.1f%%',
                legend=False,
                ax=plt.gca()
            )
            plt.title("Распределение по методам оплаты")
            plt.ylabel("")
        
        plt.tight_layout()
        plt.savefig(f"reports/sales_analysis_{today}.png", dpi=120)
        plt.close()
        
        return category_stats
    except Exception as e:
        print(f"Ошибка при анализе продаж: {str(e)}")
        raise

def generate_customer_analysis(payments, users, today):
    """Анализ клиентской базы"""
    try:
        # Соединяем платежи с пользователями
        customer_activity = payments.join(
            users,
            payments.user_id == users.id,
            "inner"
        )
        
        # RFM-анализ
        rfm = customer_activity.groupBy("user_id") \
            .agg(
                count("*").alias("frequency"),
                sum("amount").alias("monetary"),
                expr("datediff(current_date(), max(to_date(created_at)))").alias("recency")
            )
        
        # Визуализация RFM
        plt.figure(figsize=(14, 10))
        
        # Распределение по частоте покупок
        plt.subplot(2, 2, 1)
        freq_pd = rfm.select("frequency").toPandas()
        if not freq_pd.empty:
            sns.histplot(freq_pd["frequency"], bins=30, kde=True)
            plt.title("Распределение по частоте покупок")
            plt.xlabel("Количество покупок")
            plt.ylabel("Количество клиентов")
        
        # Распределение по сумме покупок
        plt.subplot(2, 2, 2)
        monetary_pd = rfm.select("monetary").toPandas()
        if not monetary_pd.empty:
            sns.histplot(monetary_pd["monetary"], bins=30, kde=True)
            plt.title("Распределение по сумме покупок")
            plt.xlabel("Сумма, руб")
            plt.ylabel("Количество клиентов")
        
        # Распределение по давности покупки
        plt.subplot(2, 2, 3)
        recency_pd = rfm.select("recency").toPandas()
        if not recency_pd.empty:
            sns.histplot(recency_pd["recency"], bins=30, kde=True)
            plt.title("Распределение по давности покупки")
            plt.xlabel("Дней с последней покупки")
            plt.ylabel("Количество клиентов")
        
        # Матрица RFM
        plt.subplot(2, 2, 4)
        rfm_pd = rfm.toPandas()
        if not rfm_pd.empty:
            sns.scatterplot(
                x="recency",
                y="frequency",
                size="monetary",
                hue="monetary",
                data=rfm_pd,
                palette="viridis",
                sizes=(20, 200)
            plt.title("RFM-анализ клиентов")
            plt.xlabel("Recency (дней)")
            plt.ylabel("Frequency")
        
        plt.tight_layout()
        plt.savefig(f"reports/customer_analysis_{today}.png", dpi=120)
        plt.close()
        
        return rfm
    except Exception as e:
        print(f"Ошибка при анализе клиентов: {str(e)}")
        raise

def main():
    try:
        print("=== Начало обработки ===")
        payments, products, comments, users = load_data_from_db()
        
        today = datetime.datetime.now().strftime("%Y-%m-%d")
        os.makedirs("reports", exist_ok=True)
        
        # Основная статистика
        print("\nОсновная статистика:")
        print(f"Платежей: {payments.count()}")
        print(f"Товаров: {products.count()}")
        print(f"Комментариев: {comments.count()}")
        print(f"Пользователей: {users.count()}")
        
        # Генерация отчетов
        print("\nГенерация отчетов...")
        generate_sales_analysis(payments, products, today)
        generate_customer_analysis(payments, users, today)
        
        print("\n=== Обработка завершена успешно ===")
        print("Отчеты сохранены в папке reports/")
        
    except Exception as e:
        print(f"\n!!! Ошибка: {str(e)}")
    finally:
        spark.stop()
        print("Spark сессия закрыта")

def run_analysis():
    spark = SparkSession.builder \
        .appName("ECommerceAnalytics") \
        .config("spark.sql.shuffle.partitions", "4") \
        .getOrCreate()
    
    try:
        payments, products, comments, users = load_data_from_db()
        today = datetime.datetime.now().strftime("%Y-%m-%d")
        
        generate_sales_analysis(payments, products, today)
        generate_customer_analysis(payments, users, today)
        
    finally:
        spark.stop()

if __name__ == "__main__":
    run_analysis()
