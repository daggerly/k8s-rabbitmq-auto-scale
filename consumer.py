# coding:utf-8
import sys
from time import sleep
import pika
from pika import BasicProperties

host = 'rabbitmq-service'
# host = '172.17.0.6'
port = 5672
username = 'guest'
password = 'guest'
exchange = 'testrabbit'
routing_key = sys.argv[1]

credentials = pika.PlainCredentials(username, password)
cp = pika.ConnectionParameters(host, port=port, credentials=credentials)
connection = pika.BlockingConnection(cp)
channel = connection.channel()
channel.basic_qos(prefetch_count=1)
properties = BasicProperties(
    content_type='application/json',
    content_encoding='utf-8',
    priority=0,
    delivery_mode=2,
)

def callback(ch, method, properties, body):
    sleep(3)
    ch.basic_ack(delivery_tag=method.delivery_tag)


channel.exchange_declare(exchange=exchange, exchange_type='topic',)
channel.queue_declare(queue=routing_key)
channel.queue_bind(queue=routing_key, routing_key=routing_key, exchange=exchange)
channel.basic_consume(routing_key, callback, auto_ack=False)

channel.start_consuming()

