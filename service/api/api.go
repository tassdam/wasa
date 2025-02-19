package api

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/tassdam/wasa/service/database"
)

type Config struct {
	Logger   logrus.FieldLogger
	Database database.AppDatabase
}

type Router interface {
	Handler() http.Handler
	Close() error
}

func New(cfg Config) (Router, error) {
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	return &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}, nil
}

type _router struct {
	router     *httprouter.Router
	baseLogger logrus.FieldLogger
	db         database.AppDatabase
}
