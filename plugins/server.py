from flask import Flask, request


app = Flask(__name__)


class StartCommand:

    def __init__(self):
        self._


@app.route('/hello', methods=['POST'])
def hello():
    print("GOT REQUEST")
    with open('request.log', 'wb') as outfile:
        outfile.write(request.data + b'\n')

    return '{"message": "REQUEST OK"}', 200



if __name__ == '__main__':
    app.run('0.0.0.0', 3000, debug=True)
