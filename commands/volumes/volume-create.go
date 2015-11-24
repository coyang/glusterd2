package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

func validateVolumeCreateRequest(msg *volume.VolCreateRequest, r *http.Request, w http.ResponseWriter) error {
	e := utils.GetJSONFromRequest(r, msg)
	if e != nil {
		log.WithField("error", e).Error("Failed to parse the JSON Request")
		rest.SendHTTPError(w, 422, errors.ErrJSONParsingFailed.Error())
		return errors.ErrJSONParsingFailed
	}

	if msg.Name == "" {
		log.Error("Volume name is empty")
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrEmptyVolName.Error())
		return errors.ErrEmptyVolName
	}
	if len(msg.Bricks) <= 0 {
		log.WithField("volume", msg.Name).Error("Brick list is empty")
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrEmptyBrickList.Error())
		return errors.ErrEmptyBrickList
	}
	return nil

}

func createVolume(msg *volume.VolCreateRequest) *volume.Volinfo {
	vol := volume.NewVolumeEntry(msg)
	return vol
}

func volumeCreateHandler(w http.ResponseWriter, r *http.Request) {

	log.Debug("In volume create")
	msg := new(volume.VolCreateRequest)

	e := validateVolumeCreateRequest(msg, r, w)
	if e != nil {
		// Response has been already sent, just return
		return
	}
	if volume.Exists(msg.Name) {
		log.WithField("volume", msg.Name).Error("Volume already exists")
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolExists.Error())
		return
	}
	vol := createVolume(msg)
	if vol == nil {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolCreateFail.Error())
		return
	}

	// Creating client  and server volfile
	e = volgen.GenerateVolfile(vol)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to generate volfile")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	e = volume.AddOrUpdateVolume(vol)
	if e != nil {
		log.WithFields(log.Fields{"error": e.Error(),
			"volume": vol.Name,
		}).Error("Failed to create volume")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	log.WithField("volume", vol.Name).Debug("NewVolume added to store")
	rest.SendHTTPResponse(w, http.StatusCreated, vol)
}