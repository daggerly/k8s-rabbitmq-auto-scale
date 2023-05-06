# coding:utf-8
import sys
from datetime import datetime
from time import sleep
import pika
from pika import BasicProperties

host = 'rabbitmq-service'
# host = '172.17.0.6'
port = 5672
username = 'guest'
password = 'guest'
routing_key = sys.argv[1]
exchange = 'testrabbit'

credentials = pika.PlainCredentials(username, password)
cp = pika.ConnectionParameters(host, port=port, credentials=credentials)
connection = pika.BlockingConnection(cp)
channel = connection.channel()
channel.exchange_declare(exchange=exchange, exchange_type='topic',)
properties = BasicProperties(
    content_type='application/json',
    content_encoding='utf-8',
    priority=0,
    delivery_mode=2,
)

try:
    while 1:
        channel.basic_publish(exchange=exchange,
                              routing_key=routing_key,
                              body=str(datetime.now()),
                              properties=properties)
        sleep(1)
finally:
    channel.close()
