package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request: alias is empty"))
			return
		}

		log.Info("deleting url with alias", slog.String("alias", alias))

		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("url not found"))
				return
			}

			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to delete url"))
			return
		}

		log.Info("url deleted successfully", slog.String("alias", alias))

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
