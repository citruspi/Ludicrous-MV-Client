from flask import Flask, request, jsonify, abort
import os
import json
import random
import string
from pymongo import MongoClient

app = Flask(__name__)

client = MongoClient('localhost', 27017)
db = client.ludicrous_mv
collection = db.files

@app.route('/upload', methods=['POST'])
def upload():

    pool = string.ascii_uppercase + string.digits
    token = ''.join((random.choice(pool)) for x in range(8))

    f = json.loads(request.form.get('file'))
    f['token'] = token

    collection.insert(f)

    return token, 200


@app.route('/download/<token>', methods=['GET'])
def download(token):

    f = collection.find_one({'token':token})

    if f is None:

        abort(404)

    else:

        del f['_id']

        return json.dumps(f)

if __name__ == "__main__":

    app.run(port=8081, debug=True)
