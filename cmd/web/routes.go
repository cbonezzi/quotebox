package main

import (
    "net/http"

    "github.com/bmizerany/pat"
    "github.com/justinas/alice"
)

// Update the signature for the routes() method so that it returns a
// http.Handler instead of *http.ServeMux.
func (app *application) routes() http.Handler {
    // Create a middleware chain containing our 'standard' middleware
    // which will be used for every request our application receives.
    standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

    // Create a new middleware chain containing the middleware specific to
    // our dynamic application routes. For now, this chain will only contain
    // the session middleware but we'll add more to it later.
    dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)
    
    //mux := http.NewServeMux()

    //this pat.New() substitute above line
    mux := pat.New()
    // Update these routes to use the new dynamic middleware chain followed
    // by the appropriate handler function.
    mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
    // Add the requireAuthentication middleware to the chain.
    mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm))
    // Add the requireAuthentication middleware to the chain.
    mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet))

    mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

    // User routes.
    mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
    mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
    mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
    mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
    // Add the requireAuthentication middleware to the chain.
    mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))
    // Add user profile 
    mux.Get("/user/profile", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.userProfile))
    mux.Get("/user/change-password", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.changePasswordForm))
    mux.Post("/user/change-password",  dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.changePassword))
    mux.Get("/user/passwordreset", dynamicMiddleware.ThenFunc(app.passwordResetForm))
    mux.Post("/user/passwordreset", dynamicMiddleware.ThenFunc(app.passwordReset))
    mux.Get("/user/password/{id}", dynamicMiddleware.ThenFunc(app.password))
    
    fileServer := http.FileServer(http.Dir("./ui/static/"))
    mux.Get("/static/", http.StripPrefix("/static", fileServer))
    
    // Pass the servemux as the 'next' parameter to the secureHeaders middleware.
    // Because secureHeaders is just a function, and the function returns a
    // http.Handler we don't need to do anything else.
    // Wrap the existing chain with the logRequest middleware.
    // Return the 'standard' middleware chain followed by the servemux.
    return standardMiddleware.Then(mux)
}

