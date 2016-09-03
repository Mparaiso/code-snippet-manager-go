# build with python 3.4.3
# author mparaiso<mparaiso@online.fr>
# license GPL-3

# SnipThis is a code snippet manager built with Python 3.
# With Snippet Manager, developers no longer have to search 
# for their code snippets in different places, all their snippets 
# are centralized on one server and easy to create,update,manage and query

import flask

app = flask.Flask(__name__)

@app.route('/')
def index():
    """ exemple of use of request object """
    user_agent = flask.request.headers.get('User-Agent')
    return "Hello Code Snippet Manager \n your user agent is %s " % user_agent

@app.route("/badrequest")
def badrequest():
    """ exemple of returning a tuple from an handler """
    return "Bad Request",400,{"X-HTTP-Special-Header":"Some special header"}

@app.route("/makeresponse")
def makeresponse():
    """ creating a response object with make_response """    
    response = flask.make_response("Some response")
    response.set_cookie("answer",'2')
    return response

@app.route('/redirect')
def redirect():
    return flask.redirect("http://yahoo.com")

@app.route("/user/<name>")
def user(name):
    """example of route with variable rendered in a template """
    return flask.render_template("user.j2.html",name=name)

# entry point
if __name__ == '__main__':
    app.run(debug=True)