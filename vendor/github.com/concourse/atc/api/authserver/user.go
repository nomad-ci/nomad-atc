package authserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/concourse/atc"
	"github.com/concourse/atc/api/present"
	"github.com/concourse/atc/auth"
)

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	hLog := s.logger.Session("user")

	var user User
	isSystem, isPresent := r.Context().Value("system").(bool)
	if isPresent && isSystem {
		user = User{System: &isSystem}
	} else {
		authTeam, authTeamFound := auth.GetTeam(r)
		if !authTeamFound {
			hLog.Error("failed-to-get-team-from-auth", errors.New("failed-to-get-team-from-auth"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		team, found, err := s.teamFactory.FindTeam(authTeam.Name())
		if err != nil {
			hLog.Error("failed-to-get-team-from-db", errors.New("failed-to-get-team-from-db"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !found {
			hLog.Error("team-not-found-in-db", errors.New("team-not-found-in-db"))
		} else {
			presentedTeam := present.Team(team)
			user = User{
				Team: &presentedTeam,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		hLog.Error("failed-to-encode-user", errors.New("failed-to-get-team-from-db"))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type User struct {
	Team   *atc.Team `json:"team,omitempty"`
	System *bool     `json:"system,omitempty"`
}
