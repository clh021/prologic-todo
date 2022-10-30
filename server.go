package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"git.mills.io/prologic/bitcask"
	"git.mills.io/prologic/todo/static"
	"git.mills.io/prologic/todo/templates"
	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/stats"
	"github.com/unrolled/logger"
)

type counters struct {
	r metrics.Registry
}

func newCounters() *counters {
	counters := &counters{
		r: metrics.NewRegistry(),
	}
	return counters
}

func (c *counters) Inc(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(1)
}

func (c *counters) Dec(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(1)
}

func (c *counters) IncBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(n)
}

func (c *counters) DecBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(n)
}

type server struct {
	bind           string
	templates      *templateManager
	router         *httprouter.Router
	maxItems       int
	maxTitleLength int

	// Logger
	logger *logger.Logger

	// Stats/Metrics
	counters *counters
	stats    *stats.Stats
}

func (s *server) render(name string, w http.ResponseWriter, ctx interface{}) {
	buf, err := s.templates.Exec(name, ctx)
	if err != nil {
		log.WithError(err).Error("error rending template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.WithError(err).Error("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type templateContext struct {
	TodoList []*Todo
}

func (s *server) IndexHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.counters.Inc("n_index")

		var todoList TodoList

		err := db.Fold(func(key []byte) error {
			if string(key) == "nextid" {
				return nil
			}

			var todo Todo

			data, err := db.Get(key)
			if err != nil {
				log.WithError(err).WithField("key", string(key)).Error("error getting todo")
				return err
			}

			err = json.Unmarshal(data, &todo)
			if err != nil {
				return err
			}
			todoList = append(todoList, &todo)
			return nil
		})
		if err != nil {
			log.WithError(err).Error("error listing todos")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sort.Sort(todoList)

		ctx := &templateContext{
			TodoList: todoList,
		}

		s.render("index", w, ctx)
	}
}

func (s *server) AddHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_add")

		var nextID uint64
		rawNextID, err := db.Get([]byte("nextid"))
		if err != nil {
			if err != bitcask.ErrKeyNotFound {
				log.WithError(err).Error("error getting nextid")
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				return
			}
		} else {
			nextID = binary.BigEndian.Uint64(rawNextID)
		}

		if db.Len() > s.maxItems {
			log.Error("error adding item - max number of items reached")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		titleString := r.FormValue("title")
		if len(titleString) > s.maxTitleLength {
			titleString = titleString[:s.maxTitleLength]
		}

		todo := newTodo(titleString)
		todo.ID = nextID

		data, err := json.Marshal(&todo)
		if err != nil {
			log.WithError(err).Error("error serializing todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		key := fmt.Sprintf("todo_%d", nextID)

		err = db.Put([]byte(key), data)
		if err != nil {
			log.WithError(err).Error("error storing todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		buf := make([]byte, 8)
		nextID++
		binary.BigEndian.PutUint64(buf, nextID)
		err = db.Put([]byte("nextid"), buf)
		if err != nil {
			log.WithError(err).Error("error storing nextid")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (s *server) DoneHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_done")

		var id string

		id = p.ByName("id")
		if id == "" {
			id = r.FormValue("id")
		}

		if id == "" {
			log.WithField("id", id).Warn("no id specified to mark as done")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.WithError(err).Error("error parsing id")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		var todo Todo

		key := fmt.Sprintf("todo_%d", i)
		data, err := db.Get([]byte(key))
		if err != nil {
			if err != bitcask.ErrKeyNotFound {
				log.WithError(err).WithField("key", key).Error("todo not found")
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			log.WithError(err).WithField("key", key).Error("error retriving todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(data, &todo)
		if err != nil {
			log.WithError(err).WithField("key", key).Error("error unmarshaling todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		todo.toggleDone()

		data, err = json.Marshal(&todo)
		if err != nil {
			log.WithError(err).WithField("key", key).Error("error marshaling todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		err = db.Put([]byte(key), data)
		if err != nil {
			log.WithError(err).WithField("key", key).Error("error storing todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (s *server) ClearHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_clear")

		var id string

		id = p.ByName("id")
		if id == "" {
			id = r.FormValue("id")
		}

		if id == "" {
			log.WithField("id", id).Warn("no id specified to mark as done")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.WithError(err).Error("error parsing id")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		key := fmt.Sprintf("todo_%d", i)
		err = db.Delete([]byte(key))
		if err != nil {
			log.WithError(err).WithField("key", key).Error("error deleting todo")
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (s *server) statsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bs, err := json.Marshal(s.stats.Data())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
	}
}

func (s *server) listenAndServe() {
	log.Fatal(
		http.ListenAndServe(
			s.bind,
			s.logger.Handler(
				s.stats.Handler(
					gziphandler.GzipHandler(
						s.router,
					),
				),
			),
		),
	)
}

func (s *server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.statsHandler())

	s.router.ServeFiles(
		"/css/*filepath",
		static.GetSubFilesystem("css"),
	)

	s.router.ServeFiles(
		"/icons/*filepath",
		static.GetSubFilesystem("icons"),
	)

	s.router.GET("/", s.IndexHandler())
	s.router.POST("/add", s.AddHandler())

	s.router.GET("/done/:id", s.DoneHandler())
	s.router.POST("/done/:id", s.DoneHandler())

	s.router.GET("/clear/:id", s.ClearHandler())
	s.router.POST("/clear/:id", s.ClearHandler())
}

func newServer(bind string, maxItems int, maxTitleLength int) *server {
	server := &server{
		bind:           bind,
		router:         httprouter.New(),
		templates:      newTemplates("base"),
		maxItems:       maxItems,
		maxTitleLength: maxTitleLength,

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "todo",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
		}),

		// Stats/Metrics
		counters: newCounters(),
		stats:    stats.New(),
	}

	// Templates
	indexTemplate := template.New("index")
	template.Must(indexTemplate.Parse(templates.MustGetTemplate("index.html")))
	template.Must(indexTemplate.Parse(templates.MustGetTemplate("base.html")))

	server.templates.Add("index", indexTemplate)

	server.initRoutes()

	return server
}
