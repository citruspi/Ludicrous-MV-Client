from flask import Flask, request, jsonify, abort
import os
from peewee import *
import json
import random
import string

app = Flask(__name__)

class Files(Model):

    hash = CharField()
    size = IntegerField()
    name = CharField()
    algorithm = CharField()
    link = CharField()

    class Meta:

        database = SqliteDatabase('files.db', threadlocals=True)

Files.create_table(True)

@app.route('/upload', methods=['POST'])
def upload():

    pool = string.ascii_uppercase + string.digits
    short = ''.join((random.choice(pool)) for x in range(8))

    f = json.loads(request.form.get('file'))

    Files.create(
        hash = f['Hash'],
        size = f['Size'],
        name = f['Name'],
        algorithm = f['Algorithm'],
        link = short
    )

    return short, 200


@app.route('/download/<short>', methods=['GET'])
def download(short):

    try:

        record = Files.select().where(Files.link == short).get()

        # [:-1] to remove the NULL terminator (?)

        return "%s;%s;%d" % (record.hash[:-1], record.name, record.size)

    except Files.DoesNotExist:

        abort(404)


if __name__ == "__main__":

    app.run(port=8081)
