# PostmanRound2
SocialMediaApp


Postman Collection Link : https://www.postman.com/technical-astronaut-87727624/workspace/socialmediaapp/collection/33010603-fe19e108-a96d-4f18-af3b-059453276de5?action=share&creator=33010603


I have managed to create the backend for the social media app.
I have incorporated almost all the features,including bookmarking feature.
I could not manage to incorporate google oauth2 and frontend for this task.

I have used a jwt based authentication system,where i return a cookie containing the token and user details which are then further used when using the functionalities.

DB : MongoDB
Framework : Gin


All the endpoints are properly documented in the postman collection.

Some points to keep in mind

Almost all the functions return a json response.
The 'Update Post' function requires to send the edited part along with the unedited part as the json body.
For 'Update User' function,only send what is required to be updated.

Also,the secured routes behind the middleware make sure that the user is logged in.So once you logout,the cookie is invalidated as the token will be revoked and hence you cannot access the secured routes anymore.
