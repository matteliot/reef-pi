package temperature

import (
	"github.com/gorilla/mux"
	"github.com/reef-pi/reef-pi/controller/utils"
	"log"
	"net/http"
)

func (t *Controller) LoadAPI(r *mux.Router) {
	r.HandleFunc("/api/tc/config", t.getConfig).Methods("GET")
	r.HandleFunc("/api/tc/config", t.updateConfig).Methods("POST")
	r.HandleFunc("/api/tc/readings", t.getReadings).Methods("GET")
}

func (t *Controller) getConfig(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) (interface{}, error) {
		return t.config, nil
	}
	utils.JSONGetResponse(fn, w, r)
}

func (t *Controller) getReadings(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) (interface{}, error) {
		readings := []Measurement{}
		t.readings.Do(func(i interface{}) {
			if i == nil {
				return
			}
			v, ok := i.(Measurement)
			if !ok {
				log.Println("ERROR: tmperature subsystem. Failed to convert historical temperature.")
				return
			}
			readings = append(readings, v)
		})
		return readings, nil
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Controller) updateConfig(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var conf Config
	fn := func(_ string) error {
		if err := saveConfig(c.store, conf); err != nil {
			return err
		}
		return c.reload()
	}
	utils.JSONUpdateResponse(&conf, fn, w, r)
}

func (c *Controller) reload() error {
	c.Stop()
	conf, err := loadConfig(c.store)
	if err != nil {
		return err
	}
	c.config = conf
	c.Start()
	return nil
}
