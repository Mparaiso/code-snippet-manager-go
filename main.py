# build with python 3.4.3
# author mparaiso<mparaiso@online.fr>
# license GPL-3

# SnipThis is a code snippet manager built with Python 3.
# With Snippet Manager, developers no longer have to search 
# for their code snippets in different places, all their snippets 
# are centralized on one server and easy to create,update,manage and query

import flask
from flask_moment import Moment
from flask.ext.wtf import Form
from wtforms import StringField,SubmitField
from wtforms.validators import Required
from datetime import datetime

app = flask.Flask(__name__)
app.config['SECRET_KEY'] = "some secret key to change later"
moment = Moment(app)


# forms
class NameForm(Form):
    name = StringField("What is your name?",validators=[Required()])
    submit = SubmitField("Submit")

# route map
@app.route('/')
def index():
    """ exemple of use of request object """
    user_agent = flask.request.headers.get('User-Agent')
    return flask.render_template('index.j2.html',
        current_time=datetime.utcnow(),user_agent=user_agent)

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

@app.route("/nameform",methods=['GET','POST'])
def nameform():
    """Rendering an html form """
    form = NameForm(flask.request.form)
    
    if form.validate_on_submit():
        flask.session['name'] = form.name.data
        return flask.redirect(flask.url_for('nameform'))
        
    return flask.render_template("nameform.j2.html",
        form=form,name=flask.session.get('name'))

@app.errorhandler(404)
def page_not_found(e):
    """ example of custom error handling page """
    return flask.render_template("404.j2.html"),404

@app.errorhandler(500)
def internal_server_error(e):
    return flask.render_template("500.j2.html"),500


# entry point
if __name__ == '__main__':
    app.run(debug=True)