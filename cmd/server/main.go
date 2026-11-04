package main

import (
        "log"
        "net/http"

        "agnis-crm/internal/handler"
        "agnis-crm/internal/storage"
        "agnis-crm/internal/usecase"
)

func main() {
        // Слои снизу вверх: storage → usecase → handler.
        st := storage.New()
        uc := usecase.New(st)
        h := handler.New(uc)

        mux := http.NewServeMux()
        h.Register(mux)

        addr := ":8080"
        log.Printf("serving HTTP on %s", addr)
        if err := http.ListenAndServe(addr, mux); err != nil {
                log.Fatal(err)
        }
}
