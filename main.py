# build with python 3.4.3
# author mparaiso<mparaiso@online.fr>
# license GPL-3

# SnipThis is a code snippet manager built with Python 3.
# With Snippet Manager, developers no longer have to search 
# for their code snippets in different places, all their snippets 
# are centralized on one server and easy to create,update,manage and query

import flask

app = flask.Flask(__name__)

# homepage
@app.route('/')
def index():
    user_agent = flask.request.headers.get('User-Agent')
    return "Hello Code Snippet Manager \n your user agent is %s " % user_agent

# returning a tuple body,status,headers
@app.route("/badrequest")
def basrequest():
    return "Bad Request",400,{"X-HTTP-Special-Header":"Some special header"}



# route with variable
@app.route("/user/<name>")
def user(name):
    return "Hello %s" % name

# entry point
if __name__ == '__main__':
    app.run(debug=True)