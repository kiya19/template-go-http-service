package main

import "net/http"

func addRoutes(
    mux *http.ServeMux,
    logger *Logger,
    config *Config,
    store *Store,
) {

    //mux.Handle("Get /home", handleGetHome(logger, store))
}
